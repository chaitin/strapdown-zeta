#!/bin/sh
set -ex

grunt
make -C server _static/version
[ "$TRAVIS_OS_NAME" = "linux" ] && docker run --rm -i -v `pwd`:/app -w /app golang:alpine sh -c 'apk add --no-cache cmake gcc git make musl-dev zlib-dev && BUILD_STATIC=1 make -C server deps all' || make -C server deps all
