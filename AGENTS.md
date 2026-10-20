# Agent instructions

## Project

Wireglow is a desktop client for managing connections through the external
free-turn-proxy and mihomo engines. The backend is Go; the desktop UI is Flutter
and Dart. Current build targets are Linux x64 and Windows x64.

Start with [README_en.md](README_en.md) and the
[development setup guide](docs/en/development-setup.md). This repository must be
usable on its own: instructions, scripts, and checks must not depend on adjacent
checkouts, private documents, or a contributor's local directory layout.

## Code map

- `cmd/controller/`: controller executable, CLI commands, desktop entry point,
  and process lifecycle integration.
- `cmd/engine-manifest/`: utility for reading pinned engine metadata.
- `internal/domain/`: profile and proxy mode models.
- `internal/freeturn/`: FreeTurn connection model, parsing, and validation.
- `internal/application/`: controller use cases, session state, and runtime ports.
- `internal/backend/`: composition of application services and infrastructure.
- `internal/infrastructure/`: engines, persistence, networking, launcher, and
  platform adapters.
- `internal/rpc/`: gRPC server, mapping, and generated Go bindings.
- `proto/controller.proto`: local controller API schema.
- `ui/`: Flutter application, Dart controller client, and UI tests.
- `scripts/`: generation, engine downloads, builds, and distribution support.
- `.github/workflows/`: CI checks, builds, and release packaging.

## Architecture and behavior

Keep session decisions in the Go application layer. The Flutter UI communicates
with the controller through its local gRPC API; it must not manage network
engines or OS proxy settings directly. Keep gRPC mapping in `internal/rpc`,
technology-specific operations in infrastructure, and wiring in the backend
and executable entry points. Application code uses runtime ports rather than
concrete platform or engine adapters.

The controller owns active session state. Preserve owner tracking, cancellation,
process cleanup, and independent proxy recovery when changing lifecycle code.
OS proxy changes must be journaled and restore only settings owned by Wireglow.
Changing between local and system proxy modes must keep an active connection
and its engine processes running.

Use the current connection model and parser as the baseline. Do not add fields,
compatibility aliases, migrations, or unsupported protocol options merely to
match upstream examples or older descriptions. TUN and routing features listed
on the roadmap are not implemented capabilities.

## Development and checks

Use the Go version required by `go.mod` and the Flutter version pinned in
`.github/workflows/build.yml`. See the development guide for native build
dependencies. On Bash, `source scripts/build-env.sh` selects repository-local
caches and adds `.cache/flutter/bin` to `PATH` when that SDK is installed.

Format changed Go files with `gofmt` and changed Dart files with `dart format`.
Run checks appropriate to the change. Standard backend checks, from the root:

```sh
go vet ./...
go test -count=1 -timeout 120s ./...
```

For platform changes, CI also runs `GOOS=linux go vet ./...` and
`GOOS=windows go vet ./...` from Linux. Cross-compilation or vet does not replace
running tests on the target OS.

Standard UI checks, from `ui/`:

```sh
flutter pub get
flutter analyze --no-pub
flutter test --no-pub --reporter expanded
```

For changes to lifecycle, proxy recovery, engine configuration, or IPC, use the
integration and Dart-to-Go smoke commands in the development guide. Ordinary
Go tests skip integrations that require explicit environment variables.
Fake-engine tests do not establish that a real remote connection works.

Full bundles are built from the repository root with:

- Linux: `./scripts/build-linux.sh`
- Linux portable: `./scripts/build-linux-portable.sh`
- Windows PowerShell: `./scripts/build-windows.ps1`
- Windows portable: `./scripts/build-windows-portable.ps1`

These scripts are currently platform-specific; a shared cross-platform build
system has not been selected. A Flutter-only build does not validate the full
controller and engine bundle.

Add or adjust tests for changed behavior. For intentional visual changes,
inspect rendered results before updating golden files. Follow
[the golden test instructions](ui/test/goldens/README.md); references are
host-specific. Do not update goldens or relax assertions just to hide a failure.

## Generated files and dependencies

Change `proto/controller.proto` before regenerating API bindings. From the root,
run `./scripts/generate-grpc.sh`, then `./scripts/generate-dart-grpc.sh`.
The Dart script currently uses the Linux x64 protoc cache prepared by the Go
script. Update both clients with API changes and preserve reserved protobuf
field numbers. Do not hand-edit generated bindings in `internal/rpc/gen/` or
`ui/lib/controller/gen/`.

Edit English and Russian UI strings in `ui/lib/l10n/app_en.arb` and
`app_ru.arb`; run `flutter gen-l10n` from `ui/` to regenerate localization classes.
Keep both languages consistent when adding or changing user-facing text.

Engine versions and download metadata belong in
`internal/infrastructure/engines/manifest/manifest.yaml`. Keep scripts and UI
metadata using this shared source. Make dependency upgrades explicit and keep
the corresponding Go or Dart dependency and lock files consistent.

## Documentation translations

When changing a README or documentation, update every existing language version
of the affected document in the same task. This also applies when a human has
edited the document as part of the current task: review those changes and bring
its translations up to date before reporting completion.

Keep `README.md` and `README_en.md` synchronized, and update matching documents
under `docs/ru/` and `docs/en/` together. Preserve equivalent meaning, working
links, and accurate commands in each language. Do not assume translations are
generated automatically; update them explicitly. If a translation cannot be
updated, report which document remains out of sync and why.

## Data and safe verification

Never add real connection links, credentials, private keys, session tokens, or
personal profiles to code, fixtures, logs, screenshots, or reports. Use synthetic
test data. Profile summaries and routine status output must not expose secrets.

Use isolated data and runtime directories for manual checks, through
`WIREGLOW_DATA_DIR` and `WIREGLOW_RUNTIME_DIR` where supported. Portable GUI mode
uses the bundle's `data/` and `data/runtime/` instead of these overrides. Preserve
existing portable user data; do not delete a populated bundle to make a build pass.
Use the isolated proxy tests rather than changing a contributor's desktop proxy
settings during routine verification.

Keep SDKs, caches, diagnostic outputs, and build artifacts in ignored locations
such as `.cache/` and `dist/`. Preserve unrelated working-tree changes.

When a tool fails, record the command, working directory, executable/version,
and relevant error without exposing secrets. Distinguish environment failures
from application failures. Report what changed, which checks passed, and any
checks that remain unverified. Do not describe a skipped integration or a
cross-compiled binary as a successful native test.
