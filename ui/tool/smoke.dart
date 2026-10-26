import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:grpc/grpc.dart';

import 'package:wireglow_ui/controller/controller_repository.dart';

// Настоящий Dart↔Go gRPC/UDS + systemd; только временные данные и фиктивные ядра.
Future<void> main() async {
  final repository = ControllerRepository();
  try {
    await repository.start();
    if ((await repository.snapshot()).state != State.STATE_IDLE) {
      throw StateError('Expected Idle');
    }
    if ((await repository.settings()).preferredMode !=
        ProxyMode.PROXY_MODE_LOCAL) {
      throw StateError('Expected local proxy by default');
    }
    await repository.capabilities();
    try {
      await repository.switchMode(ProxyMode.PROXY_MODE_LOCAL);
      throw StateError('SwitchProxyMode accepted an idle session');
    } on GrpcError catch (e) {
      if (e.code != StatusCode.failedPrecondition) rethrow;
    }
    await repository.setMode(ProxyMode.PROXY_MODE_SYSTEM);
    if ((await repository.settings()).preferredMode !=
        ProxyMode.PROXY_MODE_SYSTEM) {
      throw StateError('Mode preference was not saved');
    }
    final events = <State>[];
    final subscription = repository.watch().listen((s) => events.add(s.state));
    final link =
        'freeturn://${base64Url.encode(utf8.encode(jsonEncode({'v': 1, 'name': 'Interop smoke', 'provider': 'direct', 'peer': '127.0.0.1:56000', 'wg': '[Interface]\nPrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\nAddress = 10.13.13.3/32\n[Peer]\nPublicKey = AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=\nAllowedIPs = 0.0.0.0/0\nEndpoint = 127.0.0.1:51820'}))).replaceAll('=', '')}';
    final imported = await repository.import(link);
    final profiles = await repository.profiles();
    if (profiles.length != 1 ||
        profiles.single.id != imported.id ||
        profiles.single.peer != '127.0.0.1:56000') {
      throw StateError('Import/summary mismatch');
    }
    await repository.connect(imported.id);
    await Future<void>.delayed(const Duration(milliseconds: 600));
    if ((await repository.snapshot()).activeMode !=
        ProxyMode.PROXY_MODE_LOCAL) {
      throw StateError(
        'Explicit session mode did not override saved preference',
      );
    }
    final stopped = await repository.disconnect();
    if (stopped.state != State.STATE_IDLE) {
      throw StateError('Cleanup did not reach Idle');
    }
    await subscription.cancel();
    if (!events.contains(State.STATE_IDLE) || events.length < 2) {
      throw StateError('Snapshot stream did not report startup');
    }
    stdout.writeln(
      'PASS: Hello, proxy settings, capabilities, explicit local session mode, import, stream, cancellation, cleanup',
    );
  } finally {
    await repository.close();
  }
  final restarted = ControllerRepository();
  try {
    await restarted.start();
    if ((await restarted.settings()).preferredMode !=
        ProxyMode.PROXY_MODE_SYSTEM) {
      throw StateError('Proxy preference did not survive controller restart');
    }
    await restarted.setMode(ProxyMode.PROXY_MODE_LOCAL);
    stdout.writeln('PASS: proxy preference survives controller restart');
  } finally {
    await restarted.close();
  }
}
