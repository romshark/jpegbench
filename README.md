# Go's standard `image/jpeg` decoding optimization.

This repository demonstrates a potential optimization of decoding JPEG images
in the Go standard library. This is achieved by unrolling
unzig and dst writing loops in `scan.go`.

## Running tests

```sh
go test -v ./...
```

## Running benchmark

```sh
./bench.sh 10 .
```

## Running binary size comparison

```sh
./cmpbinsz.sh
```
