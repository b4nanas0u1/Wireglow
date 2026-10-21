import 'dart:async';
import 'dart:io';

import 'package:flutter/services.dart';
import 'package:fixnum/fixnum.dart';
import 'package:grpc/grpc.dart';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/app_info.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/controller/controller_repository.dart' hide State;
import 'package:wireglow_ui/controller/gen/controller.pb.dart' as pb;
import 'package:wireglow_ui/main.dart';
import 'package:wireglow_ui/l10n/app_localizations.dart';

import 'desktop_golden.dart';

class FakeRepository extends ControllerRepository {
  final events = StreamController<Snapshot>.broadcast();
  final lost = StreamController<void>.broadcast();
  ProxyMode savedMode = ProxyMode.PROXY_MODE_LOCAL;
  String savedProxy = '127.0.0.1:17891';
  String savedRelay = '127.0.0.1:17890';
  int savedAPI = 17892;
  Object? settingsError;
  ProxyMode? connectedMode;
  String unavailableReason = '';
  Object? modeError;
  Completer<Snapshot>? switchGate;
  Object? switchError;
  final operations = <String>[];
  @override
  Future<ClientSettings> settings() async => ClientSettings(
    preferredMode: savedMode,
    proxyAddress: savedProxy,
    relayAddress: savedRelay,
    apiPort: savedAPI,
  );
  @override
  Future<ClientSettings> setSettings(ClientSettings value) async {
    if (settingsError != null) throw settingsError!;
    savedMode = value.preferredMode;
    savedProxy = value.proxyAddress;
    savedRelay = value.relayAddress;
    savedAPI = value.apiPort;
    return settings();
  }

  @override
  Future<ClientSettings> setMode(ProxyMode mode) async {
    if (modeError != null) throw modeError!;
    savedMode = mode;
    return settings();
  }

  @override
  Future<Capabilities> capabilities() async =>
      Capabilities(systemProxyUnavailableReason: unavailableReason);
  int connects = 0;
  int disconnects = 0;
  int latencyChecks = 0;
  Completer<Snapshot>? latencyGate;
  @override
  Future<Snapshot> checkLatency() async {
    latencyChecks++;
    return latencyGate == null ? current : await latencyGate!.future;
  }

  final imports = <String>[];
  final deletions = <String>[];
  Object? deleteError;
  Completer<void>? deleteGate;
  Snapshot current = Snapshot(state: pb.State.STATE_IDLE);
  final items = [
    ProfileSummary(
      id: 'vk',
      name: 'VK Звонки',
      provider: 'vk',
      peer: 'relay.example.com:443',
      streams: Int64(8),
      streamsPerCredential: Int64(2),
    ),
    ProfileSummary(
      id: 'direct',
      name: 'Direct',
      provider: 'direct',
      peer: '203.0.113.20:443',
    ),
  ];
  @override
  Future<void> start() async {}
  @override
  Stream<void> get death => lost.stream;
  @override
  Future<List<ProfileSummary>> profiles() async => items;
  @override
  Future<Snapshot> snapshot() async => current;
  @override
  Stream<Snapshot> watch() => events.stream;
  @override
  Future<ProfileSummary> import(String link) async {
    imports.add(link);
    final imported = ProfileSummary(
      id: 'imported',
      name: 'Новое подключение',
      provider: 'direct',
      peer: 'relay.example.org:443',
    );
    items.add(imported);
    return imported;
  }

  @override
  Future<void> connect(
    String id, {
    ProxyMode mode = ProxyMode.PROXY_MODE_LOCAL,
  }) async {
    operations.add("connect");
    connectedMode = mode;
    connects++;
  }

  @override
  Future<Snapshot> disconnect() async {
    operations.add("disconnect");
    disconnects++;
    return current = Snapshot(state: pb.State.STATE_IDLE);
  }

  @override
  Future<Snapshot> switchMode(ProxyMode mode) async {
    operations.add('switch');
    if (switchError != null) throw switchError!;
    if (switchGate != null) return current = await switchGate!.future;
    current = current.deepCopy()..activeMode = mode;
    return current;
  }

