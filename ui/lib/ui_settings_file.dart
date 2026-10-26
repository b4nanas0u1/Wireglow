import 'dart:io';

import 'storage_paths.dart';

/// UI and backend share the selected data directory, including portable mode.
File? uiSettingsFile(
  String name, {
  Directory? configurationDirectory,
  StoragePaths? storagePaths,
}) {
  if (configurationDirectory == null) {
    final paths = storagePaths ?? StoragePaths();
    if (paths.portable) return File('${paths.data}/$name');
  }
  final env = Platform.environment;
  final home = env[Platform.isWindows ? 'USERPROFILE' : 'HOME'];
  final base =
      configurationDirectory?.path ??
      (Platform.isWindows
          ? env['APPDATA']
          : Platform.isMacOS
          ? (home == null ? null : '$home/Library/Application Support')
          : env['XDG_CONFIG_HOME'] ?? (home == null ? null : '$home/.config'));
  return base == null ? null : File('$base/wireglow/$name');
}
