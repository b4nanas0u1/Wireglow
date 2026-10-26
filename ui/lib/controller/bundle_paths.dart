import 'dart:io';

import '../ui_error.dart';

/// Paths belong to the installed bundle, never to the caller's working directory.
class BundlePaths {
  BundlePaths({String? executable, Map<String, String>? environment}) {
    final root = File(executable ?? Platform.resolvedExecutable).parent;
    final env = environment ?? Platform.environment;
    final suffix = Platform.isWindows ? '.exe' : '';
    String binary(String variable, String name) {
      final override = env[variable];
      return override != null && override.isNotEmpty
          ? File(override).absolute.path
          : File.fromUri(root.uri.resolve('bin/$name$suffix')).path;
    }

    controller = binary('WIREGLOW_CONTROLLER', 'controller');
    relay = binary('WIREGLOW_RELAY', 'free-turn-proxy');
    mihomo = binary('WIREGLOW_MIHOMO', 'mihomo');
  }

  late final String controller;
  late final String relay;
  late final String mihomo;

  Future<void> validate() async {
    for (final path in [controller, relay, mihomo]) {
      if (!await File(path).exists()) {
        throw const UiError(UiErrorCode.incompleteBundle);
      }
    }
  }
}
