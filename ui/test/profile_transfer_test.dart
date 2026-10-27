import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grpc/grpc.dart';
import 'package:wireglow_ui/connect_model.dart';
import 'package:wireglow_ui/controller/controller_repository.dart' hide State;
import 'package:wireglow_ui/l10n/app_localizations.dart';
import 'package:wireglow_ui/main.dart' show desktopTheme;
import 'package:wireglow_ui/profile_transfer_dialog.dart';

import 'connect_test.dart' show FakeRepository;

class Picker extends ProfileTransferFiles {
  String? input = '/tmp/wireglow-profiles.json';
  String? output = '/tmp/export.json';
  @override
  Future<String?> open() async => input;
  @override
  Future<String?> save() async => output;
}

class TransferRepository extends FakeRepository {
  TransferRepository() {
    items.first.name = 'VK Calls';
  }
  List<String>? exported;
  List<ProfileImportChoice>? imported;
  Object? previewError;
  Object? exportError;
  Completer<ImportProfileFileResponse>? gate;
  bool partial = false;
  @override
  Future<void> exportProfiles(List<String> ids, String path) async {
    if (exportError != null) throw exportError!;
    exported = ids;
  }

  @override
  Future<ProfileFilePreview> previewProfileFile(String path) async {
    if (previewError != null) throw previewError!;
    return ProfileFilePreview(
      token: 'preview-token',
      entries: [
        ProfileFileEntry(
          profile: ProfileSummary(id: 'vk', name: 'VK Calls', provider: 'vk'),
          existing: items.first,
        ),
        ProfileFileEntry(
          profile: ProfileSummary(
            id: 'new',
            name: 'Travel',
            provider: 'direct',
          ),
        ),
      ],
    );
  }

  @override
  Future<ImportProfileFileResponse> importProfileFile(
    String token,
    List<ProfileImportChoice> choices,
  ) async {
    imported = choices;
    if (gate != null) return gate!.future;
    if (!partial) {
      items.add(ProfileSummary(id: 'new', name: 'Travel', provider: 'direct'));
    }
    return ImportProfileFileResponse(
      results: [
        ProfileImportResult(index: 0, status: partial ? 'changed' : 'skipped'),
        ProfileImportResult(index: 1, status: partial ? 'failed' : 'imported'),
      ],
    );
  }
}

