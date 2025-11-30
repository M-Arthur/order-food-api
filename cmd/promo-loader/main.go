package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	var (
		filesStr    string
		tmpDir      string
		output      string
		sortBin     string
		parallelism int
		timeoutStr  string
	)

	flag.StringVar(&filesStr, "files", "", "Comma-separated list of gzip files")
	flag.StringVar(&tmpDir, "tmp-dir", "./tmp/promo", "Temporary directory for intermediate files")
	flag.StringVar(&output, "output", "./valid_promo_codes.txt", "Output file for valid promo codes")
	flag.StringVar(&sortBin, "sort-bin", "sort", "Path to sort binary")
	flag.IntVar(&parallelism, "parallelism", 3, "Number of files to process in parallel")
	flag.StringVar(&timeoutStr, "timeout", "0s", "Overall timeout (e.g. 30m, 1h); 0s = no timeout")
	flag.Parse()

	if filesStr == "" {
		fmt.Fprintln(os.Stderr, "missing -files (comma-separated list of gzip files)")
		os.Exit(1)
	}

	files := splitAndTrim(filesStr)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no files provided after parsing -files")
		os.Exit(1)
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid timeout: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := ExtractValidPromoCodes(ctx, files, tmpDir, output, sortBin, parallelism); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
