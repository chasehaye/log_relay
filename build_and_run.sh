#!/bin/sh

set -e

echo "Go environment status:"
go env

echo "Building the application..."
go build -o ./build/main ./cmd/server/main.go

echo "Starting the application..."
exec ./build/main