Future<ConnectModel> showTransfer(
  WidgetTester tester,
  TransferRepository repository,
  Picker picker, {
  bool exporting = false,
  Size size = const Size(960, 650),
  Locale locale = const Locale('en'),
  double scale = 1,
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
      builder: (context, child) => MediaQuery(
        data: MediaQuery.of(context)
            .copyWith(textScaler: TextScaler.linear(scale)),
        child: child!,
      ),
      home: Scaffold(
        body: Builder(
          builder: (context) => TextButton(
            onPressed: () => showDialog<String>(
              context: context,
              barrierDismissible: false,
              builder: (_) => ProfileTransferDialog(
                model: model,
                exporting: exporting,
                files: picker,
              ),
            ),
            child: const Text('Open'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('Open'));
  await tester.pumpAndSettle();
  addTearDown(() async {
    await tester.pumpWidget(const SizedBox());
    await model.close();
  });
  return model;
}

void main() {
  setUpAll(() async {
    final sdk = Platform.environment['FLUTTER_ROOT']!;
    for (final entry in {
      'Roboto': 'Roboto-Regular.ttf',
      'MaterialIcons': 'MaterialIcons-Regular.otf',
    }.entries) {
      final loader = FontLoader(entry.key)
        ..addFont(
          Future.value(
            ByteData.sublistView(
              await File(
                '$sdk/bin/cache/artifacts/material_fonts/${entry.value}',
              ).readAsBytes(),
            ),
          ),
        );
      await loader.load();
    }
  });
  testWidgets(
    'export selection, cancelled picker and failed write keep dialog usable',
    (tester) async {
      final repository = TransferRepository();
      final picker = Picker();
      await showTransfer(tester, repository, picker, exporting: true);
      if (Platform.isLinux) {
        await expectLater(
          find.byType(AlertDialog),
          matchesGoldenFile('goldens/profile-export-en.png'),
        );
      }
      await tester.tap(find.text('Select all'));
      await tester.pump();
      expect(
        tester
            .widget<FilledButton>(
              find.byKey(const Key('confirm-profile-export')),
            )
            .onPressed,
        isNull,
      );
      await tester.tap(find.byKey(const Key('export-profile-vk')));
      await tester.pump();
      picker.output = null;
      await tester.tap(find.byKey(const Key('confirm-profile-export')));
      await tester.pumpAndSettle();
      expect(repository.exported, isNull);
      expect(find.byType(AlertDialog), findsOneWidget);
      picker.output = '/tmp/export.json';
      repository.exportError = const GrpcError.internal(
        'do not show secret path',
      );
      await tester.tap(find.byKey(const Key('confirm-profile-export')));
      await tester.pumpAndSettle();
      expect(find.textContaining('do not show secret'), findsNothing);
      repository.exportError = null;
      await tester.tap(find.byKey(const Key('confirm-profile-export')));
      await tester.pumpAndSettle();
      expect(repository.exported, ['vk']);
      expect(find.byType(AlertDialog), findsNothing);
    },
  );
  testWidgets(
    'import previews conflicts and refreshes the model without connection calls',
    (tester) async {
      final repository = TransferRepository();
      final picker = Picker();
      final model = await showTransfer(tester, repository, picker);
      await tester.tap(find.byKey(const Key('choose-profile-file')));
      await tester.pumpAndSettle();
      expect(find.text('Skip existing'), findsOneWidget);
      expect(find.text('Import 1'), findsOneWidget);
      if (Platform.isLinux) {
        await expectLater(
          find.byType(AlertDialog),
          matchesGoldenFile('goldens/profile-import-en.png'),
        );
      }
      await tester.tap(find.byKey(const Key('profile-import-action-0')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Create a copy').last);
      await tester.pumpAndSettle();
      expect(find.text('Import 2'), findsOneWidget);
      await tester.tap(find.byKey(const Key('confirm-profile-import')));
      await tester.pumpAndSettle();
      expect(
        repository.imported!.first.action,
        ProfileImportAction.PROFILE_IMPORT_ACTION_COPY,
      );
      expect(model.profiles, hasLength(3));
      expect(repository.connects, 0);
      expect(repository.disconnects, 0);
    },
  );
  testWidgets('cancel and invalid archive do not import or expose raw errors', (
    tester,
  ) async {
    final repository = TransferRepository();
    final picker = Picker()..input = null;
    await showTransfer(tester, repository, picker);
    await tester.tap(find.byKey(const Key('choose-profile-file')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('confirm-profile-import')), findsNothing);
    picker.input = '/tmp/invalid.json';
    repository.previewError = const GrpcError.invalidArgument('secret json');
    await tester.tap(find.byKey(const Key('choose-profile-file')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Invalid Wireglow file'), findsOneWidget);
    expect(find.text('secret json'), findsNothing);
    expect(repository.imported, isNull);
  });
  testWidgets(
    'partial failure shows per-profile results and requires a new preview',
    (tester) async {
      final repository = TransferRepository()..partial = true;
      await showTransfer(tester, repository, Picker());
      await tester.tap(find.byKey(const Key('choose-profile-file')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('confirm-profile-import')));
      await tester.pumpAndSettle();
      expect(find.text('Changed after preview; not imported'), findsOneWidget);
      expect(find.text('Could not save this profile'), findsOneWidget);
      expect(find.byKey(const Key('confirm-profile-import')), findsNothing);
    },
  );
  testWidgets('small Russian dialogs support text scaling and pending state', (
    tester,
  ) async {
    final repository = TransferRepository()
      ..gate = Completer<ImportProfileFileResponse>();
    await showTransfer(
      tester,
      repository,
      Picker(),
      size: const Size(420, 600),
      locale: const Locale('ru'),
      scale: 2,
    );
    await tester.ensureVisible(find.byKey(const Key('choose-profile-file')));
    await tester.tap(find.byKey(const Key('choose-profile-file')));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.tap(find.byKey(const Key('confirm-profile-import')));
    await tester.pump();
    expect(
      tester
          .widget<FilledButton>(find.byKey(const Key('confirm-profile-import')))
          .onPressed,
      isNull,
    );
    expect(tester.takeException(), isNull);
    repository.gate!.complete(
      ImportProfileFileResponse(
        results: [ProfileImportResult(index: 1, status: 'imported')],
      ),
    );
    await tester.pumpAndSettle();
  });
}
