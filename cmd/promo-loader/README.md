# promo-loader

`promo-loader` is a lightweight command-line tool for processing very large gzip-compressed promo code files. It extracts all promo codes that:

1. Have a length between 8 and 10 characters
2. Appear in at least two different input files

The tool is optimized for large files (hundreds of MB to several GB) and does not load entire files into memory.

---

## How it works

1. Stream each `.gz` file and keep only promo codes with length 8–10.
2. Partition all filtered promo codes into buckets using a hash function.
3. For each bucket, determine which codes appear in at least two input files.
4. Write all valid promo codes into a single output file.

This avoids slow external sorting and keeps memory usage low.

---

## Usage

Example:

```
go run ./cmd/promo-loader
--files=/data/file1.gz,/data/file2.gz,/data/file3.gz
--tmp-dir=./tmp/promo
--output=./valid_promo_codes.txt
--parallelism=3
```
---

## Command Flags

- `--files`  
  Comma-separated list of gzip promo code files.

- `--tmp-dir`  
  Directory used to store intermediate files.

- `--output`  
  Path to the final output file containing valid promo codes.

- `--parallelism`  
  Number of goroutines used for filtering gzip files in parallel.

- `--timeout`  
  Optional timeout for the entire job (e.g. `30m`, `1h`).

---

## Output Format

The output file will contain one promo code per line, for example:
```
AB28DF9H
XQ91LK02
ZZ889123
```
Each promo code in the output appears in at least two input files.

---

## Performance Notes

- Filtering gzip files is parallelized.
- Hash partitioning avoids slow external sorting.
- Memory usage stays low because each bucket fits into memory.
- Good default parallelism values:
  - 1–2 for laptops
  - 2–4 for servers or large multi-core machines

---

## Troubleshooting

- Empty output file  
  Check that promo codes actually appear in at least two files.

- Missing temporary directory  
  Ensure `--tmp-dir` exists or can be created.

---

## Notes

- This implementation assumes promo codes are ASCII without spaces or tabs.
