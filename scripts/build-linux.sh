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
# Never replace a portable folder that may already contain user data.
if [[ $variant == linux-x64-portable && -e "$root/dist/$variant" ]]; then
  echo "Каталог $root/dist/$variant уже существует. Переместите его перед сборкой, чтобы сохранить данные." >&2
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
trap 'rm -rf -- "$staging"' EXIT
chmod 755 -- "$staging"
mkdir -p -- "$staging/bin"
cp -a -- "$bundle/." "$staging/"
# Keep Flutter's build-tree name for flutter run; expose the product name in dist.
mv -- "$staging/wireglow_ui" "$staging/$app_name"
install -m 755 -- "$cache_root/bin/controller" "$staging/bin/controller"
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
rm -rf -- "$root/dist/$variant"
mv -- "$staging" "$root/dist/$variant"
echo "Дистрибутив: $root/dist/$variant"
echo "Запуск: $root/dist/$variant/$app_name"
