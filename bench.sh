#!/bin/bash

if [ "$#" -lt 1 ]; then
	echo "Usage: $0 <count> [benchmark_filter]"
	exit 1
fi

if ! command -v benchstat &> /dev/null; then
	echo "benchstat not found. Installing..."
	go install golang.org/x/perf/cmd/benchstat@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
fi

COUNT=$1
FILTER=${2:-"."} # Default to all benchmarks if not provided.

STD_TXT="std.txt"
OPT_TXT="opt.txt"

echo "std: count=$COUNT and filter='$FILTER'..."
BENCH_FN="std" go test -run none -bench="$FILTER" -count="$COUNT" -benchmem > "$STD_TXT"

echo "opt: count=$COUNT and filter='$FILTER'..."
BENCH_FN="opt" go test -run none -bench="$FILTER" -count="$COUNT" -benchmem > "$OPT_TXT"

benchstat $STD_TXT $OPT_TXT
