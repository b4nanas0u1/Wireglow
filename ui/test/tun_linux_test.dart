import 'dart:async';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:flutter/material.dart';
import 'package:wireglow_ui/main.dart';
import 'package:wireglow_ui/client_settings_tab.dart';
import 'package:wireglow_ui/l10n/app_localizations.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/controller/controller_repository.dart' hide State;

import 'connect_test.dart' show FakeRepository;

import 'package:wireglow_ui/controller/gen/controller.pb.dart' as pb;

class TunnelRepository extends FakeRepository {
  String tunReason = '';
  Completer<Capabilities>? gate;
  bool fail = false;
  int checks = 0;
  @override
  Future<Capabilities> capabilities() async {
    checks++;
    if (fail) throw Exception('capability check failed');
    if (gate != null) return gate!.future;
    return Capabilities(tunUnavailableReason: tunReason);
  }
}

void main() {
  testWidgets(
    'Linux shows TUN and disables activation when helper is unavailable',
    (tester) async {
      final repository = TunnelRepository();
      final model = ConnectModel(repository);
      await model.start();
      repository.tunReason = 'helper unavailable';
      await tester.pumpWidget(
        MaterialApp(
          locale: const Locale('en'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          home: ConnectScreen(model: model),
        ),
      );
      await tester.tap(find.byKey(const Key('proxy-mode-selector')));
      await tester.pumpAndSettle();
      final item = find.widgetWithText(PopupMenuItem<ProxyMode>, 'TUN');
      expect(item, findsOneWidget);
      expect(tester.widget<PopupMenuItem<ProxyMode>>(item).enabled, isFalse);
      expect(repository.operations, isEmpty);
      await tester.pumpWidget(const SizedBox());
      await tester.runAsync(model.close);
    },
    skip: !Platform.isLinux,
    variant: TargetPlatformVariant({TargetPlatform.linux}),
  );

  for (final settings in [false, true]) {
    testWidgets(
      'refreshes TUN before opening ${settings ? "settings" : "main"} menu',
      (tester) async {
        final repository = TunnelRepository();
        final model = ConnectModel(repository);
        await model.start();
        await tester.pumpWidget(
          MaterialApp(
            locale: const Locale('en'),
            supportedLocales: AppLocalizations.supportedLocales,
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            home: settings
                ? Scaffold(body: ClientSettingsTab(model: model))
                : ConnectScreen(model: model),
          ),
        );
        final selector = find.byKey(
          Key(
            settings ? 'settings-proxy-mode-selector' : 'proxy-mode-selector',
          ),
        );
        repository.gate = Completer<Capabilities>();
        await tester.tap(selector);
        await tester.pump();
        expect(find.byType(PopupMenuItem<ProxyMode>), findsNothing);
        expect(repository.checks, 2);
        repository.gate!.complete(
          Capabilities(tunUnavailableReason: 'active VPN (amn0)'),
        );
        await tester.pumpAndSettle();
        final item = find.widgetWithText(PopupMenuItem<ProxyMode>, 'TUN');
        expect(tester.widget<PopupMenuItem<ProxyMode>>(item).enabled, isFalse);
        await tester.tapAt(const Offset(5, 5));
        await tester.pumpAndSettle();
        repository.gate = null;
        repository.tunReason = '';
        await tester.tap(selector);
        await tester.pumpAndSettle();
        expect(tester.widget<PopupMenuItem<ProxyMode>>(item).enabled, isTrue);
        expect(repository.checks, 3);
        await tester.tapAt(const Offset(5, 5));
        await tester.pumpAndSettle();
        repository.fail = true;
        await tester.tap(selector);
        await tester.pumpAndSettle();
        expect(find.byType(PopupMenuItem<ProxyMode>), findsNothing);
        expect(repository.operations, isEmpty);
        await tester.pumpWidget(const SizedBox());
        await tester.runAsync(model.close);
      },
      skip: !Platform.isLinux,
      variant: TargetPlatformVariant({TargetPlatform.linux}),
    );
  }

  test(
    'TUN selection reconnects with explicit mode and bypasses OS-only switch',
    () async {
      final repository = FakeRepository();
      repository.current = Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'vk',
        activeMode: ProxyMode.PROXY_MODE_LOCAL,
      );
      final model = ConnectModel(repository);
      await model.start();
      await model.selectMode(ProxyMode.PROXY_MODE_TUN);
      expect(repository.operations, ['disconnect', 'connect']);
      expect(repository.connectedMode, ProxyMode.PROXY_MODE_TUN);
      expect(repository.savedMode, ProxyMode.PROXY_MODE_TUN);
      await model.close();
    },
    skip: !Platform.isLinux,
  );
}
