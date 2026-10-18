#!/usr/bin/env bash
set -euo pipefail

scripts_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
if (( $# > 1 )); then
  echo "Usage: $0 [Debian-version]" >&2
  exit 1
fi
for tool in dpkg-deb dpkg-shlibdeps readelf strip; do
  command -v "$tool" >/dev/null || { echo "Missing build tool: $tool" >&2; exit 1; }
done
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-trimpath"
"$scripts_root/build-linux.sh"
exec "$scripts_root/package-debian.sh" "$@"
