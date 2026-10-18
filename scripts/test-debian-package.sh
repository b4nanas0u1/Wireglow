#!/usr/bin/env bash
set -euo pipefail

# This installs/removes packages: run only in a disposable Debian container.
if [[ ${WIREGLOW_PACKAGE_TEST:-} != 1 || $EUID != 0 || $# != 1 ]]; then
  echo "Usage (disposable container, root): WIREGLOW_PACKAGE_TEST=1 $0 package.deb" >&2
  exit 1
fi
deb=$(realpath -- "$1")
if [[ $(dpkg-deb -f "$deb" Package) != wireglow ]] || dpkg-query -W wireglow &>/dev/null; then
  echo 'Expected a Wireglow package and a container without Wireglow installed.' >&2
  exit 1
fi
temporary=$(mktemp -d)
trap 'rm -rf -- "$temporary"' EXIT
chmod 755 "$temporary"

apt-get install -y --no-install-recommends "$deb"
desktop-file-validate /usr/share/applications/dev.wireglow.wireglow_ui.desktop
[[ ! -e /usr/lib/wireglow/portable.json ]]
[[ ! -e /usr/lib/wireglow/data/profiles ]]
[[ ! -e /usr/lib/wireglow/data/runtime ]]
[[ -x /usr/bin/wireglow ]]
[[ -z $(find /usr/lib/wireglow -perm /022 -print -quit) ]]
[[ -z $(find /usr/lib/wireglow ! -user root -print -quit) ]]
while IFS= read -r -d '' file; do
  if readelf -h "$file" &>/dev/null && readelf -d "$file" 2>/dev/null | grep -q '(NEEDED)'; then
    ldd "$file" > "$temporary/ldd.txt"
    if grep -q 'not found' "$temporary/ldd.txt"; then
      cat "$temporary/ldd.txt" >&2
      exit 1
    fi
  fi
done < <(find /usr/lib/wireglow -type f -print0)

useradd --create-home --shell /bin/sh wireglow-package-test
install -d -m 700 -o wireglow-package-test -g wireglow-package-test \
  /home/wireglow-package-test/.config/wireglow \
  /home/wireglow-package-test/runtime
settings=/home/wireglow-package-test/.config/wireglow/ui-settings.json
printf '{"language":"ru"}\n' > "$settings"
chown wireglow-package-test:wireglow-package-test "$settings"
chmod 600 "$settings"
# No desktop/systemd is needed to list profiles. This exercises the installed
# controller as an unprivileged user, using the existing default config path.
su -s /bin/sh wireglow-package-test -c \
  'cd /tmp && XDG_CONFIG_HOME=/home/wireglow-package-test/.config /usr/lib/wireglow/bin/controller list --runtime-dir /home/wireglow-package-test/runtime'
# Synthetic retention fixture; it is not parsed as a connection profile.
profile=/home/wireglow-package-test/.config/wireglow/profiles/retention.json
printf '{"package_test":true}\n' > "$profile"
chown wireglow-package-test:wireglow-package-test "$profile"
chmod 600 "$profile"
before=$(sha256sum "$settings" "$profile")
/usr/lib/wireglow/bin/mihomo -v
/usr/lib/wireglow/bin/free-turn-proxy --help > "$temporary/relay-help.txt" 2>&1

# Exercise a real version upgrade without rebuilding the same payload twice.
dpkg-deb -R "$deb" "$temporary/upgrade"
version=$(dpkg-deb -f "$deb" Version)
sed -i "s/^Version: .*/Version: ${version}+packagetest/" "$temporary/upgrade/DEBIAN/control"
dpkg-deb --root-owner-group --build "$temporary/upgrade" "$temporary/upgrade.deb"
apt-get install -y --no-install-recommends "$temporary/upgrade.deb"
[[ $(dpkg-query -W -f='${Version}' wireglow) == "$version+packagetest" ]]
[[ $(sha256sum "$settings" "$profile") == "$before" ]]
apt-get remove -y wireglow
[[ ! -e /usr/bin/wireglow ]]
[[ ! -e /usr/lib/wireglow/Wireglow ]]
[[ $(sha256sum "$settings" "$profile") == "$before" ]]
apt-get install -y --no-install-recommends "$deb"
apt-get purge -y wireglow
[[ ! -e /usr/share/applications/dev.wireglow.wireglow_ui.desktop ]]
[[ $(sha256sum "$settings" "$profile") == "$before" ]]
userdel --remove wireglow-package-test
echo 'Debian package install, ELF dependencies, CLI, upgrade, remove and purge passed.'
