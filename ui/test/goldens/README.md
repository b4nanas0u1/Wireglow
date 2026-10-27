# Desktop goldens

The images in this directory are the Linux references. Windows tests
compare against `windows/` because Flutter text rasterization differs by host,
even when the same Roboto and MaterialIcons fonts are loaded from the SDK.
Comparisons remain pixel-exact; behavioral and layout assertions run on all hosts.

New GUI references must use English. `profile-editor-en.png` covers the collapsed
profile editor at a 960×650 desktop size. Its screenshot comparison currently runs
on Linux only; the editor's behavioral and responsive-layout tests run on all hosts.
Generate a native Windows reference before enabling its pixel comparison there.

Regenerate and visually review references on each supported host after intentional
visual changes. The normal test run must pass without `--update-goldens`.
Do not copy Linux images into `windows/`: Windows needs its own rasterization baseline.

To regenerate on the respective host, run from `ui/`:

```sh
flutter test --no-pub --dart-define=APP_VERSION=0.1.0 --update-goldens test/connect_test.dart test/client_settings_test.dart
flutter test --no-pub --dart-define=APP_VERSION=0.1.0 --update-goldens test/edit_profile_test.dart
flutter test --no-pub --dart-define=APP_VERSION=0.1.0
```

After intentional visual changes, review failure images before updating the
references on each supported host. Do not increase comparison tolerance to hide
rendering or layout changes.

Golden runs fix APP_VERSION to 0.1.0 to keep build metadata reproducible. The
default dev version and release values are checked separately by app_info_test.dart.
