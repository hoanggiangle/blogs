#!/usr/bin/env bash
set -e
docker-compose down

if [ "$1" = "nobuild" ]; then
 docker-compose up -d gw server envoy
else
 echo "Building source..."
 CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
 docker-compose up --build -d gw server envoy
fi;
