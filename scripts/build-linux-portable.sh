#!/usr/bin/env bash
set -euo pipefail
scripts_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
exec "$scripts_root/build-linux.sh" --portable "$@"
