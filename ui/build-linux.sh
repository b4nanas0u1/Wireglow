#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$root/scripts/build-env.sh"
export CI=true FLUTTER_SUPPRESS_ANALYTICS=true
mkdir -p "$cache_root/bin"
cd "$root"
go build -o "$cache_root/bin/controller" ./cmd/controller
go build -o "$cache_root/bin/engine-manifest" ./cmd/engine-manifest
manifest() { "$cache_root/bin/engine-manifest" get "$1"; }
mihomo_version=$(manifest mihomo.version)
mihomo_repository=$(manifest mihomo.repository)
free_turn_version=$(manifest free-turn-proxy.version)
free_turn_repository=$(manifest free-turn-proxy.repository)
cd ui
flutter clean
flutter pub get
version=$(sed -n 's/^version: \([^+]*\).*/\1/p' pubspec.yaml)
commit=$(git -C "$root" rev-parse --short HEAD 2>/dev/null || printf 'неизвестен')
tag=$(git -C "$root" describe --tags --exact-match HEAD 2>/dev/null || true)
build_time=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
flutter build linux --release --no-pub \
  --dart-define="APP_VERSION=$version" \
  --dart-define="APP_COMMIT=$commit" \
  --dart-define="APP_TAG=$tag" \
  --dart-define="APP_BUILD_TIME=$build_time" \
  --dart-define="MIHOMO_VERSION=$mihomo_version" \
  --dart-define="MIHOMO_REPOSITORY=$mihomo_repository" \
  --dart-define="FREE_TURN_VERSION=$free_turn_version" \
  --dart-define="FREE_TURN_REPOSITORY=$free_turn_repository"
