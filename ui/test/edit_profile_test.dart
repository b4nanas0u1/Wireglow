import 'dart:io';

import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/controller/controller_repository.dart' hide State;
import 'package:wireglow_ui/edit_profile_dialog.dart';
import 'package:wireglow_ui/l10n/app_localizations.dart';
import 'package:wireglow_ui/main.dart' show desktopTheme;

import 'connect_test.dart' show FakeRepository;

class EditorRepository extends FakeRepository {
  final connection = Connection(
    name: 'My server',
    provider: 'vk',
    peer: 'server.example:56000',
    vk: VKSettings(callLinks: ['https://vk.ru/call/join/synthetic']),
    dns: DNSSettings(),
    obfuscation: Obfuscation(mode: 'none'),
  );
  ProfileAdvancedEdits? saved;
  int? savedStreams, savedCredentials;
  Object? loadError;
  @override
  Future<Profile> getProfile(String id) async {
    if (loadError != null) throw loadError!;
    return Profile(id: id, connection: connection.deepCopy());
  }

  @override
  Future<ProfileSummary> updateProfile({
    required String id,
    required String name,
    required List<String> callLinks,
    required int streams,
    required int streamsPerCredential,
    ProfileAdvancedEdits? advanced,
  }) async {
    saved = advanced?.deepCopy();
    savedStreams = streams;
    savedCredentials = streamsPerCredential;
    return items.first..name = name;
  }
}

Future<ConnectModel> openEditor(
  WidgetTester tester,
  EditorRepository repository, {
  Size size = const Size(960, 720),
  Locale locale = const Locale('en'),
}) async {
  tester.view.physicalSize = size;
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final model = ConnectModel(repository);
  await model.start();
  await tester.pumpWidget(
    MaterialApp(
      locale: locale,
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      theme: desktopTheme(Brightness.light),
      home: Scaffold(
        body: Builder(
          builder: (context) => TextButton(
            onPressed: () => showDialog<bool>(
              context: context,
              builder: (_) => EditProfileDialog(
                model: model,
                summary: repository.items.first,
              ),
            ),
            child: const Text('Open editor'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('Open editor'));
  await tester.pumpAndSettle();
  addTearDown(() async {
    await tester.pumpWidget(const SizedBox());
    await model.close();
  });
  return model;
}

Finder field(String label) => find.widgetWithText(TextFormField, label);
Future<void> revealAdvanced(WidgetTester tester) async {
  await tester.ensureVisible(find.byKey(const Key('profile-advanced')));
  await tester.tap(find.text('Advanced settings'));
  await tester.pumpAndSettle();
}

void main() {
  setUpAll(() async {
    final root =
        Platform.environment['FLUTTER_ROOT'] ??
        '${Directory.current.parent.path}/.cache/flutter';
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

  testWidgets('collapsed editor fits desktop and preserves unset defaults', (
    tester,
  ) async {
    final repository = EditorRepository();
    await openEditor(tester, repository, size: const Size(960, 650));
    expect(find.text('Advanced settings').hitTestable(), findsOneWidget);
    final scrollable = find.descendant(
      of: find.byType(AlertDialog),
      matching: find.byType(Scrollable),
    );
    expect(
      tester.state<ScrollableState>(scrollable.first).position.maxScrollExtent,
      0,
    );
    expect(find.text('Node address'), findsNothing);
    expect(tester.getRect(find.text('Save')).bottom, lessThan(650));
    expect(tester.takeException(), isNull);
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(repository.savedStreams, 0);
    expect(repository.savedCredentials, 0);
    expect(repository.saved!.vkPlatform, '');
    expect(repository.saved!.turnTransport, '');
    expect(repository.saved!.dns.mode, '');
    expect(repository.saved!.pacingMs, Int64.ZERO);
  });

  testWidgets(
    'advanced controls preserve credentials and disable pacing and DNS correctly',
    (tester) async {
      final repository = EditorRepository();
      repository.connection.vk.platform = 'mobile';
      repository.connection.vk.manualCaptcha = true;
      repository.connection.vk.turnAddress = 'turn.example.org:443';
      repository.connection.dns = DNSSettings(
        mode: 'doh',
        servers: ['1.1.1.1'],
      );
      await openEditor(tester, repository);
      await revealAdvanced(tester);
      expect(find.text('Mobile'), findsOneWidget);
      expect(
        tester.widget<CheckboxListTile>(find.byType(CheckboxListTile)).value,
        isTrue,
      );
      expect(
        tester.widget<TextFormField>(field('Custom DNS servers')).enabled,
        isFalse,
      );
      expect(
        tester.widget<TextFormField>(field('Pacing, ms')).enabled,
        isFalse,
      );
      await tester.ensureVisible(field('Node address'));
      await tester.enterText(field('Node address'), '[2001:db8::1]:443');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(repository.saved!.turnAddress, '[2001:db8::1]:443');
      expect(repository.saved!.dns.servers, ['1.1.1.1']);
      expect(repository.saved!.vkPlatform, 'mobile');
    },
  );

  testWidgets(
    'pacing can be edited with obfuscation and explicit values survive',
    (tester) async {
      final repository = EditorRepository();
      repository.connection.obfuscation = Obfuscation(
        mode: 'rtpopus3',
        timingMs: Int64(20),
      );
      repository.connection.streams = Int64(8);
      repository.connection.streamsPerCredential = Int64(2);
      await openEditor(tester, repository);
      await revealAdvanced(tester);
      await tester.ensureVisible(field('Pacing, ms'));
      await tester.enterText(field('Pacing, ms'), '30');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(repository.saved!.pacingMs, Int64(30));
      expect(repository.savedStreams, 8);
      expect(repository.savedCredentials, 2);
    },
  );

  testWidgets('name validation does not save', (tester) async {
    final repository = EditorRepository();
    await openEditor(tester, repository);
    await tester.enterText(field('Profile name'), '');
    await tester.tap(find.text('Save'));
    await tester.pumpAndSettle();
    expect(find.text('Enter a profile name.'), findsOneWidget);
    expect(repository.saved, isNull);
  });

  testWidgets('narrow Russian editor lays out and scrolls without overflow', (
    tester,
  ) async {
    await openEditor(
      tester,
      EditorRepository(),
      size: const Size(480, 700),
      locale: const Locale('ru'),
    );
    await tester.ensureVisible(find.text('Дополнительные настройки'));
    await tester.tap(find.text('Дополнительные настройки'));
    await tester.pumpAndSettle();
    await tester.ensureVisible(field('Пейсинг, мс'));
    expect(tester.takeException(), isNull);
  });

  testWidgets('English GUI reference for collapsed profile editor', (
    tester,
  ) async {
    await openEditor(tester, EditorRepository(), size: const Size(960, 650));
    await expectLater(
      find.byType(AlertDialog),
      matchesGoldenFile('goldens/profile-editor-en.png'),
    );
  }, skip: !Platform.isLinux);
}
