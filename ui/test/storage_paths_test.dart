import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/storage_paths.dart';
import 'package:wireglow_ui/ui_settings_file.dart';
import 'package:wireglow_ui/language_settings.dart';
import 'package:wireglow_ui/theme_settings.dart';

void main() {
  test('portable keeps backend, runtime and UI preferences together', () async {
    final root = await Directory.systemTemp.createTemp('Wireglow пробел ');
    addTearDown(() => root.delete(recursive: true));
    await File('${root.path}/portable.json').create();
    final paths = StoragePaths(
      executable: '${root.path}/Wireglow',
      environment: {
        'WIREGLOW_DATA_DIR': '/unused-data',
        'WIREGLOW_RUNTIME_DIR': '/unused-runtime',
      },
    );
    expect(paths.portable, isTrue);
    expect(paths.controllerArguments, [
      '--data-dir',
      '${root.path}/data',
      '--runtime-dir',
      '${root.path}/data/runtime',
    ]);
    final languageFile = uiSettingsFile(
      'ui-settings.json',
      storagePaths: paths,
    )!;
    final themeFile = uiSettingsFile(
      'theme-settings.json',
      storagePaths: paths,
    )!;
    expect(languageFile.path, '${root.path}/data/ui-settings.json');
    expect(themeFile.path, '${root.path}/data/theme-settings.json');
    final language = await LanguageSettings.load(settingsFile: languageFile);
    await language.select('ru');
    final theme = await ThemeSettings.load(settingsFile: themeFile);
    await theme.select(ThemeMode.dark);
    expect(
      (await LanguageSettings.load(settingsFile: languageFile)).preference,
      'ru',
    );
    expect(
      (await ThemeSettings.load(settingsFile: themeFile)).mode,
      ThemeMode.dark,
    );
    // Moving the whole bundle preserves preferences without absolute stored paths.
    final moved = await root.rename('${root.path}-moved');
    addTearDown(() async {
      if (await moved.exists()) await moved.delete(recursive: true);
    });
    final movedPaths = StoragePaths(executable: '${moved.path}/Wireglow');
    expect(
      (await LanguageSettings.load(
        settingsFile: uiSettingsFile(
          'ui-settings.json',
          storagePaths: movedPaths,
        ),
      )).preference,
      'ru',
    );
    await moved.rename(root.path);
  });

  test('without marker defaults and explicit directory overrides remain', () {
    final defaults = StoragePaths(
      executable: '${Directory.systemTemp.path}/no-bundle/Wireglow',
      environment: {},
    );
    expect(defaults.portable, isFalse);
    expect(defaults.controllerArguments, isEmpty);
    final overridden = StoragePaths(
      executable: '${Directory.systemTemp.path}/no-bundle/Wireglow',
      environment: {
        'WIREGLOW_DATA_DIR': 'custom-data',
        'WIREGLOW_RUNTIME_DIR': 'custom-runtime',
      },
    );
    expect(overridden.data, File('custom-data').absolute.path);
    expect(overridden.runtime, File('custom-runtime').absolute.path);
  });
}
