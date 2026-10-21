#!/usr/bin/env bash
set -euo pipefail

scripts_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
root=$(cd -- "$scripts_root/.." && pwd)
if [[ $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  echo 'Сборка поддерживает Linux x86_64; запускайте её на целевой платформе.' >&2
  exit 1
fi
variant=linux-x64
case "${1:-}" in
  '') ;;
  --portable) variant=linux-x64-portable ;;
  *) echo "Использование: $0 [--portable]" >&2; exit 1 ;;
esac
if (( $# > 1 )); then
  echo "Использование: $0 [--portable]" >&2
  exit 1
fi
source "$root/scripts/build-env.sh"
# Абсолютные пути нужны после перехода в каталог Flutter.
mkdir -p -- "$cache_root"
cache_root=$(cd -- "$cache_root" && pwd)
"$scripts_root/download-linux-engines.sh"
relay="$cache_root/engines/linux-amd64/bin/free-turn-proxy"
mihomo="$cache_root/engines/linux-amd64/bin/mihomo"

"$root/ui/build-linux.sh"
bundle="$root/ui/build/linux/x64/release/bundle"
app_name=$(cat "$bundle/data/app-name.txt")
if [[ ! "$app_name" =~ ^[a-zA-Z0-9_-]+$ || ! -x "$bundle/wireglow_ui" ]]; then
  echo "Не найден собранный Flutter UI: $bundle/wireglow_ui" >&2
  exit 1
fi

# Старый дистрибутив заменяется только после успешной сборки и упаковки.
mkdir -p -- "$root/dist"
staging=$(mktemp -d "$root/dist/.$variant.XXXXXX")
target="$root/dist/$variant"
previous=
cleanup() {
  rm -rf -- "$staging"
  if [[ -n "$previous" && -d "$previous" && ! -e "$target" ]]; then
    mv -- "$previous" "$target"
  fi
}
trap cleanup EXIT
chmod 755 -- "$staging"
mkdir -p -- "$staging/bin"
cp -a -- "$bundle/." "$staging/"
# Keep Flutter's build-tree name for flutter run; expose the product name in dist.
mv -- "$staging/wireglow_ui" "$staging/$app_name"
install -m 755 -- "$cache_root/bin/controller" "$staging/bin/controller"
go -C "$root" build -o "$staging/bin/tun-helper" ./cmd/tun-helper
if [[ $variant != linux-x64-portable ]]; then
  install -m 755 -- "$scripts_root/linux/install-tun-helper.sh" "$staging/"
fi
install -m 755 -- "$relay" "$staging/bin/free-turn-proxy"
install -m 755 -- "$mihomo" "$staging/bin/mihomo"
install -m 755 -- "$scripts_root/linux/install-desktop.sh" "$staging/"
install -m 644 -- "$scripts_root/linux/README.md" "$staging/README.md"
if [[ $variant == linux-x64-portable ]]; then
  printf '{}\n' > "$staging/portable.json"
  cat "$scripts_root/linux/README-portable.md" >> "$staging/README.md"
fi
(
  cd -- "$staging"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
)
if [[ -e "$target" || -L "$target" ]]; then
  if [[ ! -d "$target" || -L "$target" ]]; then
    echo "Ожидался обычный каталог дистрибутива: $target" >&2
    exit 1
  fi
  if [[ $variant == linux-x64-portable && ( -e "$target/data" || -L "$target/data" ) ]]; then
    if [[ ! -d "$target/data" || -L "$target/data" ]]; then
      echo "Ожидался обычный каталог portable-данных: $target/data" >&2
      exit 1
    fi
    # Hold existing owner locks while copying data and replacing the bundle.
    # No new lock files are created in an unused portable distribution.
    portable_locks=()
    for lock in launcher controller; do
      lock_path="$target/data/runtime/$lock.lock"
      if [[ -f "$lock_path" ]]; then
        exec {lock_fd}<>"$lock_path"
        if ! flock -n "$lock_fd"; then
          echo 'Завершите portable-приложение перед заменой сборки.' >&2
          exit 1
        fi
        portable_locks+=("$lock_fd")
      fi
    done
    # data/ mixes user storage with Flutter resources. Replace shipped assets;
    # carry over all other entries, including hidden files and recovery state.
    # Keep user files out of the distribution's SHA256SUMS above.
    (
      shopt -s dotglob nullglob
      for entry in "$target/data/"*; do
        case "${entry##*/}" in
          flutter_assets|icudtl.dat|app-name.txt|app-icon.png) continue ;;
        esac
        cp -a -- "$entry" "$staging/data/"
      done
    )
    chmod --reference="$target/data" "$staging/data"
  fi
  previous="$staging.previous"
  mv -- "$target" "$previous"
fi
mv -- "$staging" "$target"
if [[ -n "$previous" ]]; then
  rm -rf -- "$previous"
fi
echo "Дистрибутив: $root/dist/$variant"
echo "Запуск: $root/dist/$variant/$app_name"
