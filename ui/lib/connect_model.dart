import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:grpc/grpc.dart';

import 'controller/controller_repository.dart';
import 'ui_error.dart';

class ConnectModel extends ChangeNotifier {
  ConnectModel(this.repository);
  final ControllerRepository repository;
  List<ProfileSummary> profiles = [];
  ProfileSummary? selected;
  Snapshot snapshot = Snapshot(state: State.STATE_IDLE);
  bool available = false;
  bool loading = true;
  bool commandPending = false;
  bool disposed = false;
  bool latencyPending = false;
  ProxyMode preferredMode = ProxyMode.PROXY_MODE_LOCAL;
  String proxyAddress = '127.0.0.1:17891';
  String relayAddress = '127.0.0.1:17890';
  int apiPort = 17892;
  String systemProxyUnavailableReason = '';
  bool switchingMode = false;
  bool get modeAvailable =>
      preferredMode != ProxyMode.PROXY_MODE_SYSTEM ||
      systemProxyUnavailableReason.isEmpty;

  Future<void> selectMode(ProxyMode mode) async {
    if (!available ||
        commandPending ||
        connecting ||
        stopping ||
        (mode == preferredMode &&
            (!connected || snapshot.activeMode == mode))) {
      return;
    }
    if (mode == ProxyMode.PROXY_MODE_SYSTEM &&
        systemProxyUnavailableReason.isNotEmpty) {
      error = systemProxyUnavailableReason;
      _notify();
      return;
    }
    final applyToSession = connected;
    commandPending = true;
    switchingMode = true;
    error = null;
    _notify();
    try {
      preferredMode = (await repository.setMode(mode)).preferredMode;
      if (applyToSession) {
        _accept(await repository.switchMode(preferredMode));
      }
    } catch (e) {
      error = message(e);
      try {
        _accept(await repository.snapshot());
      } catch (_) {
        _lost();
      }
    } finally {
      switchingMode = false;
      commandPending = false;
      _notify();
    }
  }

  void _acceptSettings(ClientSettings value) {
    preferredMode = value.preferredMode;
    apiPort = value.apiPort == 0 ? 17892 : value.apiPort;
    proxyAddress = value.proxyAddress.isEmpty
        ? '127.0.0.1:17891'
        : value.proxyAddress;
    relayAddress = value.relayAddress.isEmpty
        ? '127.0.0.1:17890'
        : value.relayAddress;
  }

  Future<void> saveAddresses(String proxy, String relay, int api) async {
    if (!available || commandPending || connecting || stopping) {
      throw const UiError(UiErrorCode.waitController);
    }
    commandPending = true;
    _notify();
    try {
      _acceptSettings(
        await repository.setSettings(
          ClientSettings(
            preferredMode: preferredMode,
            proxyAddress: proxy,
            relayAddress: relay,
            apiPort: api,
          ),
        ),
      );
    } finally {
      commandPending = false;
      _notify();
    }
  }

  Future<void> resetSettings() async {
    if (!available || commandPending || connecting || stopping) {
      throw const UiError(UiErrorCode.waitController);
    }
    commandPending = true;
    switchingMode = connected;
    _notify();
    try {
      _acceptSettings(
        await repository.setSettings(
          ClientSettings(
            preferredMode: ProxyMode.PROXY_MODE_LOCAL,
            proxyAddress: '127.0.0.1:17891',
            relayAddress: '127.0.0.1:17890',
            apiPort: 17892,
          ),
        ),
      );
      if (connected && snapshot.activeMode != preferredMode) {
        _accept(await repository.switchMode(preferredMode));
      }
    } catch (_) {
      try {
        _accept(await repository.snapshot());
      } catch (_) {
        _lost();
      }
      rethrow;
    } finally {
      switchingMode = false;
      commandPending = false;
      _notify();
    }
  }

  Future<void> checkLatency() async {
    if (!connected || latencyPending || snapshot.latency.checking) return;
    latencyPending = true;
    _notify();
    try {
      _accept(await repository.checkLatency());
    } catch (e) {
      error = message(e);
    } finally {
      latencyPending = false;
      _notify();
    }
  }

  Future<void>? _closing;
  Object? error;
  StreamSubscription<Snapshot>? _events;
  StreamSubscription<void>? _death;
  Timer? _retry;
  bool get busy =>
      commandPending ||
      (snapshot.state != State.STATE_IDLE &&
          snapshot.state != State.STATE_FAILED);
  bool get stopping => snapshot.state == State.STATE_STOPPING;
  bool get connected => available && snapshot.state == State.STATE_CONNECTED;
  bool get connecting => available && busy && !connected && !stopping;

