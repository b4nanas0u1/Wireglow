import 'dart:async';
import 'dart:convert';
import 'dart:io';

import '../ui_error.dart';
import '../storage_paths.dart';
import 'bundle_paths.dart';

import 'package:grpc/grpc.dart';
import 'package:fixnum/fixnum.dart';

import 'gen/controller.pbgrpc.dart';

import 'package:protobuf/well_known_types/google/protobuf/empty.pb.dart';

export 'gen/controller.pb.dart';

/// Локальный launcher владеет lease; закрытие stdin завершает его сессию.
class ControllerRepository {
  Process? _owner;
  ClientChannel? _channel;
  ControllerServiceClient? _client;
  Future<void>? _closing;
  final _death = StreamController<void>.broadcast();
  Stream<void> get death => _death.stream;
  ControllerServiceClient get client => _client!;

  Future<void> start() async {
    final paths = BundlePaths();
    await paths.validate();
    final args = <String>[
      'desktop',
      '--relay',
      paths.relay,
      '--mihomo',
      paths.mihomo,
    ];
    args.addAll(StoragePaths().controllerArguments);
    final process = await Process.start(paths.controller, args);
    _owner = process;
    // Вывод ядер и backend не попадает в UI или его логи.
    unawaited(process.stderr.drain<void>());
    final endpoint = Completer<Map<String, dynamic>>();
    process.stdout
        .transform(utf8.decoder)
        .transform(const LineSplitter())
        .listen((line) {
          if (endpoint.isCompleted) return;
          try {
            final data = jsonDecode(line) as Map<String, dynamic>;
            if (data['socket'] is String && data['token'] is String) {
              endpoint.complete(data);
            }
          } catch (_) {
            /* stdout не является RPC. */
          }
        });
    unawaited(
      process.exitCode.then((_) {
        if (!endpoint.isCompleted) {
          endpoint.completeError(const UiError(UiErrorCode.controllerFailed));
        }
        if (!_death.isClosed) _death.add(null);
      }),
    );
    try {
      final data = await endpoint.future.timeout(const Duration(seconds: 35));
      _channel = ClientChannel(
        InternetAddress(
          data['socket'] as String,
          type: InternetAddressType.unix,
        ),
        port: 0,
        options: const ChannelOptions(
          credentials: ChannelCredentials.insecure(),
        ),
      );
      _client = ControllerServiceClient(
        _channel!,
        options: CallOptions(
          metadata: {'x-wireglow-session-token': data['token'] as String},
        ),
      );
      await client.hello(
        Empty(),
        options: CallOptions(timeout: const Duration(seconds: 5)),
      );
    } catch (_) {
      await close();
      rethrow;
    }
  }

  Future<ClientSettings> settings() =>
      client.getSettings(Empty(), options: _deadline());
  Future<ClientSettings> setSettings(ClientSettings settings) =>
      client.setSettings(settings, options: _deadline());
  Future<ClientSettings> setMode(ProxyMode mode) async {
    final current = await settings();
    current.preferredMode = mode;
    return setSettings(current);
  }

  Future<Snapshot> switchMode(ProxyMode mode) => client.switchProxyMode(
    ProxyModeRequest(mode: mode),
    options: CallOptions(timeout: const Duration(seconds: 45)),
  );
  Future<Capabilities> capabilities() =>
      client.getCapabilities(Empty(), options: _deadline());

  Future<List<ProfileSummary>> profiles() async =>
      (await client.listProfiles(Empty(), options: _deadline())).profiles;
  Future<Profile> getProfile(String id) =>
      client.getProfile(ProfileRequest(id: id), options: _deadline());
  Future<Snapshot> snapshot() =>
      client.getSnapshot(Empty(), options: _deadline());
  Future<Snapshot> checkLatency() => client.checkLatency(
    Empty(),
    options: CallOptions(timeout: const Duration(seconds: 20)),
  );
  Stream<Snapshot> watch() => client.watchSnapshots(Empty());
  Future<ProfileSummary> import(String link) => client.importProfile(
    ImportProfileRequest(freeturnLink: link),
    options: _deadline(),
  );
  Future<void> exportProfiles(List<String> ids, String path) async {
    await client.exportProfiles(
      ExportProfilesRequest(ids: ids, path: path),
      options: CallOptions(timeout: const Duration(seconds: 60)),
    );
  }

  Future<ProfileFilePreview> previewProfileFile(String path) =>
      client.previewProfileFile(
        ProfileFileRequest(path: path),
        options: CallOptions(timeout: const Duration(seconds: 60)),
      );
  Future<ImportProfileFileResponse> importProfileFile(
    String token,
    List<ProfileImportChoice> choices,
  ) => client.importProfileFile(
    ImportProfileFileRequest(token: token, choices: choices),
    options: CallOptions(timeout: const Duration(seconds: 60)),
  );

  Future<void> connect(
    String id, {
    ProxyMode mode = ProxyMode.PROXY_MODE_LOCAL,
  }) async {
    await client.connect(
      ConnectRequest(id: id, mode: mode),
      options: _deadline(),
    );
  }

  Future<void> deleteProfile(String id) async {
    await client.deleteProfile(ProfileRequest(id: id), options: _deadline());
  }

  Future<ProfileSummary> updateProfile({
    required String id,
    required String name,
    required List<String> callLinks,
    required int streams,
    required int streamsPerCredential,
    ProfileAdvancedEdits? advanced,
  }) => client.updateProfile(
    ProfileEdits(
      id: id,
      name: name,
      callLinks: callLinks,
      streams: Int64(streams),
      streamsPerCredential: Int64(streamsPerCredential),
      advanced: advanced,
    ),
    options: _deadline(),
  );

  Future<Snapshot> disconnect() => client.disconnect(
    Empty(),
    options: CallOptions(timeout: const Duration(seconds: 45)),
  );
  CallOptions _deadline() => CallOptions(timeout: const Duration(seconds: 10));

  Future<void> close() => _closing ??= _close();
  Future<void> _close() async {
    await _channel?.terminate();
    await _owner?.stdin.close();
    if (_owner != null) {
      try {
        await _owner!.exitCode.timeout(const Duration(seconds: 65));
      } on TimeoutException {
        // Recovery остаётся самостоятельной службой при аварии владельца.
        _owner!.kill(ProcessSignal.sigterm);
      }
    }
    await _death.close();
  }
}
