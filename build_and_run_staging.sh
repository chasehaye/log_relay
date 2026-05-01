#!/bin/sh
set -e

cd /srv/webserver-back/fudesoftware/

echo "Deploying STAGING..."

docker compose -f compose.staging.yml up -d --build

echo "Done."