  Future<void> start() async {
    try {
      await repository.start();
      _death = repository.death.listen((_) => _lost());
      await _refresh();
      _subscribe();
    } catch (e) {
      error = message(e);
    } finally {
      loading = false;
      _notify();
    }
  }

  Future<void> _refresh() async {
    _acceptSettings(await repository.settings());
    systemProxyUnavailableReason =
        (await repository.capabilities()).systemProxyUnavailableReason;
    profiles = await repository.profiles();
    _accept(await repository.snapshot());
    selected =
        profiles.where((p) => p.id == selected?.id).firstOrNull ??
        profiles.firstOrNull;
    available = true;
    error = null;
    _notify();
  }

  void _accept(Snapshot value) {
    snapshot = value;
    if (value.profileId.isNotEmpty) {
      selected =
          profiles.where((p) => p.id == value.profileId).firstOrNull ??
          selected;
    }
  }

  void _subscribe() {
    _events = repository.watch().listen(
      (value) {
        _accept(value);
        available = true;
        _notify();
      },
      onError: (Object _) => _lost(),
      onDone: _lost,
    );
  }

  void _lost() {
    if (disposed) return;
    available = false;
    error = const UiError(UiErrorCode.lostController);
    _notify();
    _retry?.cancel();
    _retry = Timer(const Duration(seconds: 2), () async {
      if (disposed) return;
      try {
        await _events?.cancel();
        await _refresh();
        _subscribe();
      } catch (_) {
        _lost();
      }
    });
  }

  void select(ProfileSummary p) {
    if (busy || !available) return;
    selected = p;
    error = null;
    _notify();
  }

  Future<void> action() async {
    if (!available || stopping || commandPending || selected == null) return;
    final wasBusy = busySession;
    if (!wasBusy && !modeAvailable) return;
    commandPending = true;
    error = null;
    _notify();
    try {
      if (busySession) {
        _accept(await repository.disconnect());
      }
      if (!wasBusy) {
        await repository.connect(selected!.id, mode: preferredMode);
        // ACK не является признаком Connected.
        _accept(await repository.snapshot());
      }
    } catch (e) {
      error = message(e);
      try {
        _accept(await repository.snapshot());
      } catch (_) {
        _lost();
      }
    } finally {
      commandPending = false;
      _notify();
    }
  }

  bool get busySession =>
      snapshot.state != State.STATE_IDLE &&
      snapshot.state != State.STATE_FAILED;
  Future<void> add(String link) async {
    if (!available || busy) {
      throw const UiError(UiErrorCode.waitController);
    }
    commandPending = true;
    _notify();
    try {
      final imported = await repository.import(link.trim());
      await _refresh();
      selected = profiles.firstWhere((p) => p.id == imported.id);
    } finally {
      commandPending = false;
      _notify();
    }
  }

  Future<void> deleteProfile(String id) async {
    if (!available || busy || !profiles.any((p) => p.id == id)) return;
    commandPending = true;
    error = null;
    _notify();
    try {
      await repository.deleteProfile(id);
      profiles = profiles.where((p) => p.id != id).toList();
      if (selected?.id == id) selected = profiles.firstOrNull;
    } catch (e) {
      error = message(e);
    } finally {
      commandPending = false;
      _notify();
    }
  }

  Future<void> editProfile({
    required String id,
    required String name,
    required List<String> callLinks,
    required int streams,
    required int streamsPerCredential,
  }) async {
    if (!available || commandPending || connecting || stopping) {
      throw const UiError(UiErrorCode.waitController);
    }
    commandPending = true;
    _notify();
    try {
      await repository.updateProfile(
        id: id,
        name: name,
        callLinks: callLinks,
        streams: streams,
        streamsPerCredential: streamsPerCredential,
      );
      profiles = await repository.profiles();
      selected = profiles.where((p) => p.id == id).firstOrNull;
    } finally {
      commandPending = false;
      _notify();
    }
  }

  static Object message(Object e) {
    if (e is UiError) return e;
    if (e is GrpcError) return e.message ?? e.toString();
    if (e is StateError) return e.message.toString();
    return const UiError(UiErrorCode.startFailed);
  }

  void _notify() {
    if (!disposed) notifyListeners();
  }

  Future<void> close() => _closing ??= _close();
  Future<void> _close() async {
    disposed = true;
    _retry?.cancel();
    await _events?.cancel();
    await _death?.cancel();
    await repository.close();
  }

  @override
  void dispose() {
    unawaited(close());
    super.dispose();
  }
}
