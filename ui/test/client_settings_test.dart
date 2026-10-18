import 'dart:io';

import 'package:flutter/services.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grpc/grpc.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/controller/controller_repository.dart' hide State;
import 'package:wireglow_ui/controller/gen/controller.pb.dart' as pb;
import 'package:wireglow_ui/language_settings.dart';
import 'package:wireglow_ui/main.dart';
import 'package:wireglow_ui/l10n/app_localizations.dart';
import 'package:wireglow_ui/client_settings_tab.dart';

import 'connect_test.dart' show FakeRepository;
import 'desktop_golden.dart';

void main() {
  setUpAll(() async {
    final cache = '${Directory.current.parent.path}/.cache';
    final sdk = Platform.environment['FLUTTER_ROOT'] ?? '$cache/flutter';
    for (final entry in {
      'Roboto': 'Roboto-Regular.ttf',
      'MaterialIcons': 'MaterialIcons-Regular.otf',
    }.entries) {
      final loader = FontLoader(entry.key);
      loader.addFont(
        Future.value(
          ByteData.sublistView(
            await File('$sdk/bin/cache/artifacts/material_fonts/${entry.value}')
                .readAsBytes(),
          ),
        ),
      );
      await loader.load();
    }
  });
  test('addresses save without restarting active engines; reset switches only mode', () async {
    final repository = FakeRepository();
    repository.savedMode = ProxyMode.PROXY_MODE_SYSTEM;
    repository.current = Snapshot(
      state: pb.State.STATE_CONNECTED,
      profileId: 'vk',
      activeMode: ProxyMode.PROXY_MODE_SYSTEM,
    );
    final model = ConnectModel(repository);
    await model.start();
    await model.saveAddresses('127.0.0.1:28001', '127.0.0.1:28002', 28003);
    expect(repository.savedProxy, '127.0.0.1:28001');
    expect(model.preferredMode, ProxyMode.PROXY_MODE_SYSTEM);
    expect(repository.savedAPI, 28003);
    expect(repository.operations, isEmpty);
    expect(model.connected, true);
    repository.settingsError = const GrpcError.unavailable('Save failed');
    await expectLater(model.resetSettings(), throwsA(isA<GrpcError>()));
    expect(model.proxyAddress, '127.0.0.1:28001');
    expect(model.commandPending, false);
    repository.settingsError = null;
    await model.resetSettings();
    expect(model.proxyAddress, '127.0.0.1:17891');
    expect(model.relayAddress, '127.0.0.1:17890');
    expect(model.apiPort, 17892);
    expect(model.preferredMode, ProxyMode.PROXY_MODE_LOCAL);
    expect(repository.connects, 0);
    expect(repository.disconnects, 0);
    expect(repository.operations, ['switch']);
    await model.close();
  });

  testWidgets('settings layout supports narrow windows and text scaling', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(960, 1100);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final model = ConnectModel(FakeRepository());
    final languages = LanguageSettings(preference: 'ru');
    addTearDown(languages.dispose);
    await model.start();
    for (final scale in [1.0, 2.0]) {
      tester.view.physicalSize = scale == 1
          ? const Size(960, 1100)
          : const Size(420, 400);
      await tester.pumpWidget(
        MaterialApp(
          locale: const Locale('ru'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          theme: desktopTheme(Brightness.light),
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(context)
                .copyWith(textScaler: TextScaler.linear(scale)),
            child: child!,
          ),
          home: RepaintBoundary(
            key: const Key('settings-screen'),
            child: Scaffold(
              body: ClientSettingsTab(model: model, languages: languages),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      if (scale == 1) {
        await expectLater(
          find.byKey(const Key('settings-screen')),
          matchesDesktopGolden('settings'),
        );
      }
    }
    await tester.runAsync(model.close);
    await tester.pumpWidget(const SizedBox());
  }, skip: Platform.isWindows);

  testWidgets(
    'settings tab validates, saves, reports failures and confirms reset',
    (tester) async {
      tester.view.physicalSize = const Size(960, 1100);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final repository = FakeRepository();
      repository.savedProxy = '127.0.0.1:28001';
      repository.savedRelay = '127.0.0.1:28002';
      final model = ConnectModel(repository);
      await model.start();
      final languages = LanguageSettings(preference: 'ru');
      await tester.pumpWidget(WireglowApp(model: model, languages: languages));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Настройки'));
      await tester.pumpAndSettle();
      final proxy = find.byKey(const Key('proxy-address'));
      final relay = find.byKey(const Key('relay-address'));
      final api = find.byKey(const Key('api-port'));
      expect(find.text('Общее'), findsOneWidget);
      expect(find.text('Сеть'), findsOneWidget);
      expect(find.text('Релей'), findsOneWidget);
      expect(find.text('FreeTurn'), findsOneWidget);
      expect(
        (tester.widget<TextFormField>(proxy).controller!).text,
        '127.0.0.1:28001',
      );
      final save = find.widgetWithText(FilledButton, 'Сохранить');
      bool saveEnabled() => tester.widget<FilledButton>(save).onPressed != null;
      expect(saveEnabled(), false);
      await tester.enterText(proxy, '127.0.0.1:29001');
      await tester.pump();
      expect(saveEnabled(), true);
      await tester.enterText(proxy, '127.0.0.1:28001');
      await tester.pump();
      expect(saveEnabled(), false);
      await tester.enterText(proxy, ' 127.0.0.1:028001 ');
      await tester.pump();
      expect(saveEnabled(), false);
      await tester.enterText(relay, '127.0.0.1:28003');
      await tester.pump();
      expect(saveEnabled(), true);
      await tester.enterText(relay, '127.0.0.1:28002');
      await tester.pump();
      expect(saveEnabled(), false);
      await tester.enterText(api, '28001');
      await tester.pump();
      expect(saveEnabled(), true);
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(find.text('Порты должны отличаться.'), findsNWidgets(2));
      await tester.enterText(api, '28002');
      await tester.pump();
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(find.text('Порты должны отличаться.'), findsNWidgets(2));
      await tester.enterText(api, '65536');
      await tester.pump();
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(find.text('Введите порт от 1 до 65535.'), findsOneWidget);
      await tester.enterText(api, '28003');
      await tester.enterText(proxy, '0.0.0.0:9000');
      await tester.pump();
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(
        find.text('Введите 127.0.0.1 и порт от 1 до 65535.'),
        findsOneWidget,
      );
      await tester.enterText(proxy, '127.0.0.1:28002');
      await tester.pump();
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(find.text('Порты должны отличаться.'), findsNWidgets(2));
      await tester.enterText(proxy, '127.0.0.1:29001');
      await tester.pump();
      repository.settingsError = const GrpcError.unavailable('Save failed');
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(find.text('Save failed'), findsOneWidget);
      expect(saveEnabled(), true);
      expect(repository.savedProxy, '127.0.0.1:28001');
      repository.settingsError = null;
      await tester.tap(find.text('Сохранить'));
      await tester.pumpAndSettle();
      expect(repository.savedProxy, '127.0.0.1:29001');
      expect(repository.savedAPI, 28003);
      expect(saveEnabled(), false);
      expect(repository.connects, 0);
      await tester.tap(find.text('Вернуть значения по умолчанию'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Отмена'));
      await tester.pumpAndSettle();
      expect(repository.savedProxy, '127.0.0.1:29001');
      await tester.tap(find.text('Вернуть значения по умолчанию'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Сбросить'));
      await tester.pumpAndSettle();
      expect(repository.savedProxy, '127.0.0.1:17891');
      expect(
        (tester.widget<TextFormField>(relay).controller!).text,
        '127.0.0.1:17890',
      );
      expect(repository.savedAPI, 17892);
      expect((tester.widget<TextFormField>(api).controller!).text, '17892');
      expect(languages.preference, 'system');
      expect(
        tester
            .widget<FilledButton>(find.widgetWithText(FilledButton, 'Save'))
            .onPressed,
        isNull,
      );
      expect(model.profiles, hasLength(2));
      expect(tester.takeException(), isNull);
      await tester.runAsync(model.close);
      await tester.pumpWidget(const SizedBox());
      languages.dispose();
    },
  );
}
