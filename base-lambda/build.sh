#!/bin/bash

# Clean any existing build artifacts
rm -f bootstrap

# Build the Go binary for Linux
GOOS=linux GOARCH=amd64 go build -o build/bootstrap

# Make the binary executable
chmod +x build/bootstrap