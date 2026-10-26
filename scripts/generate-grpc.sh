#!/usr/bin/env bash
set -euo pipefail

# Последние стабильные версии, проверенные 2026-10-13; обновлять вместе с go.mod.
PROTOC_VERSION=36.2
PROTOBUF_GO_VERSION=v1.36.12
GRPC_GO_PLUGIN_VERSION=v1.6.2
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$root/scripts/build-env.sh"
cache="$cache_root/grpc-tools"
mkdir -p "$cache"
cache=$(cd -- "$cache" && pwd)

exe=
case "$(uname -s)-$(uname -m)" in
  Linux-x86_64) platform=linux-x86_64 ;;
  Linux-aarch64|Linux-arm64) platform=linux-aarch_64 ;;
  Darwin-x86_64) platform=osx-x86_64 ;;
  Darwin-arm64) platform=osx-aarch_64 ;;
  MINGW*|MSYS*) platform=win64; exe=.exe ;;
  *) echo 'Unsupported platform for protoc; use Linux, macOS or Git Bash on Windows.' >&2; exit 1 ;;
esac
protoc_dir="$cache/protoc-$PROTOC_VERSION-$platform"
if [[ ! -x "$protoc_dir/bin/protoc" && ! -f "$protoc_dir/bin/protoc.exe" ]]; then
  archive=$(mktemp "$cache/protoc.XXXXXX.zip")
  trap 'rm -f "$archive"' EXIT
  curl --fail --location --silent --show-error \
    "https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-$platform.zip" \
    -o "$archive"
  mkdir -p "$protoc_dir"
  unzip -q -o "$archive" -d "$protoc_dir"
  rm -f "$archive"
  trap - EXIT
fi
# Инструменты находятся в cache проекта, не заменяют пользовательские binaries.
if [[ ! -x "$cache/bin/protoc-gen-go$exe" ]] || [[ "$("$cache/bin/protoc-gen-go$exe" --version)" != "protoc-gen-go $PROTOBUF_GO_VERSION" ]]; then
  GOBIN="$cache/bin" go install "google.golang.org/protobuf/cmd/protoc-gen-go@$PROTOBUF_GO_VERSION"
fi
if [[ ! -x "$cache/bin/protoc-gen-go-grpc$exe" ]] || [[ "$("$cache/bin/protoc-gen-go-grpc$exe" --version)" != "protoc-gen-go-grpc ${GRPC_GO_PLUGIN_VERSION#v}" ]]; then
  GOBIN="$cache/bin" go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@$GRPC_GO_PLUGIN_VERSION"
fi
cd "$root"
PATH="$cache/bin:$PATH" "$protoc_dir/bin/protoc$exe" \
  --proto_path=proto --proto_path="$protoc_dir/include" \
  --go_out=. --go_opt=module=wireglow \
  --go-grpc_out=. --go-grpc_opt=module=wireglow \
  proto/controller.proto
