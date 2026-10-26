import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/controller/bundle_paths.dart';
import 'package:wireglow_ui/ui_error.dart';

void main() {
  late Directory bundle;
  final suffix = Platform.isWindows ? '.exe' : '';

  setUp(() async {
    bundle = await Directory.systemTemp.createTemp('Wireglow пробел ');
    await Directory.fromUri(bundle.uri.resolve('bin/')).create();
    for (final name in ['controller', 'free-turn-proxy', 'mihomo']) {
      await File.fromUri(bundle.uri.resolve('bin/$name$suffix')).create();
    }
  });
  tearDown(() => bundle.delete(recursive: true));

  test('finds a complete bundle relative to executable, not cwd', () async {
    final paths = BundlePaths(
      executable: File.fromUri(bundle.uri.resolve('Wireglow$suffix')).path,
      environment: {},
    );
    expect(
      paths.controller,
      File.fromUri(bundle.uri.resolve('bin/controller$suffix')).path,
    );
    await paths.validate();
    await File(paths.relay).delete();
    await expectLater(paths.validate(), throwsA(isA<UiError>()));
  });

  test('development overrides replace only the selected component', () async {
    final replacement = File.fromUri(bundle.uri.resolve('custom$suffix'));
    await replacement.create();
    final paths = BundlePaths(
      executable: File.fromUri(bundle.uri.resolve('Wireglow$suffix')).path,
      environment: {
        'WIREGLOW_CONTROLLER': replacement.path,
        'WIREGLOW_RELAY': '',
      },
    );
    expect(paths.controller, replacement.path);
    expect(
      paths.relay,
      File.fromUri(bundle.uri.resolve('bin/free-turn-proxy$suffix')).path,
    );
    await paths.validate();
  });
}
