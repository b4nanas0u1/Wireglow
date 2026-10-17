#!/usr/bin/env bash
# Shared cache locations for repository build and development scripts.
cache_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)/.cache"
export GOCACHE="$cache_root/go-build"
export GOPATH="$cache_root/go"
export GOMODCACHE="$GOPATH/pkg/mod"
export PUB_CACHE="$cache_root/pub"
export ANALYZER_STATE_LOCATION_OVERRIDE="$cache_root/dart-analyzer"
export PATH="$cache_root/flutter/bin:$PATH"
