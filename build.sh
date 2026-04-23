#!/bin/sh
set -e
cd "$(dirname "$0")"

mkdir -p dist

# Cloud builds set GOOS/GOARCH automatically from meta.json build.arch.
GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"

CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o dist/metricscapturemodule ./cmd/metricscapturemodule
cp run.sh dist/run.sh
chmod +x dist/metricscapturemodule dist/run.sh

tar -czf dist/archive.tar.gz -C dist metricscapturemodule run.sh
