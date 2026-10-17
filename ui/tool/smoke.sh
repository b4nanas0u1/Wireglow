#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
source "$root/scripts/build-env.sh"
mkdir -p -- "$cache_root"
dir=$(mktemp -d "$cache_root/wireglow-ui-smoke.XXXXXX")
trap 'rm -rf "$dir"' EXIT
cat > "$dir/engine" <<'ENGINE'
#!/bin/sh
case "$1" in
-h) echo '-provider -peer -listen -mode -transport -dns-mode -obf-profile -routes -debug -obf-timing -bond'; exit 0;;
-t) exit 0;;
esac
exec /bin/sleep 300
ENGINE
chmod +x "$dir/engine"
export WIREGLOW_CONTROLLER="$cache_root/bin/controller"
export WIREGLOW_RELAY="$dir/engine" WIREGLOW_MIHOMO="$dir/engine"
export WIREGLOW_DATA_DIR="$dir/data" WIREGLOW_RUNTIME_DIR="$dir/run"
export PUB_CACHE="$cache_root/pub" CI=true
cd "$root/ui"
dart run tool/smoke.dart
if compgen -G "$dir/run/owner-*" >/dev/null; then
 echo 'Launcher runtime directory survived cleanup' >&2
 exit 1
fi
