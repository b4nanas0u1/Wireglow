#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$root/scripts/build-env.sh"
export WIREGLOW_CONTROLLER=${WIREGLOW_CONTROLLER:-"$cache_root/bin/controller"}
export WIREGLOW_RELAY=${WIREGLOW_RELAY:-"$cache_root/engines/linux-amd64/bin/free-turn-proxy"}
export WIREGLOW_MIHOMO=${WIREGLOW_MIHOMO:-"$cache_root/engines/linux-amd64/bin/mihomo"}
for binary in "$WIREGLOW_CONTROLLER" "$WIREGLOW_RELAY" "$WIREGLOW_MIHOMO"; do
  if [[ ! -x "$binary" ]]; then
    echo "Нет executable: $binary. См. ui/README.md." >&2
    exit 1
  fi
done
bundle="$root/ui/build/linux/x64/release/bundle/wireglow_ui"
if [[ ! -x "$bundle" ]]; then
  echo 'Сначала выполните ./ui/build-linux.sh' >&2
  exit 1
fi
exec "$bundle" "$@"
