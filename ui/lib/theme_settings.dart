import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';

import 'ui_settings_file.dart';

/// Appearance is stored separately so language and theme writes cannot collide.
class ThemeSettings extends ChangeNotifier {
  ThemeSettings({this.file});

  final File? file;
  ThemeMode _mode = ThemeMode.system;
  ThemeMode get mode => _mode;
  Future<void> _writes = Future.value();

  static Future<ThemeSettings> load({
    File? settingsFile,
    Directory? configurationDirectory,
  }) async {
    final file =
        settingsFile ??
        uiSettingsFile(
          'theme-settings.json',
          configurationDirectory: configurationDirectory,
        );
    if (file == null) return ThemeSettings();
    try {
      final json = jsonDecode(await file.readAsString());
      final mode = ThemeMode.values.firstWhere(
        (mode) => mode.name == json['theme'],
        orElse: () => ThemeMode.system,
      );
      return ThemeSettings(file: file).._mode = mode;
    } on Object {
      return ThemeSettings(file: file);
    }
  }

  Future<void> select(ThemeMode mode) {
    _mode = mode;
    notifyListeners();
    final target = file;
    if (target == null) return Future.value();
    final write = _writes.catchError((Object _) {}).then((_) async {
      await target.parent.create(recursive: true);
      final temporary = File('${target.path}.tmp');
      await temporary.writeAsString(
        jsonEncode({'theme': mode.name}),
        flush: true,
      );
      await temporary.rename(target.path);
    });
    _writes = write;
    return write;
  }
}
