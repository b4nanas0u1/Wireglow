#!/usr/bin/env bash
set -euo pipefail

# Package only a freshly built regular Linux bundle. Never copy a user's data/.
scripts_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
root=$(cd -- "$scripts_root/.." && pwd)
metadata="$scripts_root/linux/debian"
bundle="$root/dist/linux-x64"
if (( $# > 1 )) || [[ $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  echo "Usage: $0 [Debian-version] (Linux x86_64 only)" >&2
  exit 1
fi
version=${1:-}
if [[ -z "$version" ]]; then
  tag=$(git -C "$root" describe --tags --exact-match HEAD 2>/dev/null || true)
  if [[ "$tag" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+([~+.-][0-9A-Za-z.+~-]+)?$ ]]; then
    version=${tag#v}
    version=${version/-/\~}
  else
    version=$(sed -n 's/^version: //p' "$root/ui/pubspec.yaml")
    version=${version/-/\~}
    version+="+git$(git -C "$root" log -1 --format=%ct).$(git -C "$root" rev-parse --short=8 HEAD)"
  fi
  version+="-1"
fi
# Restrict the filename as well as validating Debian's version syntax.
if [[ ! "$version" =~ ^[0-9][0-9A-Za-z.+~\-]*$ ]] || ! dpkg --validate-version "$version"; then
  echo "Invalid Debian version: $version" >&2
  exit 1
fi
if [[ -e "$bundle/portable.json" ]]; then
  echo 'Refusing to package a portable bundle.' >&2
  exit 1
fi
IFS= read -r app_name < "$bundle/data/app-name.txt"
if [[ ! "$app_name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
  echo 'Invalid application name.' >&2
  exit 1
fi
for executable in "$app_name" bin/controller bin/free-turn-proxy bin/mihomo; do
  [[ -x "$bundle/$executable" ]] || { echo "Missing executable: $executable" >&2; exit 1; }
done

export SOURCE_DATE_EPOCH=${SOURCE_DATE_EPOCH:-$(git -C "$root" log -1 --format=%ct)}
output="$root/dist/debian"
mkdir -p -- "$output"
staging=$(mktemp -d "$output/.package.XXXXXX")
trap 'rm -rf -- "$staging"' EXIT
package="$staging/package"
app="$package/usr/lib/wireglow"
docs="$package/usr/share/doc/wireglow"
install -d -m 755 "$app/bin" "$app/data" "$docs/licenses" "$package/DEBIAN" \
  "$package/usr/bin" "$package/usr/share/applications" \
  "$package/usr/share/lintian/overrides" "$package/usr/share/man/man1" \
  "$package/usr/share/icons/hicolor/256x256/apps"
install -m 755 "$bundle/$app_name" "$app/$app_name"
for executable in controller free-turn-proxy mihomo; do
  install -m 755 "$bundle/bin/$executable" "$app/bin/$executable"
done
# Strip only our executables. Preserve pinned upstream engine bytes unchanged.
strip --strip-unneeded "$app/$app_name" "$app/bin/controller"
cp -a -- "$bundle/lib" "$app/lib"
cp -a -- "$bundle/data/flutter_assets" "$app/data/flutter_assets"
for asset in icudtl.dat app-name.txt app-icon.png; do
  install -m 644 "$bundle/data/$asset" "$app/data/$asset"
done
# A package may contain only regular files/directories: no symlink escapes.
if [[ -n $(find "$package" ! -type d ! -type f -print -quit) ]]; then
  echo 'Unexpected symlink or special file in bundle.' >&2
  exit 1
fi
find "$package" -type d -exec chmod 755 {} +
find "$app/lib" "$app/data" -type f -exec chmod 644 {} +
cat > "$package/usr/bin/wireglow" <<EOF
#!/bin/sh
exec /usr/lib/wireglow/$app_name "\$@"
EOF
chmod 755 "$package/usr/bin/wireglow"
install -m 644 "$metadata/wireglow.desktop" "$package/usr/share/applications/dev.wireglow.wireglow_ui.desktop"
install -m 644 "$bundle/data/app-icon.png" "$package/usr/share/icons/hicolor/256x256/apps/dev.wireglow.wireglow_ui.png"
install -m 644 "$metadata/README.Debian" "$docs/README.Debian"
install -m 644 "$metadata/copyright" "$docs/copyright"
install -m 644 "$metadata/lintian-overrides" "$package/usr/share/lintian/overrides/wireglow"
gzip -n -9 -c "$metadata/wireglow.1" > "$package/usr/share/man/man1/wireglow.1.gz"
cat > "$docs/changelog.Debian" <<EOF
wireglow ($version) trixie; urgency=medium

  * Package the Wireglow UI, controller and pinned network engines.

 -- Wireglow contributors <b4nanas0u1@users.noreply.github.com>  $(LC_ALL=C date -u -d "@$SOURCE_DATE_EPOCH" -R)
EOF
gzip -n -9 "$docs/changelog.Debian"
install -m 644 "$metadata/licenses/"* "$docs/licenses/"
gzip -n -9 "$docs/licenses/"*
install -m 644 "$root/internal/infrastructure/engines/manifest/manifest.yaml" "$docs/engines.yaml"

# Resolve system dependencies of every ELF, including downloaded engines.
# dpkg-shlibdeps needs a source control file, but no source package is built.
install -d "$staging/debian"
cp "$metadata/control" "$staging/debian/control"
elf_args=()
while IFS= read -r -d '' file; do
  if readelf -h "$file" &>/dev/null && readelf -d "$file" 2>/dev/null | grep -q '(NEEDED)'; then
    elf_args+=("-e$file")
  fi
done < <(find "$app" -type f -print0)
dependencies=$(
  cd -- "$staging"
  dpkg-shlibdeps -O "-l$app/lib" "${elf_args[@]}" | sed -n 's/^shlibs:Depends=//p'
)
[[ -n "$dependencies" ]] || { echo 'No ELF dependencies resolved.' >&2; exit 1; }
installed_size=$(du -sk "$package/usr" | cut -f 1)
cat > "$package/DEBIAN/control" <<EOF
Package: wireglow
Version: $version
Section: net
Priority: optional
Architecture: amd64
Maintainer: Wireglow contributors <b4nanas0u1@users.noreply.github.com>
Installed-Size: $installed_size
Depends: $dependencies, libegl1, libgl1, ca-certificates, systemd, dbus-user-session
Recommends: gsettings-desktop-schemas, dconf-gsettings-backend
Homepage: https://github.com/b4nanas0u1/Wireglow
Description: Desktop client for Free Turn Proxy connections
 Flutter UI, Go controller, and pinned network engines.
 Requires a graphical session with a running systemd user manager.
EOF
(
  cd -- "$package"
  find usr -type f -print0 | sort -z | xargs -0 md5sum > DEBIAN/md5sums
)
# Do not replace an existing artifact until assembly has succeeded.
filename="wireglow_${version}_amd64.deb"
dpkg-deb --root-owner-group -Zxz --build "$package" "$staging/$filename"
mv -f -- "$staging/$filename" "$output/$filename"
(
  cd -- "$output"
  sha256sum -- "$filename" > "$filename.sha256"
)
echo "Package: $output/$filename"
