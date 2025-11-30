package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const (
	minLength        = 8
	maxLength        = 10
	buf1MB           = 1 << 20
	logProgressEvery = 1_000_000 // log every 1M lines
)

// ExtractValidPromoCodes:
//   - files: list of gzip files with raw codes (one per line)
//   - tmpDir: directory to store intermediate files
//   - output: final file path containing valid promo codes
//   - sortBin: path to "sort" binary (use "sort" on most systems)
//   - parallelism: number of files to filter in parallel (0 or <0 → 1)
func ExtractValidPromoCodes(
	ctx context.Context,
	files []string,
	tmpDir string,
	output string,
	sortBin string,
	parallelism int,
) error {
	if len(files) == 0 {
		return fmt.Errorf("no input files provided")
	}
	if tmpDir == "" {
		return fmt.Errorf("tmpDir is empty")
	}
	if sortBin == "" {
		sortBin = "sort"
	}
	if parallelism <= 0 {
		parallelism = 1
	}

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}

	// 1) Filter each gzip file into tmp raw files (just length checks).
	rawPaths := make([]string, len(files))
	for i := range files {
		rawPaths[i] = filepath.Join(tmpDir, fmt.Sprintf("file_%d.raw", i+1))
	}

	if err := filterAllFiles(ctx, files, rawPaths, parallelism); err != nil {
		return fmt.Errorf("filter stage: %w", err)
	}

	// 2) Sort each raw file using external sort.
	sortedPaths := make([]string, len(rawPaths))
	for i, raw := range rawPaths {
		sorted := filepath.Join(tmpDir, fmt.Sprintf("file_%d.sorted", i+1))
		if err := externalSort(raw, sorted, sortBin); err != nil {
			return fmt.Errorf("sort stage: %w", err)
		}
		sortedPaths[i] = sorted
	}

	// 3) Merge sorted files, keeping codes that appear in >=2 files.
	if err := mergeSortedFiles(sortedPaths, output); err != nil {
		return fmt.Errorf("merge stage: %w", err)
	}

	return nil
}

// filterAllFiles runs filterFile in parallel for each gzip file.
func filterAllFiles(ctx context.Context, inputs, outputs []string, parallelism int) error {
	type job struct {
		in  string
		out string
	}

	jobs := make(chan job)
	errCh := make(chan error, parallelism)

	var wg sync.WaitGroup
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				// allow cancellation
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				default:
				}

				if err := filterFile(j.in, j.out); err != nil {
					errCh <- fmt.Errorf("filter %s: %w", j.in, err)
					return
				}
			}
		}()
	}

	for i := range inputs {
		jobs <- job{in: inputs[i], out: outputs[i]}
	}
	close(jobs)

	wg.Wait()
	close(errCh)

	// return first error if any
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// filterFile reads a .gz file, filters lines by length, writes valid lines to output.
func filterFile(inputPath, outputPath string) error {
	fmt.Printf("[FILTER] Start processing %s\n", inputPath)

	in, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer func() {
		_ = in.Close()
	}()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer func() {
		_ = gz.Close()
	}()

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	w := bufio.NewWriterSize(out, buf1MB)
	defer func() {
		_ = w.Flush()
	}()

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 0, buf1MB), buf1MB)

	var total, kept int64

	for scanner.Scan() {
		line := scanner.Text()
		total++

		if total%logProgressEvery == 0 {
			fmt.Printf("[FILTER] %s processed %d lines\n", filepath.Base(inputPath), total)
		}

		if isValidLength(line) {
			_, _ = w.WriteString(line + "\n")
			kept++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	fmt.Printf("[FILTER] Done %s → total=%d kept=%d\n",
		filepath.Base(inputPath), total, kept)

	return nil
}

func isValidLength(s string) bool {
	n := len(s)
	return n >= minLength && n <= maxLength
}

// externalSort uses system "sort" to sort input into output.
func externalSort(inputPath, outputPath, sortBin string) error {
	fmt.Printf("[SORT] Sorting %s → %s\n", filepath.Base(inputPath), filepath.Base(outputPath))

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create sort output: %w", err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	cmd := exec.Command(sortBin, inputPath)
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sort failed: %w", err)
	}

	fmt.Printf("[SORT] Done sorting %s\n", filepath.Base(inputPath))
	return nil
}

// mergeSortedFiles:
//   - inputs: sorted files, one code per line
//   - output: file with codes that appear in at least 2 different input files
func mergeSortedFiles(inputs []string, output string) error {
	fmt.Printf("[MERGE] Starting merge of %d files\n", len(inputs))

	files := make([]*os.File, len(inputs))
	scanners := make([]*bufio.Scanner, len(inputs))

	for i, path := range inputs {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}
		files[i] = f

		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, buf1MB), buf1MB)
		scanners[i] = sc
	}
	defer func() {
		for _, f := range files {
			if f != nil {
				_ = f.Close()
			}
		}
	}()

	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	w := bufio.NewWriterSize(out, buf1MB)
	defer func() {
		_ = w.Flush()
	}()

	current := make([]string, len(scanners))
	eof := make([]bool, len(scanners))

	// prime each scanner
	for i, sc := range scanners {
		if sc.Scan() {
			current[i] = sc.Text()
		} else {
			eof[i] = true
		}
	}

	var processed, validCount int64

	for {
		// Check if all files are done
		allEOF := true
		for _, e := range eof {
			if !e {
				allEOF = false
				break
			}
		}
		if allEOF {
			break
		}

		// Find smallest current value
		var min string
		first := true
		for i := range current {
			if eof[i] {
				continue
			}
			if first || current[i] < min {
				min = current[i]
				first = false
			}
		}

		// Count how many files have this code
		count := 0
		for i := range current {
			if eof[i] {
				continue
			}
			if current[i] == min {
				count++

				if scanners[i].Scan() {
					current[i] = scanners[i].Text()
				} else {
					eof[i] = true
				}
			}
		}

		processed++
		if processed%logProgressEvery == 0 {
			fmt.Printf("[MERGE] processed %d merge steps...\n", processed)
		}

		if count >= 2 {
			validCount++
			_, _ = w.WriteString(min + "\n")
		}
	}

	fmt.Printf("[MERGE] Completed merge → valid codes: %d\n", validCount)
	return nil
}
