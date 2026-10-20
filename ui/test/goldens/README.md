# Desktop goldens

The images in this directory are the Linux references. Windows tests
compare against `windows/` because Flutter text rasterization differs by host,
even when the same Roboto and MaterialIcons fonts are loaded from the SDK.
Comparisons remain pixel-exact; behavioral and layout assertions run on all hosts.

Linux references were regenerated and visually reviewed on 2026-10-17 with
Flutter 3.47.7 / Dart 3.13.5. Connect, Settings and About use the Wireglow warm
light palette; Connect and About show the approved icon and Wireglow name.
The normal test run must pass without `--update-goldens` after regeneration.

Windows references still date from 2026-10-14 and contain the previous design.
The three widget tests comparing Settings, About, and Connect goldens are
temporarily skipped on Windows until these references are updated and reviewed.
They continue to run on Linux.
They need regeneration and visual review on an actual Windows host; this Linux
update does not validate them. Do not copy Linux images into `windows/`, because
that would replace the Windows rasterization baseline with the wrong host's
output. No Windows build or test run was performed for this update.

To regenerate on the respective host, run from `ui/`:

```sh
flutter test --no-pub --dart-define=APP_VERSION=0.1.0 --update-goldens test/connect_test.dart test/client_settings_test.dart
flutter test --no-pub --dart-define=APP_VERSION=0.1.0
```

After intentional visual changes, review failure images before updating the
references on each supported host. Do not increase comparison tolerance to hide
rendering or layout changes.

Golden runs fix APP_VERSION to 0.1.0 to keep build metadata reproducible. The
default dev version and release values are checked separately by app_info_test.dart.
