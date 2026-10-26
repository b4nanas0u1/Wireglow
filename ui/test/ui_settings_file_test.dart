import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/language_settings.dart';
import 'package:wireglow_ui/theme_settings.dart';

void main() {
  test('fresh preferences save only in wireglow', () async {
    final base = await Directory.systemTemp.createTemp('wireglow-ui-fresh-');
    addTearDown(() => base.delete(recursive: true));
    final language = await LanguageSettings.load(configurationDirectory: base);
    final theme = await ThemeSettings.load(configurationDirectory: base);
    addTearDown(language.dispose);
    addTearDown(theme.dispose);
    expect(language.preference, 'system');
    expect(theme.mode, ThemeMode.system);
    await language.select('ru');
    await theme.select(ThemeMode.dark);
    expect(
      await File('${base.path}/wireglow/ui-settings.json').exists(),
      isTrue,
    );
    expect(
      await File('${base.path}/wireglow/theme-settings.json').exists(),
      isTrue,
    );
    final restoredLanguage = await LanguageSettings.load(
      configurationDirectory: base,
    );
    final restoredTheme = await ThemeSettings.load(
      configurationDirectory: base,
    );
    addTearDown(restoredLanguage.dispose);
    addTearDown(restoredTheme.dispose);
    expect(restoredLanguage.preference, 'ru');
    expect(restoredTheme.mode, ThemeMode.dark);

    await language.file!.writeAsString('invalid');
    await theme.file!.writeAsString('invalid');
    final fallbackLanguage = await LanguageSettings.load(
      configurationDirectory: base,
    );
    final fallbackTheme = await ThemeSettings.load(
      configurationDirectory: base,
    );
    addTearDown(fallbackLanguage.dispose);
    addTearDown(fallbackTheme.dispose);
    expect(fallbackLanguage.preference, 'system');
    expect(fallbackTheme.mode, ThemeMode.system);
  });
}