  @override
  Future<void> deleteProfile(String id) async {
    deletions.add(id);
    if (deleteGate != null) await deleteGate!.future;
    if (deleteError != null) throw deleteError!;
    items.removeWhere((p) => p.id == id);
  }

  @override
  Future<void> close() async {
    await events.close();
    await lost.close();
  }
}

void main() {
  testWidgets(
    'header proxy selector switches OS settings without reconnecting',
    (tester) async {
      final repository = FakeRepository();
      repository.current = Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'vk',
        activeMode: ProxyMode.PROXY_MODE_LOCAL,
      );
      final model = ConnectModel(repository);
      await model.start();
      await tester.pumpWidget(
        MaterialApp(
          locale: const Locale('ru'),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          home: ConnectScreen(model: model),
        ),
      );
      await tester.tap(find.byKey(const Key('proxy-mode-selector')));
      await tester.pumpAndSettle();
      await tester.tap(
        find.widgetWithText(PopupMenuItem<ProxyMode>, 'Системный прокси'),
      );
      await tester.pumpAndSettle();
      expect(repository.savedMode, ProxyMode.PROXY_MODE_SYSTEM);
      expect(find.text('Переподключить'), findsNothing);
      expect(repository.operations, ['switch']);
      expect(model.snapshot.activeMode, ProxyMode.PROXY_MODE_SYSTEM);
      expect(repository.connects, 0);
      expect(repository.disconnects, 0);
      await tester.pumpWidget(const SizedBox());
      await tester.runAsync(model.close);
    },
  );

  test(
    'failed OS switch keeps the session; failed save preserves selection',
    () async {
      final repository = FakeRepository();
      repository.current = Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'vk',
        activeMode: ProxyMode.PROXY_MODE_LOCAL,
      );
      final model = ConnectModel(repository);
      await model.start();
      repository.modeError = StateError('save failed');
      await model.selectMode(ProxyMode.PROXY_MODE_SYSTEM);
      expect(model.preferredMode, ProxyMode.PROXY_MODE_LOCAL);
      expect(repository.disconnects, 0);
      repository.modeError = null;
      repository.switchError = StateError('settings change failed');
      await model.selectMode(ProxyMode.PROXY_MODE_SYSTEM);
      expect(repository.connects, 0);
      expect(model.snapshot.activeMode, ProxyMode.PROXY_MODE_LOCAL);
      await model.close();
    },
  );

  test('default is local; unavailable system proxy is rejected', () async {
    final repository = FakeRepository()..unavailableReason = 'Unavailable';
    final model = ConnectModel(repository);
    await model.start();
    expect(model.preferredMode, ProxyMode.PROXY_MODE_LOCAL);
    await model.selectMode(ProxyMode.PROXY_MODE_SYSTEM);
    expect(model.preferredMode, ProxyMode.PROXY_MODE_LOCAL);
    expect(repository.connects, 0);
    expect(model.error, 'Unavailable');
    await model.close();
  });
  test(
    'mode change waits for OS update and rejects repeated commands',
    () async {
      final repository = FakeRepository();
      repository.current = Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'vk',
        activeMode: ProxyMode.PROXY_MODE_LOCAL,
      );
      repository.switchGate = Completer<Snapshot>();
      final model = ConnectModel(repository);
      await model.start();
      final switching = model.selectMode(ProxyMode.PROXY_MODE_SYSTEM);
      await Future<void>.delayed(Duration.zero);
      expect(model.switchingMode, true);
      expect(repository.operations, ['switch']);
      expect(model.snapshot.activeMode, ProxyMode.PROXY_MODE_LOCAL);
      await model.selectMode(ProxyMode.PROXY_MODE_LOCAL);
      await model.action();
      expect(repository.operations, ['switch']);
      repository.switchGate!.complete(
        Snapshot(
          state: pb.State.STATE_CONNECTED,
          profileId: 'vk',
          activeMode: ProxyMode.PROXY_MODE_SYSTEM,
        ),
      );
      await switching;
      expect(repository.operations, ['switch']);
      expect(model.snapshot.activeMode, ProxyMode.PROXY_MODE_SYSTEM);
      expect(repository.connects, 0);
      expect(repository.disconnects, 0);
      expect(model.switchingMode, false);
      await model.close();
    },
  );

  test(
    'idle mode change saves preference without starting a session',
    () async {
      final repository = FakeRepository();
      final model = ConnectModel(repository);
      await model.start();
      await model.selectMode(ProxyMode.PROXY_MODE_SYSTEM);
      expect(repository.savedMode, ProxyMode.PROXY_MODE_SYSTEM);
      expect(repository.operations, isEmpty);
      await model.close();
    },
  );

  test('OS update failure preserves active cores and is shown', () async {
    final repository = FakeRepository();
    repository.current = Snapshot(
      state: pb.State.STATE_CONNECTED,
      profileId: 'vk',
      activeMode: ProxyMode.PROXY_MODE_SYSTEM,
    );
    repository.savedMode = ProxyMode.PROXY_MODE_SYSTEM;
    repository.switchError = StateError('settings change failed');
    final model = ConnectModel(repository);
    await model.start();
    await model.selectMode(ProxyMode.PROXY_MODE_LOCAL);
    expect(repository.operations, ['switch']);
    expect(model.snapshot.activeMode, ProxyMode.PROXY_MODE_SYSTEM);
    expect(repository.connects, 0);
    expect(repository.disconnects, 0);
    expect(model.error, 'settings change failed');
    expect(model.connected, true);
    expect(model.commandPending, false);
    await model.close();
  });

  testWidgets('about tab is read-only and preserves selection while offline', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(960, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = FakeRepository();
    final model = ConnectModel(repository);
    await model.start();
    model.select(repository.items.last);
    model.available = false;
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale("ru"),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        theme: desktopTheme(Brightness.light),
        home: ConnectScreen(model: model),
      ),
    );
    await tester.tap(find.text('О клиенте'));
    await tester.pumpAndSettle();
    expect(find.text('Версия'), findsOneWidget);
    expect(find.text(appVersion), findsOneWidget);
    expect(find.text('Тег'), findsNothing);
    expect(
      tester.getTopLeft(find.text('Версия')).dy,
      lessThan(tester.getTopLeft(find.text('Коммит')).dy),
    );
    expect(
      tester.getTopLeft(find.text('Коммит')).dy,
      lessThan(tester.getTopLeft(find.text('Время сборки')).dy),
    );
    expect(find.text('Коммит'), findsOneWidget);
    expect(find.text('Время сборки'), findsOneWidget);
    expect(find.text('mihomo'), findsOneWidget);
    expect(find.text('free-turn'), findsOneWidget);
    expect(find.byType(TextField), findsNothing);
    expect(find.text('https://github.com/MetaCubeX/mihomo'), findsOneWidget);
    expect(
      find.text('https://github.com/hackdiaz-dev/free-turn-proxy'),
      findsOneWidget,
    );
    expect(repository.connects, 0);
    expect(repository.disconnects, 0);
    await expectLater(
      find.byType(ConnectScreen),
      matchesDesktopGolden('about'),
    );
    // The content also scrolls without overflowing a small desktop window.
    tester.view.physicalSize = const Size(420, 400);
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.tap(find.text('Подключения'));
    await tester.pumpAndSettle();
    expect(model.selected!.id, 'direct');
    await tester.pumpWidget(const SizedBox());
    await tester.runAsync(model.close);
  });

  testWidgets('latency refresh, failure and disconnect', (tester) async {
    final repository = FakeRepository();
    repository.current = Snapshot(
      state: pb.State.STATE_CONNECTED,
      profileId: 'vk',
      latency: Latency(delayMs: Int64(123), successAt: Int64(1700000000000)),
    );
    final model = ConnectModel(repository);
    await model.start();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale("ru"),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        theme: desktopTheme(Brightness.light),
        home: ConnectScreen(model: model),
      ),
    );
    expect(find.text('Задержка: 123 мс'), findsOneWidget);
    repository.latencyGate = Completer<Snapshot>();
    await tester.tap(find.byTooltip('Проверить задержку'));
    await tester.pump();
    expect(find.text('Проверяем задержку…'), findsOneWidget);
    await model.checkLatency();
    expect(repository.latencyChecks, 1);
    repository.latencyGate!.complete(
      Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'vk',
        latency: Latency(
          delayMs: Int64(123),
          successAt: Int64(1700000000000),
          error: 'HTTPS check failed',
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Проверка не прошла'), findsOneWidget);
    expect(find.textContaining('Последний успех:'), findsOneWidget);
    expect(model.connected, true);
    await model.action();
    await tester.pumpAndSettle();
    expect(find.byTooltip('Проверить задержку'), findsNothing);
    expect(model.snapshot.latency.successAt.toInt(), 0);
    await tester.pumpWidget(const SizedBox());
    await tester.runAsync(model.close);
  });

  setUpAll(() async {
    final cache = '${Directory.current.parent.path}/.cache';
    final root = Platform.environment['FLUTTER_ROOT'] ?? '$cache/flutter';
    for (final entry in {
      'Roboto': 'Roboto-Regular.ttf',
      'MaterialIcons': 'MaterialIcons-Regular.otf',
    }.entries) {
      final loader = FontLoader(entry.key);
      loader.addFont(
        Future.value(
          ByteData.sublistView(
            await File(
              '$root/bin/cache/artifacts/material_fonts/${entry.value}',
            ).readAsBytes(),
          ),
        ),
      );
      await loader.load();
    }
  });
  test('ACK does not mark Connected; startup can be cancelled', () async {
    final repository = FakeRepository();
    final model = ConnectModel(repository);
    await model.start();
    await model.action();
    expect(repository.connects, 1);
    expect(model.connected, false);
    repository.current = Snapshot(
      state: pb.State.STATE_STARTING_RELAY,
      profileId: 'vk',
    );
    repository.events.add(repository.current);
    await Future<void>.delayed(Duration.zero);
    model.select(repository.items.last);
    expect(model.selected!.id, 'vk');
    await model.action();
    expect(repository.disconnects, 1);
    expect(model.snapshot.state, pb.State.STATE_IDLE);
    await model.close();
  });
  testWidgets('adding a link saves it and connects only on explicit action', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(960, 650);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = FakeRepository();
    final model = ConnectModel(repository);
    await model.start();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale("ru"),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        theme: desktopTheme(Brightness.light),
        home: ConnectScreen(model: model),
      ),
    );
    await tester.tap(find.text('Добавить по ссылке'));
    await tester.pumpAndSettle();
    expect(find.text('Добавить и подключиться'), findsNothing);
    expect(find.byType(TextField), findsOneWidget);
    await tester.enterText(find.byType(TextField), '  freeturn://test  ');
    await tester.tap(find.text('Добавить'));
    await tester.pumpAndSettle();
    expect(find.byType(AddProfileDialog), findsNothing);
    expect(repository.imports, ['freeturn://test']);
    expect(model.profiles.map((p) => p.id), contains('imported'));
    expect(model.selected!.id, 'imported');
    expect(find.text('Новое подключение'), findsWidgets);
    expect(model.snapshot.state, pb.State.STATE_IDLE);
    expect(model.commandPending, false);
    expect(repository.connects, 0);
    expect(repository.disconnects, 0);
    await tester.tap(find.byTooltip('Подключить'));
    await tester.pumpAndSettle();
    expect(repository.connects, 1);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
    await tester.runAsync(model.close);
  });
  test(
    'deletion updates selection, preserves failures and clears last profile',
    () async {
      final repository = FakeRepository();
      final model = ConnectModel(repository);
      await model.start();
      repository.deleteError = GrpcError.unavailable('Backend unavailable');
      await model.deleteProfile('direct');
      expect(model.profiles, hasLength(2));
      expect(model.selected!.id, 'vk');
      expect(model.error, 'Backend unavailable');
      expect(model.commandPending, false);
      repository.deleteError = null;
      await model.deleteProfile('direct');
      expect(model.selected!.id, 'vk');
      expect(model.error, isNull);
      await model.deleteProfile('vk');
      expect(model.profiles, isEmpty);
      expect(model.selected, isNull);
      await model.close();
    },
  );
  test(
    'deletion blocks repeated commands and is disabled during a session',
    () async {
      final repository = FakeRepository();
      final model = ConnectModel(repository);
      await model.start();
      repository.deleteGate = Completer<void>();
      final deletion = model.deleteProfile('vk');
      expect(model.commandPending, true);
      await model.deleteProfile('vk');
      await model.action();
      expect(repository.deletions, ['vk']);
      expect(repository.connects, 0);
      repository.deleteGate!.complete();
      await deletion;
      expect(model.selected!.id, 'direct');
      model.snapshot = Snapshot(
        state: pb.State.STATE_CONNECTED,
        profileId: 'direct',
      );
      await model.deleteProfile('direct');
      model.available = false;
      model.snapshot = Snapshot(state: pb.State.STATE_IDLE);
      await model.deleteProfile('direct');
      expect(repository.deletions, ['vk']);
      await model.close();
    },
  );
  testWidgets('profile deletion requires confirmation and handles empty list', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(960, 650);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = FakeRepository();
    final model = ConnectModel(repository);
    await model.start();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale("ru"),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        theme: desktopTheme(Brightness.light),
        home: ConnectScreen(model: model),
      ),
    );
    await tester.tap(find.byTooltip('Действия с профилем').first);
    await tester.pumpAndSettle();
    expect(find.text('Редактировать'), findsOneWidget);
    await tester.tap(find.text('Удалить'));
    await tester.pumpAndSettle();
    expect(find.text('Удалить профиль?'), findsOneWidget);
    expect(find.textContaining('«VK Звонки»'), findsOneWidget);
    await tester.tap(find.text('Отмена'));
    await tester.pumpAndSettle();
    expect(repository.deletions, isEmpty);
    for (final id in ['vk', 'direct']) {
      await tester.tap(find.byTooltip('Действия с профилем').first);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Удалить'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Удалить'));
      await tester.pumpAndSettle();
      expect(repository.deletions.last, id);
    }
    expect(find.byTooltip('Действия с профилем'), findsNothing);
    expect(find.text('Удалить профиль'), findsNothing);
    expect(find.text('Добавьте ссылку для подключения'), findsOneWidget);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox());
    await tester.runAsync(model.close);
  });
  test('lost RPC clears confirmed connection', () async {
    final repository = FakeRepository();
    repository.current = Snapshot(
      state: pb.State.STATE_CONNECTED,
      profileId: 'vk',
    );
    final model = ConnectModel(repository);
    await model.start();
    expect(model.connected, true);
    repository.lost.add(null);
    await Future<void>.delayed(Duration.zero);
    expect(model.connected, false);
    expect(model.available, false);
    await model.close();
  });
  testWidgets('desktop layout, Direct properties and scale remain usable', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(960, 650);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    final repository = FakeRepository();
    final model = ConnectModel(repository);
    await model.start();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale("ru"),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        theme: desktopTheme(Brightness.light),
        home: RepaintBoundary(
          key: const Key('screen'),
          child: ConnectScreen(model: model),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('vk — relay.example.com:443 — 8/2'), findsOneWidget);
    await expectLater(
      find.byKey(const Key('screen')),
      matchesDesktopGolden('connect'),
    );
    await tester.tap(find.text('Direct'));
    await tester.pumpAndSettle();
    expect(find.text('direct — 203.0.113.20:443'), findsOneWidget);
    await tester.tap(find.text('Свойства подключения'));
    await tester.pumpAndSettle();
    expect(find.text('Не применяется'), findsOneWidget);
    expect(find.text('Потоки'), findsNothing);
    for (final size in [const Size(640, 480), const Size(480, 350)]) {
      tester.view.physicalSize = size;
      await tester.pumpWidget(
        MaterialApp(
          locale: const Locale("ru"),
          supportedLocales: AppLocalizations.supportedLocales,
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(context)
                .copyWith(textScaler: const TextScaler.linear(2)),
            child: child!,
          ),
          home: ConnectScreen(model: model),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    }
    await tester.pumpWidget(const SizedBox());
    await tester.runAsync(model.close);
  });
}
