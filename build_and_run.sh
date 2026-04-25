#!/bin/sh
set -e

echo "Building..."
go build -o ./build/main ./cmd/server/main.go

echo "Stopping old process..."
pkill main || true

echo "Starting new process..."
nohup ./build/main > app.log 2>&1 &