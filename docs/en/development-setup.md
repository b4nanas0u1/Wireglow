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
The build refuses to replace an existing portable directory: move it with its
user data first. When updating, preserve user files in `data/` while replacing
Flutter resources from the new bundle. See the bundle README for details.


The script downloads the pinned network engines and builds the controller and UI. Go, Dart, generator, and engine caches stay in `.cache/` in the repository root. The Flutter SDK is resolved from `.cache/flutter/` or `PATH`. No adjacent project directories are required.


When a tag is published, GitHub Actions builds both Linux variants and attaches
`Wireglow-<version>-linux-x64.zip` and `Wireglow-<version>-linux-x64-portable.zip`
to the release. Each ZIP contains the client files and directories directly at
its root, without an enclosing distribution directory. The release `SHA256SUMS`
contains archive checksums. Local build scripts only create directories.

## Common commands

From the repository root:

```bash
go test ./...
```

Build the full distribution with the platform script above. The build scripts run `flutter pub get` themselves. To work on only the Flutter UI, change to `ui/` and use regular Flutter commands, such as `flutter run -d windows` or `flutter run -d linux`.

The currently supported local build targets are Windows x64 and Linux x64. The Bash and PowerShell build scripts are temporary and independent; a shared cross-platform build system has not been selected yet.
