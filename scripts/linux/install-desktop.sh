#!/usr/bin/env bash
set -euo pipefail

# GNOME/Wayland resolves the window's application ID through this desktop entry.
# Optional menu registration; the application itself launches directly.
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
icon=${2:-"$root/data/app-icon.png"}
IFS= read -r app_name < "$(dirname -- "$icon")/app-name.txt"
if [[ ! "$app_name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
  echo 'Некорректное имя приложения.' >&2
  exit 1
fi
launcher=${1:-"$root/$app_name"}
if [[ "$launcher" != /* || ! -x "$launcher" || "$icon" != /* || ! -f "$icon" ]]; then
  echo 'Для desktop entry нужны абсолютные пути к launcher и иконке.' >&2
  exit 1
fi

escape_value() {
  local value=$1
  value=${value//\\/\\\\}
  value=${value//$'\n'/\\n}
  value=${value//$'\r'/\\r}
  value=${value//$'\t'/\\t}
  printf '%s' "$value"
}

escape_exec() {
  local value=$1
  value=${value//\\/\\\\}
  value=${value//\"/\\\"}
  value=${value//\$/\\\$}
  value=${value//\`/\\\`}
  value=${value//%/%%}
  escape_value "\"$value\""
}

data_home=${XDG_DATA_HOME:-"$HOME/.local/share"}
if [[ "$data_home" != /* ]]; then
  data_home="$HOME/.local/share"
fi
applications="$data_home/applications"
mkdir -p -- "$applications"
desktop="$applications/dev.wireglow.wireglow_ui.desktop"
temporary=$(mktemp "$applications/.wireglow-desktop.XXXXXX")
trap 'rm -f -- "$temporary"' EXIT
cat > "$temporary" <<EOF
[Desktop Entry]
Type=Application
Name=$(escape_value "$app_name")
Comment=Wireglow — Free Turn Proxy Client
Exec=$(escape_exec "$launcher")
Icon=$(escape_value "$icon")
Terminal=false
Categories=Network;
StartupWMClass=dev.wireglow.wireglow_ui
EOF
if ! cmp -s -- "$temporary" "$desktop"; then
  chmod 644 -- "$temporary"
  mv -f -- "$temporary" "$desktop"
fi
