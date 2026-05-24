#!/bin/bash

# Build script for git-ego - creates binaries for all major platforms
set -e

VERSION=${1:-$(git describe --tags --always)}
OUTPUT_DIR="./dist"

echo "Building git-ego version: $VERSION"

# Clean and create output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build targets (OS/ARCH combinations)
declare -a targets=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64" 
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

for target in "${targets[@]}"; do
    OS=$(echo $target | cut -d'/' -f1)
    ARCH=$(echo $target | cut -d'/' -f2)
    
    output_name="git-ego-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building $output_name..."
    GOOS=$OS GOARCH=$ARCH go build -ldflags "-X github.com/bgreenwell/git-ego/cmd.version=$VERSION" -o "$OUTPUT_DIR/$output_name" .
done

echo "Build complete! Binaries are in $OUTPUT_DIR"
ls -la "$OUTPUT_DIR"