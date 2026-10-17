#!/usr/bin/env bash
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$root/scripts/build-env.sh"
cache="$cache_root/engines/linux-amd64"
mkdir -p -- "$cache/bin"
staging=$(mktemp -d "$cache/.download.XXXXXX")
trap 'rm -rf -- "$staging"' EXIT

# Утилита и controller читают один и тот же встроенный YAML-манифест.
(
  cd -- "$root"
  go build -o "$staging/engine-manifest" ./cmd/engine-manifest
)
manifest() { "$staging/engine-manifest" get "$1"; }
fetch() {
  local repo=$1 version=$2 asset=$3 checksum=$4
  local cached="$cache/$repo/$version/$asset"
  mkdir -p -- "$(dirname -- "$cached")"
  if [[ -f "$cached" ]] && [[ $(sha256sum -- "$cached" | cut -d ' ' -f 1) == "$checksum" ]]; then
    echo "Кэш: $repo $version / $asset"
    return
  fi
  echo "Скачивание: $repo $version / $asset"
  curl --fail --location --show-error --retry 3 --connect-timeout 30 \
    --output "$staging/$asset" \
    "https://github.com/$repo/releases/download/$version/$asset"
  if [[ $(sha256sum -- "$staging/$asset" | cut -d ' ' -f 1) != "$checksum" ]]; then
    echo "Неверная SHA-256: $repo $version / $asset" >&2
    exit 1
  fi
  mv -- "$staging/$asset" "$cached"
}

for engine in free-turn-proxy mihomo; do
  repo=$(manifest "$engine.repository")
  version=$(manifest "$engine.version")
  asset=$(manifest "$engine.platforms.linux-amd64.asset")
  checksum=$(manifest "$engine.platforms.linux-amd64.sha256")
  compression=$(manifest "$engine.platforms.linux-amd64.compression")
  fetch "$repo" "$version" "$asset" "$checksum"
  case "$compression" in
    none) install -m 755 -- "$cache/$repo/$version/$asset" "$staging/$engine" ;;
    gzip)
      gzip -dc -- "$cache/$repo/$version/$asset" > "$staging/$engine"
      chmod 755 -- "$staging/$engine"
      ;;
    *) echo "Неподдерживаемое сжатие: $compression" >&2; exit 1 ;;
  esac
done
mv -- "$staging/free-turn-proxy" "$cache/bin/free-turn-proxy"
mv -- "$staging/mihomo" "$cache/bin/mihomo"
