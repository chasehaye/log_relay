#!/bin/sh

set -e

echo "Building containers..."
docker compose up -d --build