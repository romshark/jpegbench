#!/bin/bash

# Ensure cleanup on exit.
cleanup() {
    rm -f "$OPT_BINARY" "$STD_BINARY"
}
trap cleanup EXIT

OPT_BINARY=$(mktemp)
STD_BINARY=$(mktemp)

OPT_DIR="./cmd/opt"
STD_DIR="./cmd/std"

echo "Compiling $OPT_DIR..."
GOOS=linux GOARCH=amd64 go build -o "$OPT_BINARY" "$OPT_DIR" || \
    { echo "Compilation failed for $OPT_DIR"; exit 1; }

echo "Compiling $STD_DIR..."
GOOS=linux GOARCH=amd64 go build -o "$STD_BINARY" "$STD_DIR" || \
    { echo "Compilation failed for $STD_DIR"; exit 1; }

# Get binary sizes.
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS `stat` format.
    OPT_SIZE=$(stat -f%z "$OPT_BINARY")
    STD_SIZE=$(stat -f%z "$STD_BINARY")
else
    # Linux `stat` format.
    OPT_SIZE=$(stat -c%s "$OPT_BINARY")
    STD_SIZE=$(stat -c%s "$STD_BINARY")
fi

# Convert sizes to human-readable format.
OPT_SIZE_HR=$(du -h "$OPT_BINARY" | awk '{print $1}')
STD_SIZE_HR=$(du -h "$STD_BINARY" | awk '{print $1}')

# Calculate difference.
DIFF_SIZE=$((OPT_SIZE - STD_SIZE))
DIFF_SIZE=${DIFF_SIZE#-}

# Convert to human-readable format.
if command -v numfmt &>/dev/null; then
    DIFF_SIZE_HR=$(numfmt --to=iec-i --suffix=B "$DIFF_SIZE")
else
    # Fallback: simple conversion if `numfmt` isn't available.
    DIFF_SIZE_HR="${DIFF_SIZE}B"
fi

# Display results.
echo "Binary sizes:"
echo "  optimized: $OPT_SIZE ($OPT_SIZE_HR)"
echo "  standard:  $STD_SIZE ($STD_SIZE_HR)"
echo "optimized is bigger by: $DIFF_SIZE ($DIFF_SIZE_HR)"
