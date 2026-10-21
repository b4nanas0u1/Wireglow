# Development environment setup

This guide covers building Wireglow from source on Windows and Linux x64. The commands assume a fresh checkout. Run development and build commands from the root of the Git repository.

## Tools to install

| Tool | Project version | Link |
|---|---:|---|
| Go | 1.27.1 or a newer compatible version required by `go.mod` | [go.dev/dl](https://go.dev/dl/) |
| Flutter (includes Dart) | 3.47.7, stable | [Install Flutter](https://docs.flutter.dev/install) |
| Git | current stable release | [git-scm.com](https://git-scm.com/downloads) |

Flutter 3.47.7 is pinned in CI. The Go modules require at least Go 1.27.1. Use these versions for reproducible development.

## Windows 11 x64

### 1. Install the tools

Install [Git for Windows](https://git-scm.com/download/win), [Go](https://go.dev/dl/), and the [Flutter SDK](https://docs.flutter.dev/install/manual). For a manual Flutter install, extract the SDK to a path without spaces, such as `C:\src\flutter`, and add `C:\src\flutter\bin` to `PATH`.

For the Flutter Windows runner, install [Visual Studio 2022 Build Tools](https://visualstudio.microsoft.com/downloads/) with the **Desktop development with C++** workload. Make sure it includes the MSVC C++ build tools, CMake tools for Windows, and the Windows 10/11 SDK. Visual Studio Code alone is not sufficient.

### 2. Verify the installation

Open a new PowerShell window:

```powershell
git --version
go version
flutter --version
flutter doctor -v
```

In `flutter doctor -v`, check that Flutter detects Visual Studio and the Windows SDK. Warnings about missing Android or web tools do not affect Wireglow's Windows build.

### 3. Get the source and build the app

```powershell
git clone <repository URL>
cd <repository directory>
./scripts/build-windows.ps1
```

Replace the placeholder with the URL from the repository's public page. If you already have the repository, change to its root and run the last command. The script downloads the pinned network engines, builds the Go controller and Flutter UI, and prepares the full bundle in `dist/windows-x64/`. Launch it with `dist\windows-x64\Wireglow.exe`.

Flutter from `PATH` is used automatically. You can also place the SDK at `.cache\flutter-windows\flutter` in the repository root; the build script will find it there. Build caches stay in the repository’s `.cache/` directory.

### Portable Windows bundle

Run `./scripts/build-windows-portable.ps1` (or `./scripts/build-windows.ps1 -Portable`).
The separate output is `dist/windows-x64-portable/`; launch `Wireglow.exe` directly.
The `portable.json` marker keeps profiles, UI/backend settings, and the proxy
recovery journal in the bundle's `data/`, with locks and launcher state
in `data/runtime/`. Portable GUI ignores `WIREGLOW_DATA_DIR` and
`WIREGLOW_RUNTIME_DIR`. The folder must be writable; write failures do not fall
back to AppData. Visual C++ runtime DLLs are included, as in the regular bundle.

An existing portable output is never replaced: move it with its data before
rebuilding. Close the application and wait for recovery before moving or updating
the bundle. Preserve user files in `data/`, replacing only shipped Flutter resources
(`app.so`, `icudtl.dat`, `flutter_assets/`) and other distribution files. Never
apply an unfinished proxy recovery journal on another computer. Verify the bundled
`SHA256SUMS` before first launch.

Only transient IPC lives outside the portable folder. Windows creates its socket
in a separate `%TEMP%/wg-<run ID>/` directory restricted to the current user and
SYSTEM; Linux uses `$XDG_RUNTIME_DIR/wg-<run ID>/` (or `/run/user/<UID>/wg-<run ID>/`),
with directory mode 0700 and socket mode 0600. The independent watchdog removes
this directory after controller shutdown, including a GUI crash. IPC no longer
depends on the extraction path; persistent data and recovery state remain portable.

CI builds both Windows variants and publishes `Wireglow-<version>-windows-x64.zip`
and `Wireglow-<version>-windows-x64-portable.zip` as artifacts and release assets.
Their files are directly at the ZIP root. Release `SHA256SUMS` includes both archives.

## Linux x64 (Ubuntu/Debian)

### 1. Install system dependencies

The Flutter Linux runner needs GTK 3, Clang, CMake, Ninja, pkg-config, and liblzma. These commands are for Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install -y git curl xz-utils clang cmake ninja-build pkg-config libgtk-3-dev liblzma-dev
```

### 2. Install Go and Flutter

Install Go 1.27.1 or newer using the [official Linux instructions](https://go.dev/doc/install). Do not rely on the version in your distribution's default package repository; it may be too old. Verify it with `go version` after installation.

Install Flutter 3.47.7 stable using the [official instructions](https://docs.flutter.dev/install/manual) or select that version from the [Flutter SDK archive](https://docs.flutter.dev/install/archive). Add `flutter/bin` to `PATH`, for example for a local install:

```bash
export PATH="$HOME/development/flutter/bin:$PATH"
```

Add this line to `~/.bashrc` or your shell's configuration file to keep the setting. Install Git if it is not already available (the package command above includes it).

### 3. Verify the installation

Open a new terminal:

```bash
git --version
go version
flutter --version
flutter doctor -v
```

Check that `flutter doctor -v` recognizes the Linux toolchain. Android and web toolchains are not needed to build Wireglow.

### 4. Get the source and build the app

```bash
git clone <repository URL>
cd <repository directory>
./scripts/build-linux.sh
```

Replace the placeholder with the URL from the repository's public page. If you already have the repository, change to its root and run `./scripts/build-linux.sh`. The full bundle will be created in `dist/linux-x64/`; launch it with `./dist/linux-x64/Wireglow`.

For a portable bundle, run `./scripts/build-linux-portable.sh`
(equivalent to `./scripts/build-linux.sh --portable`). The output is
`dist/linux-x64-portable/`.
The `portable.json` marker beside `Wireglow` selects `data/` for profiles,
UI/backend preferences and the recovery journal, and `data/runtime/` for temporary
session state. Without the marker, the existing OS directories are used.
The portable GUI ignores `WIREGLOW_DATA_DIR` and `WIREGLOW_RUNTIME_DIR`;
explicit controller CLI flags still take precedence.
Rebuilding replaces the existing portable directory in `dist/` after successful
packaging: obsolete application files are removed, user files in `data/` are
preserved, and Flutter resources are refreshed. Close the portable app first.
A failed build leaves the previous bundle intact. See the bundle README for details.


The script downloads the pinned network engines and builds the controller and UI. Go, Dart, generator, and engine caches stay in `.cache/` in the repository root. The Flutter SDK is resolved from `.cache/flutter/` or `PATH`. No adjacent project directories are required.


When a tag is published, GitHub Actions builds both Linux variants and attaches
`Wireglow-<version>-linux-x64.zip` and `Wireglow-<version>-linux-x64-portable.zip`
to the release. Each ZIP contains the client files and directories directly at
its root, without an enclosing distribution directory. The release `SHA256SUMS`
contains archive checksums. Local build scripts only create directories.

## Checks and CI

GitHub Actions runs checks and regular/portable Linux and Windows builds on
every PR update, every commit to `master`, and release publication. Manual builds
through `workflow_dispatch` remain available for `master`.
Stages run in order: lints → tests → builds → release publication.
Regular and integration tests run in parallel after both lint jobs pass;
builds wait for all tests. A failure blocks subsequent stages.
Go vet, Flutter analyze, Go tests, Flutter tests, Go integrations, and Dart↔Go IPC
appear as separate jobs. Both lint jobs run only on Linux; Go vet checks Linux
and Windows code using `GOOS`. Linux and Windows tests are separate.
Archives are retained
as workflow artifacts. Archives and `SHA256SUMS` are attached to a GitHub Release
only on `release: published`, after all checks and builds pass.

Lint commands (Linux), followed by tests (Linux and Windows):

```sh
GOOS=linux go vet ./...
GOOS=windows go vet ./...
cd ui
flutter pub get
flutter analyze --no-pub
```

```sh
go test -count=1 -timeout 120s ./...
cd ui
flutter pub get
flutter test --no-pub --reporter expanded --dart-define=APP_VERSION=0.1.0
```

Separate Linux jobs check Go integrations (systemd lifecycle, proxy recovery,
GNOME settings in an isolated D-Bus session, and the real Mihomo engine)
and Dart↔Go IPC.
Local integration runs require a working user systemd, `dbus-run-session`,
`dconf-cli`, `gsettings-desktop-schemas`, Go, and Flutter. From the repository root:

```bash
source scripts/build-env.sh
./scripts/download-linux-engines.sh
WIREGLOW_SYSTEMD_TEST=1 WIREGLOW_RECOVERY_TEST=1 WIREGLOW_GSETTINGS_TEST=1 \
  WIREGLOW_MIHOMO_TEST_BINARY="$cache_root/engines/linux-amd64/bin/mihomo" \
  go test -count=1 -timeout 180s -v ./cmd/controller ./internal/infrastructure/platform ./internal/infrastructure/engines
mkdir -p "$cache_root/bin"
go build -o "$cache_root/bin/controller" ./cmd/controller
(cd ui && flutter pub get)
./ui/tool/smoke.sh
```

Without these variables, ordinary Go tests skip the additional Linux/Mihomo
integrations. The smoke test uses fake engines and does not verify a real VPN connection.

## Common commands

From the repository root:

```bash
go test ./...
```

Build the full distribution with the platform script above. The build scripts run `flutter pub get` themselves. To work on only the Flutter UI, change to `ui/` and use regular Flutter commands, such as `flutter run -d windows` or `flutter run -d linux`.

The currently supported local build targets are Windows x64 and Linux x64. The Bash and PowerShell build scripts are temporary and independent; a shared cross-platform build system has not been selected yet.


For the optional Linux TUN helper, installation details and an isolated packet test that preserves other VPNs, see [TUN mode](tun-mode.md).

The IPC smoke keeps temporary launcher state in a short system temporary directory. Test profile data stays in the repository cache; the Linux socket uses the private system runtime described below.

On Linux, desktop controller IPC also uses a private short `$XDG_RUNTIME_DIR/wg-<run ID>` directory (or `/run/user/<UID>/wg-<run ID>`), independent of the bundle path length. Portable profiles, locks, lease files and recovery journals remain inside the bundle; the worker and watchdog clean up the transient socket directory.

Portable Linux TUN uses a temporary polkit helper, with no service installation or autostart. The optional `install-tun-helper.sh` is shipped only with the regular Linux bundle. See [TUN mode](tun-mode.md).

The About tab shows the exact Git tag pointing at the built commit (including
any `v` prefix), or `dev` when no tag points at it. Local scripts resolve the tag
from Git; CI fetches tags and full history. Flutter builds run directly without
`APP_VERSION` also show `dev`. The Flutter package version in `pubspec.yaml`
is not the displayed client version.

Golden test commands pin `APP_VERSION=0.1.0` to preserve the visual references.
To check the default and tagged metadata separately, run from `ui/`:

```sh
flutter test --no-pub test/app_info_test.dart
flutter test --no-pub --dart-define=APP_VERSION=v1.2.3 test/app_info_test.dart
```
