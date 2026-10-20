import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/language_settings.dart';
import 'package:wireglow_ui/main.dart';
import 'package:grpc/grpc.dart';

import 'connect_test.dart' show FakeRepository;

void main() {
  test('system language resolution always falls back to English', () {
    const supported = [Locale('en'), Locale('ru')];
    expect(
      LanguageSettings.resolve([const Locale('ru', 'RU')], supported),
      const Locale('ru'),
    );
    expect(
      LanguageSettings.resolve([const Locale('en', 'GB')], supported),
      const Locale('en'),
    );
    expect(
      LanguageSettings.resolve([const Locale('de')], supported),
      const Locale('en'),
    );
    expect(LanguageSettings.resolve(null, supported), const Locale('en'));
  });

  test(
    'preference survives restart and corrupted settings fall back to system',
    () async {
      final dir = await Directory.systemTemp.createTemp('wireglow-locale-');
      addTearDown(() => dir.delete(recursive: true));
      final file = File('${dir.path}/settings.json');
      final settings = LanguageSettings(file: file);
      await Future.wait([settings.select('ru'), settings.select('en')]);
      expect(jsonDecode(await file.readAsString())['language'], 'en');
      final restored = await LanguageSettings.load(settingsFile: file);
      expect(restored.locale, const Locale('en'));
      await file.writeAsString('invalid');
      expect(
        (await LanguageSettings.load(settingsFile: file)).preference,
        'system',
      );
      settings.dispose();
      restored.dispose();
    },
  );

  testWidgets('switch languages live without losing connection, tab or form', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1000, 800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = FakeRepository();
    final model = ConnectModel(repository);
    await model.start();
    final languages = LanguageSettings(preference: 'en');
    await tester.pumpWidget(WireglowApp(model: model, languages: languages));
    await tester.pumpAndSettle();
    expect(find.text('Connections'), findsOneWidget);
    await tester.tap(find.byKey(const Key('language-selector')));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(PopupMenuItem<String>, 'Русский'));
    await tester.pumpAndSettle();
    expect(find.text('Подключения'), findsOneWidget);
    expect(
      tester.getCenter(find.byKey(const Key('proxy-mode-selector'))).dx,
      lessThan(tester.getCenter(find.byKey(const Key('language-selector'))).dx),
    );
    tester.view.physicalSize = const Size(420, 400);
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    tester.view.physicalSize = const Size(1000, 800);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Добавить по ссылке'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField).first, 'freeturn://draft');
    await languages.select('en');
    await tester.pumpAndSettle();
    expect(find.text('Add profile'), findsOneWidget);
    expect(find.text('freeturn://draft'), findsOneWidget);
    expect(model.selected!.id, 'vk');
    expect(repository.connects, 0);
    expect(repository.disconnects, 0);
    await tester.tap(find.byTooltip('Close add profile'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('About'));
    await tester.pumpAndSettle();
    await languages.select('ru');
    await tester.pumpAndSettle();
    expect(find.text('Версия'), findsOneWidget);
    await tester.runAsync(model.close);
    await tester.pumpWidget(const SizedBox());
    languages.dispose();
  });

  test('backend errors are preserved verbatim', () {
    expect(
      ConnectModel.message(GrpcError.invalidArgument('Invalid FreeTurn link')),
      'Invalid FreeTurn link',
    );
  });
}
