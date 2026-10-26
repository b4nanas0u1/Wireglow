#!/usr/bin/env bash
# Explicit administrator operation; never run by the GUI or regular builds.
set -euo pipefail
if [[ $(uname -s) != Linux || $EUID != 0 || $# != 1 || ! $1 =~ ^[0-9]+$ || $1 == 0 ]]; then
  echo "Usage: sudo $0 <permitted non-root UID>" >&2
  exit 1
fi
client_uid=$1
getent passwd "$client_uid" >/dev/null
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# Run the bundled copy, alongside bin/ and SHA256SUMS.
if [[ ! -f "$root/bin/tun-helper" ]]; then
  echo 'Run this script from the built Linux bundle (bin/tun-helper required).' >&2
  exit 1
fi
# SHA256SUMS is delivered with the bundle. Only install a trusted distribution:
# checksums detect corruption, not an attacker replacing the entire bundle.
(cd "$root" && sha256sum --check --status SHA256SUMS)
command -v systemctl >/dev/null
[[ -x /usr/bin/ip && -c /dev/net/tun ]]
systemctl stop wireglow-tun.service 2>/dev/null || true
install -d -o root -g root -m 755 /usr/local/libexec/wireglow-tun
install -o root -g root -m 755 "$root/bin/tun-helper" /usr/local/libexec/wireglow-tun/tun-helper
install -o root -g root -m 755 "$root/bin/mihomo" /usr/local/libexec/wireglow-tun/mihomo
cat > /etc/systemd/system/wireglow-tun.service <<UNIT
[Unit]
Description=Wireglow restricted TUN helper
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/libexec/wireglow-tun/tun-helper --uid=$client_uid
Restart=on-failure
RestartSec=1
KillMode=control-group
TimeoutStopSec=25
RuntimeDirectory=wireglow-tun
RuntimeDirectoryMode=0755
RuntimeDirectoryPreserve=yes
UMask=0077
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/run/wireglow-tun
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW CAP_SYS_PTRACE CAP_DAC_READ_SEARCH CAP_CHOWN

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable --now wireglow-tun.service
