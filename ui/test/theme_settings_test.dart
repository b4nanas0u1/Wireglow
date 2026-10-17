import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/controller/controller_repository.dart' hide State;
import 'package:wireglow_ui/controller/gen/controller.pb.dart' as pb;
import 'package:wireglow_ui/language_settings.dart';
import 'package:wireglow_ui/main.dart';
import 'package:wireglow_ui/theme_settings.dart';

import 'connect_test.dart' show FakeRepository;

void main() {
  test('theme survives restart, ordered writes and invalid settings', () async {
    final dir = await Directory.systemTemp.createTemp('wireglow-theme-');
    addTearDown(() => dir.delete(recursive: true));
    final file = File('${dir.path}/theme.json');
    final settings = ThemeSettings(file: file);
    addTearDown(settings.dispose);
    await Future.wait([
      settings.select(ThemeMode.dark),
      settings.select(ThemeMode.light),
    ]);
    final restored = await ThemeSettings.load(settingsFile: file);
    addTearDown(restored.dispose);
    expect(restored.mode, ThemeMode.light);
    for (final content in ['invalid', '{"theme":"unknown"}', '{}']) {
      await file.writeAsString(content);
      final fallback = await ThemeSettings.load(settingsFile: file);
      expect(fallback.mode, ThemeMode.system);
      fallback.dispose();
    }
  });

  testWidgets(
    'theme menu matches language and switches live without disconnecting',
    (tester) async {
      tester.view.physicalSize = const Size(1000, 800);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      addTearDown(tester.platformDispatcher.clearPlatformBrightnessTestValue);
      tester.platformDispatcher.platformBrightnessTestValue = Brightness.light;
      final repository = FakeRepository();
      repository.current = Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'vk',
      );
      final model = ConnectModel(repository);
      await model.start();
      final languages = LanguageSettings(preference: 'ru');
      final appearance = ThemeSettings();
      addTearDown(languages.dispose);
      addTearDown(appearance.dispose);
      await tester.pumpWidget(
        WireglowApp(model: model, languages: languages, appearance: appearance),
      );
      await tester.pumpAndSettle();
      Brightness brightness() =>
          Theme.of(tester.element(find.byType(ConnectScreen))).brightness;
      expect(brightness(), Brightness.light);
      expect(
        tester.getCenter(find.byKey(const Key('theme-selector'))).dx,
        greaterThan(
          tester.getCenter(find.byKey(const Key('language-selector'))).dx,
        ),
      );

      await tester.tap(find.byKey(const Key('language-selector')));
      await tester.pumpAndSettle();
      final languageRow = tester.getSize(
        find.byType(PopupMenuItem<String>).first,
      );
      await tester.tap(find.widgetWithText(PopupMenuItem<String>, 'Русский'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('theme-selector')));
      await tester.pumpAndSettle();
      expect(
        tester.getSize(find.byType(PopupMenuItem<ThemeMode>).first),
        languageRow,
      );
      expect(find.byType(PopupMenuItem<ThemeMode>), findsNWidgets(3));
      await tester.tap(find.widgetWithText(PopupMenuItem<ThemeMode>, 'Тёмная'));
      await tester.pumpAndSettle();
      expect(brightness(), Brightness.dark);
      await tester.tap(find.text('О клиенте'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('theme-selector')));
      await tester.pumpAndSettle();
      await tester.tap(
        find.widgetWithText(PopupMenuItem<ThemeMode>, 'Светлая'),
      );
      await tester.pumpAndSettle();
      expect(brightness(), Brightness.light);
      expect(find.text('Версия клиента'), findsOneWidget);
      await tester.tap(find.byKey(const Key('theme-selector')));
      await tester.pumpAndSettle();
      await tester.tap(
        find.widgetWithText(PopupMenuItem<ThemeMode>, 'Системная'),
      );
      await tester.pumpAndSettle();
      tester.platformDispatcher.platformBrightnessTestValue = Brightness.dark;
      await tester.pumpAndSettle();
      expect(brightness(), Brightness.dark);
      tester.platformDispatcher.platformBrightnessTestValue = Brightness.light;
      await tester.pumpAndSettle();
      expect(brightness(), Brightness.light);
      expect(model.connected, isTrue);
      expect(model.selected!.id, 'vk');
      expect(repository.connects, 0);
      expect(repository.disconnects, 0);
      tester.view.physicalSize = const Size(420, 400);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.runAsync(model.close);
      await tester.pumpWidget(const SizedBox());
    },
  );
}
