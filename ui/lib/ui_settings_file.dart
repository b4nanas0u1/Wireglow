import 'dart:io';

/// Resolves UI preferences independently of backend profiles and runtime data.
File? uiSettingsFile(String name, {Directory? configurationDirectory}) {
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
