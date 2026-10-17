import 'dart:convert';
import 'dart:io';

import 'package:flutter/widgets.dart';

import 'ui_settings_file.dart';

/// System follows the OS; unsupported locales always fall back to English.
class LanguageSettings extends ChangeNotifier {
  LanguageSettings({this.file, String preference = 'system'})
    : _preference = supported.contains(preference) ? preference : 'system';
  static const supported = ['system', 'en', 'ru'];
  final File? file;
  String _preference;
  String get preference => _preference;
  Locale? get locale => _preference == 'system' ? null : Locale(_preference);
  Future<void> _writes = Future.value();

  static Locale resolve(List<Locale>? locales, Iterable<Locale> supported) {
    for (final locale in locales ?? <Locale>[]) {
      if (locale.languageCode == 'ru') return const Locale('ru');
      if (locale.languageCode == 'en') return const Locale('en');
    }
    return const Locale('en');
  }

  static Future<LanguageSettings> load({
    File? settingsFile,
    Directory? configurationDirectory,
  }) async {
    final file =
        settingsFile ??
        uiSettingsFile(
          'ui-settings.json',
          configurationDirectory: configurationDirectory,
        );
    if (file == null) return LanguageSettings();
    try {
      final json = jsonDecode(await file.readAsString());
      return LanguageSettings(
        file: file,
        preference: json['language'] as String,
      );
    } on Object {
      return LanguageSettings(file: file);
    }
  }

  Future<void> select(String preference) {
    if (!supported.contains(preference)) throw ArgumentError.value(preference);
    _preference = preference;
    notifyListeners();
    final target = file;
    if (target == null) return Future.value();
    final write = _writes.catchError((Object _) {}).then((_) async {
      await target.parent.create(recursive: true);
      final temporary = File('${target.path}.tmp');
      await temporary.writeAsString(
        jsonEncode({'language': preference}),
        flush: true,
      );
      await temporary.rename(target.path);
    });
    _writes = write;
    return write;
  }
}
