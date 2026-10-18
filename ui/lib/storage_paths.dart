import 'dart:io';

/// A marker beside the executable keeps every portable path inside the bundle.
class StoragePaths {
  StoragePaths({String? executable, Map<String, String>? environment}) {
    final root = File(executable ?? Platform.resolvedExecutable)
        .absolute
        .parent;
    final env = environment ?? Platform.environment;
    final marker = '${root.path}/portable.json';
    final type = FileSystemEntity.typeSync(marker);
    if (type != FileSystemEntityType.notFound &&
        type != FileSystemEntityType.file) {
      throw FileSystemException('portable.json must be a regular file', marker);
    }
    portable = type == FileSystemEntityType.file;
    data = portable ? '${root.path}/data' : _override(env, 'WIREGLOW_DATA_DIR');
    runtime = portable
        ? '${root.path}/data/runtime'
        : _override(env, 'WIREGLOW_RUNTIME_DIR');
  }

  static String? _override(Map<String, String> env, String key) {
    final value = env[key];
    return value == null || value.isEmpty ? null : File(value).absolute.path;
  }

  late final bool portable;
  late final String? data;
  late final String? runtime;

  List<String> get controllerArguments => [
    if (data != null) ...['--data-dir', data!],
    if (runtime != null) ...['--runtime-dir', runtime!],
  ];
}
