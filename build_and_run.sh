#!/bin/sh
set -e

cd /srv/webserver-back/fudesoftware/

echo "Deploying PRODUCTION..."

docker compose up -d --build

echo "Done."