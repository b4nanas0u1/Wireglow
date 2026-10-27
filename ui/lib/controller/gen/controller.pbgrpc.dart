// This is a generated file - do not edit.
//
// Generated from controller.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/empty.pb.dart' as $0;

import 'controller.pb.dart' as $1;

export 'controller.pb.dart';

/// Локальный controller; каждый RPC требует metadata x-wireglow-session-token.
@$pb.GrpcServiceName('wireglow.control.ControllerService')
class ControllerServiceClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  ControllerServiceClient(super.channel, {super.options, super.interceptors});

  $grpc.ResponseFuture<$1.ClientSettings> getSettings(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getSettings, request, options: options);
  }

  $grpc.ResponseFuture<$1.ClientSettings> setSettings(
    $1.ClientSettings request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$setSettings, request, options: options);
  }

  $grpc.ResponseFuture<$1.Capabilities> getCapabilities(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getCapabilities, request, options: options);
  }

  $grpc.ResponseFuture<$1.HelloResponse> hello(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$hello, request, options: options);
  }

  $grpc.ResponseFuture<$1.ListProfilesResponse> listProfiles(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$listProfiles, request, options: options);
  }

  $grpc.ResponseFuture<$1.ProfileSummary> previewImport(
    $1.ImportProfileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$previewImport, request, options: options);
  }

  $grpc.ResponseFuture<$1.ProfileSummary> importProfile(
    $1.ImportProfileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$importProfile, request, options: options);
  }

  /// Явное чтение для редактора: ответ содержит секреты, не предназначен для логов.
  $grpc.ResponseFuture<$1.Profile> getProfile(
    $1.ProfileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getProfile, request, options: options);
  }

  $grpc.ResponseFuture<$1.ProfileSummary> updateProfile(
    $1.ProfileEdits request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$updateProfile, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> exportProfiles(
    $1.ExportProfilesRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$exportProfiles, request, options: options);
  }

  $grpc.ResponseFuture<$1.ProfileFilePreview> previewProfileFile(
    $1.ProfileFileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$previewProfileFile, request, options: options);
  }

  $grpc.ResponseFuture<$1.ImportProfileFileResponse> importProfileFile(
    $1.ImportProfileFileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$importProfileFile, request, options: options);
  }

  $grpc.ResponseFuture<$0.Empty> deleteProfile(
    $1.ProfileRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$deleteProfile, request, options: options);
  }

  /// ACK означает принятие запуска; результат и прогресс читаются из snapshots.
  $grpc.ResponseFuture<$0.Empty> connect(
    $1.ConnectRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$connect, request, options: options);
  }

  /// Отменяет также незавершённый Connect; ждёт сетевого cleanup.
  $grpc.ResponseFuture<$1.Snapshot> disconnect(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$disconnect, request, options: options);
  }

  /// Меняет настройки ОС активной сессии без перезапуска ядер.
  $grpc.ResponseFuture<$1.Snapshot> switchProxyMode(
    $1.ProxyModeRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$switchProxyMode, request, options: options);
  }

  $grpc.ResponseFuture<$1.Snapshot> getSnapshot(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getSnapshot, request, options: options);
  }

  /// HTTPS latency through the active connection; failure does not disconnect.
  $grpc.ResponseFuture<$1.Snapshot> checkLatency(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$checkLatency, request, options: options);
  }

  /// Первое сообщение — текущее состояние. Медленные клиенты пропускают промежуточные состояния.
  /// Поток не является журналом событий и не продлевает lease владельца.
  $grpc.ResponseStream<$1.Snapshot> watchSnapshots(
    $0.Empty request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$watchSnapshots, $async.Stream.fromIterable([request]),
        options: options);
  }

  // method descriptors

  static final _$getSettings = $grpc.ClientMethod<$0.Empty, $1.ClientSettings>(
      '/wireglow.control.ControllerService/GetSettings',
      ($0.Empty value) => value.writeToBuffer(),
      $1.ClientSettings.fromBuffer);
  static final _$setSettings =
      $grpc.ClientMethod<$1.ClientSettings, $1.ClientSettings>(
          '/wireglow.control.ControllerService/SetSettings',
          ($1.ClientSettings value) => value.writeToBuffer(),
          $1.ClientSettings.fromBuffer);
  static final _$getCapabilities =
      $grpc.ClientMethod<$0.Empty, $1.Capabilities>(
          '/wireglow.control.ControllerService/GetCapabilities',
          ($0.Empty value) => value.writeToBuffer(),
          $1.Capabilities.fromBuffer);
  static final _$hello = $grpc.ClientMethod<$0.Empty, $1.HelloResponse>(
      '/wireglow.control.ControllerService/Hello',
      ($0.Empty value) => value.writeToBuffer(),
      $1.HelloResponse.fromBuffer);
  static final _$listProfiles =
      $grpc.ClientMethod<$0.Empty, $1.ListProfilesResponse>(
          '/wireglow.control.ControllerService/ListProfiles',
          ($0.Empty value) => value.writeToBuffer(),
          $1.ListProfilesResponse.fromBuffer);
  static final _$previewImport =
      $grpc.ClientMethod<$1.ImportProfileRequest, $1.ProfileSummary>(
          '/wireglow.control.ControllerService/PreviewImport',
          ($1.ImportProfileRequest value) => value.writeToBuffer(),
          $1.ProfileSummary.fromBuffer);
  static final _$importProfile =
      $grpc.ClientMethod<$1.ImportProfileRequest, $1.ProfileSummary>(
          '/wireglow.control.ControllerService/ImportProfile',
          ($1.ImportProfileRequest value) => value.writeToBuffer(),
          $1.ProfileSummary.fromBuffer);
  static final _$getProfile = $grpc.ClientMethod<$1.ProfileRequest, $1.Profile>(
      '/wireglow.control.ControllerService/GetProfile',
      ($1.ProfileRequest value) => value.writeToBuffer(),
      $1.Profile.fromBuffer);
  static final _$updateProfile =
      $grpc.ClientMethod<$1.ProfileEdits, $1.ProfileSummary>(
          '/wireglow.control.ControllerService/UpdateProfile',
          ($1.ProfileEdits value) => value.writeToBuffer(),
          $1.ProfileSummary.fromBuffer);
  static final _$exportProfiles =
      $grpc.ClientMethod<$1.ExportProfilesRequest, $0.Empty>(
          '/wireglow.control.ControllerService/ExportProfiles',
          ($1.ExportProfilesRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$previewProfileFile =
      $grpc.ClientMethod<$1.ProfileFileRequest, $1.ProfileFilePreview>(
          '/wireglow.control.ControllerService/PreviewProfileFile',
          ($1.ProfileFileRequest value) => value.writeToBuffer(),
          $1.ProfileFilePreview.fromBuffer);
  static final _$importProfileFile = $grpc.ClientMethod<
          $1.ImportProfileFileRequest, $1.ImportProfileFileResponse>(
      '/wireglow.control.ControllerService/ImportProfileFile',
      ($1.ImportProfileFileRequest value) => value.writeToBuffer(),
      $1.ImportProfileFileResponse.fromBuffer);
  static final _$deleteProfile =
      $grpc.ClientMethod<$1.ProfileRequest, $0.Empty>(
          '/wireglow.control.ControllerService/DeleteProfile',
          ($1.ProfileRequest value) => value.writeToBuffer(),
          $0.Empty.fromBuffer);
  static final _$connect = $grpc.ClientMethod<$1.ConnectRequest, $0.Empty>(
      '/wireglow.control.ControllerService/Connect',
      ($1.ConnectRequest value) => value.writeToBuffer(),
      $0.Empty.fromBuffer);
  static final _$disconnect = $grpc.ClientMethod<$0.Empty, $1.Snapshot>(
      '/wireglow.control.ControllerService/Disconnect',
      ($0.Empty value) => value.writeToBuffer(),
      $1.Snapshot.fromBuffer);
  static final _$switchProxyMode =
      $grpc.ClientMethod<$1.ProxyModeRequest, $1.Snapshot>(
          '/wireglow.control.ControllerService/SwitchProxyMode',
          ($1.ProxyModeRequest value) => value.writeToBuffer(),
          $1.Snapshot.fromBuffer);
  static final _$getSnapshot = $grpc.ClientMethod<$0.Empty, $1.Snapshot>(
      '/wireglow.control.ControllerService/GetSnapshot',
      ($0.Empty value) => value.writeToBuffer(),
      $1.Snapshot.fromBuffer);
  static final _$checkLatency = $grpc.ClientMethod<$0.Empty, $1.Snapshot>(
      '/wireglow.control.ControllerService/CheckLatency',
      ($0.Empty value) => value.writeToBuffer(),
      $1.Snapshot.fromBuffer);
  static final _$watchSnapshots = $grpc.ClientMethod<$0.Empty, $1.Snapshot>(
      '/wireglow.control.ControllerService/WatchSnapshots',
      ($0.Empty value) => value.writeToBuffer(),
      $1.Snapshot.fromBuffer);
}

@$pb.GrpcServiceName('wireglow.control.ControllerService')
abstract class ControllerServiceBase extends $grpc.Service {
  $core.String get $name => 'wireglow.control.ControllerService';

  ControllerServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.ClientSettings>(
        'GetSettings',
        getSettings_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.ClientSettings value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ClientSettings, $1.ClientSettings>(
        'SetSettings',
        setSettings_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ClientSettings.fromBuffer(value),
        ($1.ClientSettings value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.Capabilities>(
        'GetCapabilities',
        getCapabilities_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.Capabilities value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.HelloResponse>(
        'Hello',
        hello_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.HelloResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.ListProfilesResponse>(
        'ListProfiles',
        listProfiles_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.ListProfilesResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ImportProfileRequest, $1.ProfileSummary>(
        'PreviewImport',
        previewImport_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.ImportProfileRequest.fromBuffer(value),
        ($1.ProfileSummary value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ImportProfileRequest, $1.ProfileSummary>(
        'ImportProfile',
        importProfile_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.ImportProfileRequest.fromBuffer(value),
        ($1.ProfileSummary value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ProfileRequest, $1.Profile>(
        'GetProfile',
        getProfile_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ProfileRequest.fromBuffer(value),
        ($1.Profile value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ProfileEdits, $1.ProfileSummary>(
        'UpdateProfile',
        updateProfile_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ProfileEdits.fromBuffer(value),
        ($1.ProfileSummary value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ExportProfilesRequest, $0.Empty>(
        'ExportProfiles',
        exportProfiles_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.ExportProfilesRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$1.ProfileFileRequest, $1.ProfileFilePreview>(
            'PreviewProfileFile',
            previewProfileFile_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $1.ProfileFileRequest.fromBuffer(value),
            ($1.ProfileFilePreview value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ImportProfileFileRequest,
            $1.ImportProfileFileResponse>(
        'ImportProfileFile',
        importProfileFile_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $1.ImportProfileFileRequest.fromBuffer(value),
        ($1.ImportProfileFileResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ProfileRequest, $0.Empty>(
        'DeleteProfile',
        deleteProfile_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ProfileRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ConnectRequest, $0.Empty>(
        'Connect',
        connect_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ConnectRequest.fromBuffer(value),
        ($0.Empty value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.Snapshot>(
        'Disconnect',
        disconnect_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.Snapshot value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$1.ProxyModeRequest, $1.Snapshot>(
        'SwitchProxyMode',
        switchProxyMode_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $1.ProxyModeRequest.fromBuffer(value),
        ($1.Snapshot value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.Snapshot>(
        'GetSnapshot',
        getSnapshot_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.Snapshot value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.Snapshot>(
        'CheckLatency',
        checkLatency_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.Snapshot value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.Empty, $1.Snapshot>(
        'WatchSnapshots',
        watchSnapshots_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.Empty.fromBuffer(value),
        ($1.Snapshot value) => value.writeToBuffer()));
  }

  $async.Future<$1.ClientSettings> getSettings_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getSettings($call, await $request);
  }

  $async.Future<$1.ClientSettings> getSettings(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.ClientSettings> setSettings_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ClientSettings> $request) async {
    return setSettings($call, await $request);
  }

  $async.Future<$1.ClientSettings> setSettings(
      $grpc.ServiceCall call, $1.ClientSettings request);

  $async.Future<$1.Capabilities> getCapabilities_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getCapabilities($call, await $request);
  }

  $async.Future<$1.Capabilities> getCapabilities(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.HelloResponse> hello_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return hello($call, await $request);
  }

  $async.Future<$1.HelloResponse> hello(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.ListProfilesResponse> listProfiles_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return listProfiles($call, await $request);
  }

  $async.Future<$1.ListProfilesResponse> listProfiles(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.ProfileSummary> previewImport_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ImportProfileRequest> $request) async {
    return previewImport($call, await $request);
  }

  $async.Future<$1.ProfileSummary> previewImport(
      $grpc.ServiceCall call, $1.ImportProfileRequest request);

  $async.Future<$1.ProfileSummary> importProfile_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ImportProfileRequest> $request) async {
    return importProfile($call, await $request);
  }

  $async.Future<$1.ProfileSummary> importProfile(
      $grpc.ServiceCall call, $1.ImportProfileRequest request);

  $async.Future<$1.Profile> getProfile_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ProfileRequest> $request) async {
    return getProfile($call, await $request);
  }

  $async.Future<$1.Profile> getProfile(
      $grpc.ServiceCall call, $1.ProfileRequest request);

  $async.Future<$1.ProfileSummary> updateProfile_Pre(
      $grpc.ServiceCall $call, $async.Future<$1.ProfileEdits> $request) async {
    return updateProfile($call, await $request);
  }

  $async.Future<$1.ProfileSummary> updateProfile(
      $grpc.ServiceCall call, $1.ProfileEdits request);

  $async.Future<$0.Empty> exportProfiles_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ExportProfilesRequest> $request) async {
    return exportProfiles($call, await $request);
  }

  $async.Future<$0.Empty> exportProfiles(
      $grpc.ServiceCall call, $1.ExportProfilesRequest request);

  $async.Future<$1.ProfileFilePreview> previewProfileFile_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$1.ProfileFileRequest> $request) async {
    return previewProfileFile($call, await $request);
  }

  $async.Future<$1.ProfileFilePreview> previewProfileFile(
      $grpc.ServiceCall call, $1.ProfileFileRequest request);

  $async.Future<$1.ImportProfileFileResponse> importProfileFile_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$1.ImportProfileFileRequest> $request) async {
    return importProfileFile($call, await $request);
  }

  $async.Future<$1.ImportProfileFileResponse> importProfileFile(
      $grpc.ServiceCall call, $1.ImportProfileFileRequest request);

  $async.Future<$0.Empty> deleteProfile_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ProfileRequest> $request) async {
    return deleteProfile($call, await $request);
  }

  $async.Future<$0.Empty> deleteProfile(
      $grpc.ServiceCall call, $1.ProfileRequest request);

  $async.Future<$0.Empty> connect_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ConnectRequest> $request) async {
    return connect($call, await $request);
  }

  $async.Future<$0.Empty> connect(
      $grpc.ServiceCall call, $1.ConnectRequest request);

  $async.Future<$1.Snapshot> disconnect_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return disconnect($call, await $request);
  }

  $async.Future<$1.Snapshot> disconnect(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.Snapshot> switchProxyMode_Pre($grpc.ServiceCall $call,
      $async.Future<$1.ProxyModeRequest> $request) async {
    return switchProxyMode($call, await $request);
  }

  $async.Future<$1.Snapshot> switchProxyMode(
      $grpc.ServiceCall call, $1.ProxyModeRequest request);

  $async.Future<$1.Snapshot> getSnapshot_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return getSnapshot($call, await $request);
  }

  $async.Future<$1.Snapshot> getSnapshot(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Future<$1.Snapshot> checkLatency_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async {
    return checkLatency($call, await $request);
  }

  $async.Future<$1.Snapshot> checkLatency(
      $grpc.ServiceCall call, $0.Empty request);

  $async.Stream<$1.Snapshot> watchSnapshots_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.Empty> $request) async* {
    yield* watchSnapshots($call, await $request);
  }

  $async.Stream<$1.Snapshot> watchSnapshots(
      $grpc.ServiceCall call, $0.Empty request);
}
