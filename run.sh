#!/usr/bin/env bash
set -ex
cd "$(dirname "$0")"

TAG=homebox:dev
docker build -t=$TAG .
docker run --rm -p 7745:7745 -v ./data:/data $TAG
