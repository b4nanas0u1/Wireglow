#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$root/scripts/build-env.sh"
protoc="$cache_root/grpc-tools/protoc-36.2-linux-x86_64/bin/protoc"
cd "$root/ui"
# Optional pub arguments (for example --offline) are forwarded for cached builds.
flutter pub get "$@"
mkdir -p lib/controller/gen
"$protoc" --proto_path="$root/proto" \
 --proto_path="$cache_root/grpc-tools/protoc-36.2-linux-x86_64/include" \
 --plugin="protoc-gen-dart=$root/ui/tool/protoc-gen-dart" \
 --dart_out=grpc:lib/controller/gen controller.proto
