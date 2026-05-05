#!/bin/bash

set -e

echo "Running tests with coverage..."

# Clean up old coverage files
rm -f coverage.out lcov.info
rm -rf coverage/
mkdir -p coverage

# Ensure ~/go/bin is in PATH for installed Go tools
export PATH="$HOME/go/bin:$PATH"

# Run tests for all packages and collect coverage
echo "Testing all packages..."
go test ./... -coverprofile=coverage.out -covermode=atomic

# Install gcov2lcov if not already installed
if ! command -v gcov2lcov &> /dev/null; then
    echo "Installing gcov2lcov..."
    go install github.com/jandelgado/gcov2lcov@latest
fi

# Convert Go coverage to lcov format
echo "Converting coverage to lcov format..."
gcov2lcov -infile=coverage.out -outfile=lcov.info

# Display coverage summary
echo ""
echo "Coverage Summary:"
go tool cover -func=coverage.out | tail -10

echo ""
echo "Coverage report generated:"
echo "  - Go format: coverage.out"
echo "  - LCOV format: lcov.info"
echo ""
echo "To view HTML report:"
echo "  go tool cover -html=coverage.out -o coverage.html"
echo ""
echo "To generate lcov HTML report:"
echo "  genhtml lcov.info -o coverage_html"
echo "  open coverage_html/index.html"
