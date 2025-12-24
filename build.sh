#!/bin/bash

# Script build binary cho Linux
# Sử dụng: ./build.sh

echo "Building QH88 server for Linux..."

# Build binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qh88-server main.go

if [ $? -eq 0 ]; then
    echo "✓ Build thành công! File: qh88-server"
    echo "  Bạn có thể upload file này lên VPS và chạy:"
    echo "  chmod +x qh88-server"
    echo "  ./qh88-server"
else
    echo "✗ Build thất bại!"
    exit 1
fi

