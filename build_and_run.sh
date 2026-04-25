#!/bin/sh

echo "Go environment status:" && \
go env && \
GOMODCACHE=`go env GOMODCACHE` && \
echo "Go mod cache directory:" && \
echo $GOMODCACHE && \
ls -lsa $GOMODCACHE && \
echo "Building the application..." && \
go build -o ./build/main ./cmd/server/main.go && \
echo "Starting the application..." && \
./build/main
