#!/bin/sh

set -e

cd /srv/webserver-back/fudesoftware/

echo "Building the application..."
go build -o ./build/main ./cmd/server/main.go

echo "Stopping old process..."
pkill main || true

echo "Starting new process..."

nohup ./build/main > ./build/server.log 2>&1 &

echo "Server started. Logs at ./build/server.log"