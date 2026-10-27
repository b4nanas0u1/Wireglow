// This is a generated file - do not edit.
//
// Generated from controller.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'controller.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'controller.pbenum.dart';

class HelloResponse extends $pb.GeneratedMessage {
  factory HelloResponse({
    $core.String? buildId,
    $core.String? epoch,
  }) {
    final result = HelloResponse._();
    if (buildId != null) result.buildId = buildId;
    if (epoch != null) result.epoch = epoch;
    return result;
  }

  HelloResponse._();

  factory HelloResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HelloResponse()..mergeFromBuffer(data, registry);
  factory HelloResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HelloResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HelloResponse',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: HelloResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'buildId')
    ..aOS(2, _omitFieldNames ? '' : 'epoch')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HelloResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HelloResponse copyWith(void Function(HelloResponse) updates) =>
      super.copyWith((message) => updates(message as HelloResponse))
          as HelloResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HelloResponse() / HelloResponse.new instead')
  static HelloResponse create() => HelloResponse._();
  static $pb.GeneratedMessage $_createMessage() => HelloResponse._();
  @$core.override
  HelloResponse createEmptyInstance() => HelloResponse._();
  @$core.pragma('dart2js:noInline')
  static HelloResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HelloResponse>(
          HelloResponse.$_createMessage);
  static HelloResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get buildId => $_getSZ(0);
  @$pb.TagNumber(1)
  set buildId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBuildId() => $_has(0);
  @$pb.TagNumber(1)
  void clearBuildId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get epoch => $_getSZ(1);
  @$pb.TagNumber(2)
  set epoch($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEpoch() => $_has(1);
  @$pb.TagNumber(2)
  void clearEpoch() => $_clearField(2);
}

class ConnectRequest extends $pb.GeneratedMessage {
  factory ConnectRequest({
    $core.String? id,
    ProxyMode? mode,
  }) {
    final result = ConnectRequest._();
    if (id != null) result.id = id;
    if (mode != null) result.mode = mode;
    return result;
  }

  ConnectRequest._();

  factory ConnectRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConnectRequest()..mergeFromBuffer(data, registry);
  factory ConnectRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConnectRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConnectRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ConnectRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aE<ProxyMode>(2, _omitFieldNames ? '' : 'mode',
        enumValues: ProxyMode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConnectRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConnectRequest copyWith(void Function(ConnectRequest) updates) =>
      super.copyWith((message) => updates(message as ConnectRequest))
          as ConnectRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ConnectRequest() / ConnectRequest.new instead')
  static ConnectRequest create() => ConnectRequest._();
  static $pb.GeneratedMessage $_createMessage() => ConnectRequest._();
  @$core.override
  ConnectRequest createEmptyInstance() => ConnectRequest._();
  @$core.pragma('dart2js:noInline')
  static ConnectRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ConnectRequest>(
          ConnectRequest.$_createMessage);
  static ConnectRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  ProxyMode get mode => $_getN(1);
  @$pb.TagNumber(2)
  set mode(ProxyMode value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasMode() => $_has(1);
  @$pb.TagNumber(2)
  void clearMode() => $_clearField(2);
}

class ProxyModeRequest extends $pb.GeneratedMessage {
  factory ProxyModeRequest({
    ProxyMode? mode,
  }) {
    final result = ProxyModeRequest._();
    if (mode != null) result.mode = mode;
    return result;
  }

  ProxyModeRequest._();

  factory ProxyModeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProxyModeRequest()..mergeFromBuffer(data, registry);
  factory ProxyModeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProxyModeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProxyModeRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProxyModeRequest.$_createMessage)
    ..aE<ProxyMode>(1, _omitFieldNames ? '' : 'mode',
        enumValues: ProxyMode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProxyModeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProxyModeRequest copyWith(void Function(ProxyModeRequest) updates) =>
      super.copyWith((message) => updates(message as ProxyModeRequest))
          as ProxyModeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProxyModeRequest() / ProxyModeRequest.new instead')
  static ProxyModeRequest create() => ProxyModeRequest._();
  static $pb.GeneratedMessage $_createMessage() => ProxyModeRequest._();
  @$core.override
  ProxyModeRequest createEmptyInstance() => ProxyModeRequest._();
  @$core.pragma('dart2js:noInline')
  static ProxyModeRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProxyModeRequest>(
          ProxyModeRequest.$_createMessage);
  static ProxyModeRequest? _defaultInstance;

  @$pb.TagNumber(1)
  ProxyMode get mode => $_getN(0);
  @$pb.TagNumber(1)
  set mode(ProxyMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearMode() => $_clearField(1);
}

class ClientSettings extends $pb.GeneratedMessage {
  factory ClientSettings({
    ProxyMode? preferredMode,
    $core.String? proxyAddress,
    $core.String? relayAddress,
    $core.int? apiPort,
  }) {
    final result = ClientSettings._();
    if (preferredMode != null) result.preferredMode = preferredMode;
    if (proxyAddress != null) result.proxyAddress = proxyAddress;
    if (relayAddress != null) result.relayAddress = relayAddress;
    if (apiPort != null) result.apiPort = apiPort;
    return result;
  }

  ClientSettings._();

  factory ClientSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ClientSettings()..mergeFromBuffer(data, registry);
  factory ClientSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ClientSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ClientSettings',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ClientSettings.$_createMessage)
    ..aE<ProxyMode>(1, _omitFieldNames ? '' : 'preferredMode',
        enumValues: ProxyMode.values)
    ..aOS(2, _omitFieldNames ? '' : 'proxyAddress')
    ..aOS(3, _omitFieldNames ? '' : 'relayAddress')
    ..aI(4, _omitFieldNames ? '' : 'apiPort', fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ClientSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ClientSettings copyWith(void Function(ClientSettings) updates) =>
      super.copyWith((message) => updates(message as ClientSettings))
          as ClientSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ClientSettings() / ClientSettings.new instead')
  static ClientSettings create() => ClientSettings._();
  static $pb.GeneratedMessage $_createMessage() => ClientSettings._();
  @$core.override
  ClientSettings createEmptyInstance() => ClientSettings._();
  @$core.pragma('dart2js:noInline')
  static ClientSettings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ClientSettings>(
          ClientSettings.$_createMessage);
  static ClientSettings? _defaultInstance;

  @$pb.TagNumber(1)
  ProxyMode get preferredMode => $_getN(0);
  @$pb.TagNumber(1)
  set preferredMode(ProxyMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasPreferredMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearPreferredMode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get proxyAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set proxyAddress($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProxyAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearProxyAddress() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get relayAddress => $_getSZ(2);
  @$pb.TagNumber(3)
  set relayAddress($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRelayAddress() => $_has(2);
  @$pb.TagNumber(3)
  void clearRelayAddress() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get apiPort => $_getIZ(3);
  @$pb.TagNumber(4)
  set apiPort($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasApiPort() => $_has(3);
  @$pb.TagNumber(4)
  void clearApiPort() => $_clearField(4);
}

class Capabilities extends $pb.GeneratedMessage {
  factory Capabilities({
    $core.String? systemProxyUnavailableReason,
    $core.String? tunUnavailableReason,
  }) {
    final result = Capabilities._();
    if (systemProxyUnavailableReason != null)
      result.systemProxyUnavailableReason = systemProxyUnavailableReason;
    if (tunUnavailableReason != null)
      result.tunUnavailableReason = tunUnavailableReason;
    return result;
  }

  Capabilities._();

  factory Capabilities.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Capabilities()..mergeFromBuffer(data, registry);
  factory Capabilities.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Capabilities()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Capabilities',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: Capabilities.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'systemProxyUnavailableReason')
    ..aOS(2, _omitFieldNames ? '' : 'tunUnavailableReason')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Capabilities clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Capabilities copyWith(void Function(Capabilities) updates) =>
      super.copyWith((message) => updates(message as Capabilities))
          as Capabilities;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Capabilities() / Capabilities.new instead')
  static Capabilities create() => Capabilities._();
  static $pb.GeneratedMessage $_createMessage() => Capabilities._();
  @$core.override
  Capabilities createEmptyInstance() => Capabilities._();
  @$core.pragma('dart2js:noInline')
  static Capabilities getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Capabilities>(
          Capabilities.$_createMessage);
  static Capabilities? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get systemProxyUnavailableReason => $_getSZ(0);
  @$pb.TagNumber(1)
  set systemProxyUnavailableReason($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSystemProxyUnavailableReason() => $_has(0);
  @$pb.TagNumber(1)
  void clearSystemProxyUnavailableReason() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get tunUnavailableReason => $_getSZ(1);
  @$pb.TagNumber(2)
  set tunUnavailableReason($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTunUnavailableReason() => $_has(1);
  @$pb.TagNumber(2)
  void clearTunUnavailableReason() => $_clearField(2);
}

class ProfileRequest extends $pb.GeneratedMessage {
  factory ProfileRequest({
    $core.String? id,
  }) {
    final result = ProfileRequest._();
    if (id != null) result.id = id;
    return result;
  }

  ProfileRequest._();

  factory ProfileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileRequest()..mergeFromBuffer(data, registry);
  factory ProfileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileRequest copyWith(void Function(ProfileRequest) updates) =>
      super.copyWith((message) => updates(message as ProfileRequest))
          as ProfileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileRequest() / ProfileRequest.new instead')
  static ProfileRequest create() => ProfileRequest._();
  static $pb.GeneratedMessage $_createMessage() => ProfileRequest._();
  @$core.override
  ProfileRequest createEmptyInstance() => ProfileRequest._();
  @$core.pragma('dart2js:noInline')
  static ProfileRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProfileRequest>(
          ProfileRequest.$_createMessage);
  static ProfileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

/// Editable client-side settings; server keys and WG configuration are preserved.
class ProfileEdits extends $pb.GeneratedMessage {
  factory ProfileEdits({
    $core.String? id,
    $core.String? name,
    $core.Iterable<$core.String>? callLinks,
    $fixnum.Int64? streams,
    $fixnum.Int64? streamsPerCredential,
    ProfileAdvancedEdits? advanced,
  }) {
    final result = ProfileEdits._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (callLinks != null) result.callLinks.addAll(callLinks);
    if (streams != null) result.streams = streams;
    if (streamsPerCredential != null)
      result.streamsPerCredential = streamsPerCredential;
    if (advanced != null) result.advanced = advanced;
    return result;
  }

  ProfileEdits._();

  factory ProfileEdits.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileEdits()..mergeFromBuffer(data, registry);
  factory ProfileEdits.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileEdits()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileEdits',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileEdits.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..pPS(3, _omitFieldNames ? '' : 'callLinks')
    ..aInt64(4, _omitFieldNames ? '' : 'streams')
    ..aInt64(5, _omitFieldNames ? '' : 'streamsPerCredential')
    ..aOM<ProfileAdvancedEdits>(6, _omitFieldNames ? '' : 'advanced',
        subBuilder: ProfileAdvancedEdits.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileEdits clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileEdits copyWith(void Function(ProfileEdits) updates) =>
      super.copyWith((message) => updates(message as ProfileEdits))
          as ProfileEdits;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileEdits() / ProfileEdits.new instead')
  static ProfileEdits create() => ProfileEdits._();
  static $pb.GeneratedMessage $_createMessage() => ProfileEdits._();
  @$core.override
  ProfileEdits createEmptyInstance() => ProfileEdits._();
  @$core.pragma('dart2js:noInline')
  static ProfileEdits getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProfileEdits>(
          ProfileEdits.$_createMessage);
  static ProfileEdits? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get callLinks => $_getList(2);

  @$pb.TagNumber(4)
  $fixnum.Int64 get streams => $_getI64(3);
  @$pb.TagNumber(4)
  set streams($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasStreams() => $_has(3);
  @$pb.TagNumber(4)
  void clearStreams() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get streamsPerCredential => $_getI64(4);
  @$pb.TagNumber(5)
  set streamsPerCredential($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasStreamsPerCredential() => $_has(4);
  @$pb.TagNumber(5)
  void clearStreamsPerCredential() => $_clearField(5);

  @$pb.TagNumber(6)
  ProfileAdvancedEdits get advanced => $_getN(5);
  @$pb.TagNumber(6)
  set advanced(ProfileAdvancedEdits value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasAdvanced() => $_has(5);
  @$pb.TagNumber(6)
  void clearAdvanced() => $_clearField(6);
  @$pb.TagNumber(6)
  ProfileAdvancedEdits ensureAdvanced() => $_ensure(5);
}

class ProfileAdvancedEdits extends $pb.GeneratedMessage {
  factory ProfileAdvancedEdits({
    $core.String? turnTransport,
    DNSSettings? dns,
    $core.bool? manualCaptcha,
    $core.String? vkPlatform,
    $core.String? turnAddress,
    $fixnum.Int64? pacingMs,
  }) {
    final result = ProfileAdvancedEdits._();
    if (turnTransport != null) result.turnTransport = turnTransport;
    if (dns != null) result.dns = dns;
    if (manualCaptcha != null) result.manualCaptcha = manualCaptcha;
    if (vkPlatform != null) result.vkPlatform = vkPlatform;
    if (turnAddress != null) result.turnAddress = turnAddress;
    if (pacingMs != null) result.pacingMs = pacingMs;
    return result;
  }

  ProfileAdvancedEdits._();

  factory ProfileAdvancedEdits.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileAdvancedEdits()..mergeFromBuffer(data, registry);
  factory ProfileAdvancedEdits.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileAdvancedEdits()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileAdvancedEdits',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileAdvancedEdits.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'turnTransport')
    ..aOM<DNSSettings>(2, _omitFieldNames ? '' : 'dns',
        subBuilder: DNSSettings.$_createMessage)
    ..aOB(3, _omitFieldNames ? '' : 'manualCaptcha')
    ..aOS(4, _omitFieldNames ? '' : 'vkPlatform')
    ..aOS(5, _omitFieldNames ? '' : 'turnAddress')
    ..aInt64(6, _omitFieldNames ? '' : 'pacingMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileAdvancedEdits clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileAdvancedEdits copyWith(void Function(ProfileAdvancedEdits) updates) =>
      super.copyWith((message) => updates(message as ProfileAdvancedEdits))
          as ProfileAdvancedEdits;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ProfileAdvancedEdits() / ProfileAdvancedEdits.new instead')
  static ProfileAdvancedEdits create() => ProfileAdvancedEdits._();
  static $pb.GeneratedMessage $_createMessage() => ProfileAdvancedEdits._();
  @$core.override
  ProfileAdvancedEdits createEmptyInstance() => ProfileAdvancedEdits._();
  @$core.pragma('dart2js:noInline')
  static ProfileAdvancedEdits getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProfileAdvancedEdits>(
          ProfileAdvancedEdits.$_createMessage);
  static ProfileAdvancedEdits? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get turnTransport => $_getSZ(0);
  @$pb.TagNumber(1)
  set turnTransport($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTurnTransport() => $_has(0);
  @$pb.TagNumber(1)
  void clearTurnTransport() => $_clearField(1);

  @$pb.TagNumber(2)
  DNSSettings get dns => $_getN(1);
  @$pb.TagNumber(2)
  set dns(DNSSettings value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasDns() => $_has(1);
  @$pb.TagNumber(2)
  void clearDns() => $_clearField(2);
  @$pb.TagNumber(2)
  DNSSettings ensureDns() => $_ensure(1);

  @$pb.TagNumber(3)
  $core.bool get manualCaptcha => $_getBF(2);
  @$pb.TagNumber(3)
  set manualCaptcha($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasManualCaptcha() => $_has(2);
  @$pb.TagNumber(3)
  void clearManualCaptcha() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get vkPlatform => $_getSZ(3);
  @$pb.TagNumber(4)
  set vkPlatform($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasVkPlatform() => $_has(3);
  @$pb.TagNumber(4)
  void clearVkPlatform() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get turnAddress => $_getSZ(4);
  @$pb.TagNumber(5)
  set turnAddress($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTurnAddress() => $_has(4);
  @$pb.TagNumber(5)
  void clearTurnAddress() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get pacingMs => $_getI64(5);
  @$pb.TagNumber(6)
  set pacingMs($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPacingMs() => $_has(5);
  @$pb.TagNumber(6)
  void clearPacingMs() => $_clearField(6);
}

class ImportProfileRequest extends $pb.GeneratedMessage {
  factory ImportProfileRequest({
    $core.String? freeturnLink,
  }) {
    final result = ImportProfileRequest._();
    if (freeturnLink != null) result.freeturnLink = freeturnLink;
    return result;
  }

  ImportProfileRequest._();

  factory ImportProfileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ImportProfileRequest()..mergeFromBuffer(data, registry);
  factory ImportProfileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ImportProfileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ImportProfileRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ImportProfileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'freeturnLink')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ImportProfileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ImportProfileRequest copyWith(void Function(ImportProfileRequest) updates) =>
      super.copyWith((message) => updates(message as ImportProfileRequest))
          as ImportProfileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ImportProfileRequest() / ImportProfileRequest.new instead')
  static ImportProfileRequest create() => ImportProfileRequest._();
  static $pb.GeneratedMessage $_createMessage() => ImportProfileRequest._();
  @$core.override
  ImportProfileRequest createEmptyInstance() => ImportProfileRequest._();
  @$core.pragma('dart2js:noInline')
  static ImportProfileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ImportProfileRequest>(
          ImportProfileRequest.$_createMessage);
  static ImportProfileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get freeturnLink => $_getSZ(0);
  @$pb.TagNumber(1)
  set freeturnLink($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFreeturnLink() => $_has(0);
  @$pb.TagNumber(1)
  void clearFreeturnLink() => $_clearField(1);
}

class ProfileSummary extends $pb.GeneratedMessage {
  factory ProfileSummary({
    $core.String? id,
    $core.String? name,
    $core.String? provider,
    $core.String? peer,
    $core.String? turnTransport,
    $core.String? relayMode,
    $fixnum.Int64? streams,
    $fixnum.Int64? streamsPerCredential,
    $core.String? obfuscationMode,
  }) {
    final result = ProfileSummary._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (provider != null) result.provider = provider;
    if (peer != null) result.peer = peer;
    if (turnTransport != null) result.turnTransport = turnTransport;
    if (relayMode != null) result.relayMode = relayMode;
    if (streams != null) result.streams = streams;
    if (streamsPerCredential != null)
      result.streamsPerCredential = streamsPerCredential;
    if (obfuscationMode != null) result.obfuscationMode = obfuscationMode;
    return result;
  }

  ProfileSummary._();

  factory ProfileSummary.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileSummary()..mergeFromBuffer(data, registry);
  factory ProfileSummary.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileSummary()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileSummary',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileSummary.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'provider')
    ..aOS(4, _omitFieldNames ? '' : 'peer')
    ..aOS(5, _omitFieldNames ? '' : 'turnTransport')
    ..aOS(6, _omitFieldNames ? '' : 'relayMode')
    ..aInt64(7, _omitFieldNames ? '' : 'streams')
    ..aInt64(8, _omitFieldNames ? '' : 'streamsPerCredential')
    ..aOS(9, _omitFieldNames ? '' : 'obfuscationMode')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileSummary clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileSummary copyWith(void Function(ProfileSummary) updates) =>
      super.copyWith((message) => updates(message as ProfileSummary))
          as ProfileSummary;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileSummary() / ProfileSummary.new instead')
  static ProfileSummary create() => ProfileSummary._();
  static $pb.GeneratedMessage $_createMessage() => ProfileSummary._();
  @$core.override
  ProfileSummary createEmptyInstance() => ProfileSummary._();
  @$core.pragma('dart2js:noInline')
  static ProfileSummary getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProfileSummary>(
          ProfileSummary.$_createMessage);
  static ProfileSummary? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get provider => $_getSZ(2);
  @$pb.TagNumber(3)
  set provider($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProvider() => $_has(2);
  @$pb.TagNumber(3)
  void clearProvider() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get peer => $_getSZ(3);
  @$pb.TagNumber(4)
  set peer($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPeer() => $_has(3);
  @$pb.TagNumber(4)
  void clearPeer() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get turnTransport => $_getSZ(4);
  @$pb.TagNumber(5)
  set turnTransport($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTurnTransport() => $_has(4);
  @$pb.TagNumber(5)
  void clearTurnTransport() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get relayMode => $_getSZ(5);
  @$pb.TagNumber(6)
  set relayMode($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasRelayMode() => $_has(5);
  @$pb.TagNumber(6)
  void clearRelayMode() => $_clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get streams => $_getI64(6);
  @$pb.TagNumber(7)
  set streams($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasStreams() => $_has(6);
  @$pb.TagNumber(7)
  void clearStreams() => $_clearField(7);

  @$pb.TagNumber(8)
  $fixnum.Int64 get streamsPerCredential => $_getI64(7);
  @$pb.TagNumber(8)
  set streamsPerCredential($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasStreamsPerCredential() => $_has(7);
  @$pb.TagNumber(8)
  void clearStreamsPerCredential() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.String get obfuscationMode => $_getSZ(8);
  @$pb.TagNumber(9)
  set obfuscationMode($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasObfuscationMode() => $_has(8);
  @$pb.TagNumber(9)
  void clearObfuscationMode() => $_clearField(9);
}

class ListProfilesResponse extends $pb.GeneratedMessage {
  factory ListProfilesResponse({
    $core.Iterable<ProfileSummary>? profiles,
  }) {
    final result = ListProfilesResponse._();
    if (profiles != null) result.profiles.addAll(profiles);
    return result;
  }

  ListProfilesResponse._();

  factory ListProfilesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProfilesResponse()..mergeFromBuffer(data, registry);
  factory ListProfilesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListProfilesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProfilesResponse',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ListProfilesResponse.$_createMessage)
    ..pPM<ProfileSummary>(1, _omitFieldNames ? '' : 'profiles',
        subBuilder: ProfileSummary.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProfilesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProfilesResponse copyWith(void Function(ListProfilesResponse) updates) =>
      super.copyWith((message) => updates(message as ListProfilesResponse))
          as ListProfilesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListProfilesResponse() / ListProfilesResponse.new instead')
  static ListProfilesResponse create() => ListProfilesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListProfilesResponse._();
  @$core.override
  ListProfilesResponse createEmptyInstance() => ListProfilesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListProfilesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProfilesResponse>(
          ListProfilesResponse.$_createMessage);
  static ListProfilesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ProfileSummary> get profiles => $_getList(0);
}

class Profile extends $pb.GeneratedMessage {
  factory Profile({
    $core.String? id,
    Connection? connection,
  }) {
    final result = Profile._();
    if (id != null) result.id = id;
    if (connection != null) result.connection = connection;
    return result;
  }

  Profile._();

  factory Profile.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Profile()..mergeFromBuffer(data, registry);
  factory Profile.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Profile()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Profile',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: Profile.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOM<Connection>(2, _omitFieldNames ? '' : 'connection',
        subBuilder: Connection.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Profile clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Profile copyWith(void Function(Profile) updates) =>
      super.copyWith((message) => updates(message as Profile)) as Profile;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Profile() / Profile.new instead')
  static Profile create() => Profile._();
  static $pb.GeneratedMessage $_createMessage() => Profile._();
  @$core.override
  Profile createEmptyInstance() => Profile._();
  @$core.pragma('dart2js:noInline')
  static Profile getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Profile>(Profile.$_createMessage);
  static Profile? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  Connection get connection => $_getN(1);
  @$pb.TagNumber(2)
  set connection(Connection value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasConnection() => $_has(1);
  @$pb.TagNumber(2)
  void clearConnection() => $_clearField(2);
  @$pb.TagNumber(2)
  Connection ensureConnection() => $_ensure(1);
}

/// Исходные настройки без effective defaults. Локальный ID принадлежит Profile.
class Connection extends $pb.GeneratedMessage {
  factory Connection({
    $core.String? name,
    $core.String? clientId,
    $core.String? provider,
    $core.String? turnTransport,
    $core.String? relayMode,
    VKSettings? vk,
    DNSSettings? dns,
    $fixnum.Int64? streams,
    $fixnum.Int64? streamsPerCredential,
    $core.bool? bond,
    KCPSettings? kcp,
    WireGuardSettings? wg,
    $core.String? peer,
    Obfuscation? obfuscation,
  }) {
    final result = Connection._();
    if (name != null) result.name = name;
    if (clientId != null) result.clientId = clientId;
    if (provider != null) result.provider = provider;
    if (turnTransport != null) result.turnTransport = turnTransport;
    if (relayMode != null) result.relayMode = relayMode;
    if (vk != null) result.vk = vk;
    if (dns != null) result.dns = dns;
    if (streams != null) result.streams = streams;
    if (streamsPerCredential != null)
      result.streamsPerCredential = streamsPerCredential;
    if (bond != null) result.bond = bond;
    if (kcp != null) result.kcp = kcp;
    if (wg != null) result.wg = wg;
    if (peer != null) result.peer = peer;
    if (obfuscation != null) result.obfuscation = obfuscation;
    return result;
  }

  Connection._();

  factory Connection.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Connection()..mergeFromBuffer(data, registry);
  factory Connection.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Connection()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Connection',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: Connection.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'clientId')
    ..aOS(3, _omitFieldNames ? '' : 'provider')
    ..aOS(4, _omitFieldNames ? '' : 'turnTransport')
    ..aOS(5, _omitFieldNames ? '' : 'relayMode')
    ..aOM<VKSettings>(6, _omitFieldNames ? '' : 'vk',
        subBuilder: VKSettings.$_createMessage)
    ..aOM<DNSSettings>(7, _omitFieldNames ? '' : 'dns',
        subBuilder: DNSSettings.$_createMessage)
    ..aInt64(8, _omitFieldNames ? '' : 'streams')
    ..aInt64(9, _omitFieldNames ? '' : 'streamsPerCredential')
    ..aOB(10, _omitFieldNames ? '' : 'bond')
    ..aOM<KCPSettings>(11, _omitFieldNames ? '' : 'kcp',
        subBuilder: KCPSettings.$_createMessage)
    ..aOM<WireGuardSettings>(12, _omitFieldNames ? '' : 'wg',
        subBuilder: WireGuardSettings.$_createMessage)
    ..aOS(13, _omitFieldNames ? '' : 'peer')
    ..aOM<Obfuscation>(14, _omitFieldNames ? '' : 'obfuscation',
        subBuilder: Obfuscation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Connection clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Connection copyWith(void Function(Connection) updates) =>
      super.copyWith((message) => updates(message as Connection)) as Connection;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Connection() / Connection.new instead')
  static Connection create() => Connection._();
  static $pb.GeneratedMessage $_createMessage() => Connection._();
  @$core.override
  Connection createEmptyInstance() => Connection._();
  @$core.pragma('dart2js:noInline')
  static Connection getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Connection>(Connection.$_createMessage);
  static Connection? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get clientId => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasClientId() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get provider => $_getSZ(2);
  @$pb.TagNumber(3)
  set provider($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProvider() => $_has(2);
  @$pb.TagNumber(3)
  void clearProvider() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get turnTransport => $_getSZ(3);
  @$pb.TagNumber(4)
  set turnTransport($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTurnTransport() => $_has(3);
  @$pb.TagNumber(4)
  void clearTurnTransport() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get relayMode => $_getSZ(4);
  @$pb.TagNumber(5)
  set relayMode($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasRelayMode() => $_has(4);
  @$pb.TagNumber(5)
  void clearRelayMode() => $_clearField(5);

  @$pb.TagNumber(6)
  VKSettings get vk => $_getN(5);
  @$pb.TagNumber(6)
  set vk(VKSettings value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasVk() => $_has(5);
  @$pb.TagNumber(6)
  void clearVk() => $_clearField(6);
  @$pb.TagNumber(6)
  VKSettings ensureVk() => $_ensure(5);

  @$pb.TagNumber(7)
  DNSSettings get dns => $_getN(6);
  @$pb.TagNumber(7)
  set dns(DNSSettings value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasDns() => $_has(6);
  @$pb.TagNumber(7)
  void clearDns() => $_clearField(7);
  @$pb.TagNumber(7)
  DNSSettings ensureDns() => $_ensure(6);

  @$pb.TagNumber(8)
  $fixnum.Int64 get streams => $_getI64(7);
  @$pb.TagNumber(8)
  set streams($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasStreams() => $_has(7);
  @$pb.TagNumber(8)
  void clearStreams() => $_clearField(8);

  @$pb.TagNumber(9)
  $fixnum.Int64 get streamsPerCredential => $_getI64(8);
  @$pb.TagNumber(9)
  set streamsPerCredential($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasStreamsPerCredential() => $_has(8);
  @$pb.TagNumber(9)
  void clearStreamsPerCredential() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.bool get bond => $_getBF(9);
  @$pb.TagNumber(10)
  set bond($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasBond() => $_has(9);
  @$pb.TagNumber(10)
  void clearBond() => $_clearField(10);

  @$pb.TagNumber(11)
  KCPSettings get kcp => $_getN(10);
  @$pb.TagNumber(11)
  set kcp(KCPSettings value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasKcp() => $_has(10);
  @$pb.TagNumber(11)
  void clearKcp() => $_clearField(11);
  @$pb.TagNumber(11)
  KCPSettings ensureKcp() => $_ensure(10);

  @$pb.TagNumber(12)
  WireGuardSettings get wg => $_getN(11);
  @$pb.TagNumber(12)
  set wg(WireGuardSettings value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasWg() => $_has(11);
  @$pb.TagNumber(12)
  void clearWg() => $_clearField(12);
  @$pb.TagNumber(12)
  WireGuardSettings ensureWg() => $_ensure(11);

  @$pb.TagNumber(13)
  $core.String get peer => $_getSZ(12);
  @$pb.TagNumber(13)
  set peer($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasPeer() => $_has(12);
  @$pb.TagNumber(13)
  void clearPeer() => $_clearField(13);

  @$pb.TagNumber(14)
  Obfuscation get obfuscation => $_getN(13);
  @$pb.TagNumber(14)
  set obfuscation(Obfuscation value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasObfuscation() => $_has(13);
  @$pb.TagNumber(14)
  void clearObfuscation() => $_clearField(14);
  @$pb.TagNumber(14)
  Obfuscation ensureObfuscation() => $_ensure(13);
}

class VKSettings extends $pb.GeneratedMessage {
  factory VKSettings({
    $core.Iterable<$core.String>? callLinks,
    $core.bool? manualCaptcha,
    $core.String? platform,
    $core.String? turnAddress,
  }) {
    final result = VKSettings._();
    if (callLinks != null) result.callLinks.addAll(callLinks);
    if (manualCaptcha != null) result.manualCaptcha = manualCaptcha;
    if (platform != null) result.platform = platform;
    if (turnAddress != null) result.turnAddress = turnAddress;
    return result;
  }

  VKSettings._();

  factory VKSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VKSettings()..mergeFromBuffer(data, registry);
  factory VKSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      VKSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VKSettings',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: VKSettings.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'callLinks')
    ..aOB(2, _omitFieldNames ? '' : 'manualCaptcha')
    ..aOS(3, _omitFieldNames ? '' : 'platform')
    ..aOS(4, _omitFieldNames ? '' : 'turnAddress')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VKSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VKSettings copyWith(void Function(VKSettings) updates) =>
      super.copyWith((message) => updates(message as VKSettings)) as VKSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use VKSettings() / VKSettings.new instead')
  static VKSettings create() => VKSettings._();
  static $pb.GeneratedMessage $_createMessage() => VKSettings._();
  @$core.override
  VKSettings createEmptyInstance() => VKSettings._();
  @$core.pragma('dart2js:noInline')
  static VKSettings getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VKSettings>(VKSettings.$_createMessage);
  static VKSettings? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get callLinks => $_getList(0);

  @$pb.TagNumber(2)
  $core.bool get manualCaptcha => $_getBF(1);
  @$pb.TagNumber(2)
  set manualCaptcha($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasManualCaptcha() => $_has(1);
  @$pb.TagNumber(2)
  void clearManualCaptcha() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get platform => $_getSZ(2);
  @$pb.TagNumber(3)
  set platform($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPlatform() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlatform() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get turnAddress => $_getSZ(3);
  @$pb.TagNumber(4)
  set turnAddress($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTurnAddress() => $_has(3);
  @$pb.TagNumber(4)
  void clearTurnAddress() => $_clearField(4);
}

class DNSSettings extends $pb.GeneratedMessage {
  factory DNSSettings({
    $core.String? mode,
    $core.Iterable<$core.String>? servers,
  }) {
    final result = DNSSettings._();
    if (mode != null) result.mode = mode;
    if (servers != null) result.servers.addAll(servers);
    return result;
  }

  DNSSettings._();

  factory DNSSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DNSSettings()..mergeFromBuffer(data, registry);
  factory DNSSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DNSSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DNSSettings',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: DNSSettings.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'mode')
    ..pPS(2, _omitFieldNames ? '' : 'servers')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DNSSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DNSSettings copyWith(void Function(DNSSettings) updates) =>
      super.copyWith((message) => updates(message as DNSSettings))
          as DNSSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DNSSettings() / DNSSettings.new instead')
  static DNSSettings create() => DNSSettings._();
  static $pb.GeneratedMessage $_createMessage() => DNSSettings._();
  @$core.override
  DNSSettings createEmptyInstance() => DNSSettings._();
  @$core.pragma('dart2js:noInline')
  static DNSSettings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<DNSSettings>(
          DNSSettings.$_createMessage);
  static DNSSettings? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get mode => $_getSZ(0);
  @$pb.TagNumber(1)
  set mode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearMode() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get servers => $_getList(1);
}

class KCPSettings extends $pb.GeneratedMessage {
  factory KCPSettings({
    $fixnum.Int64? noDelay,
    $fixnum.Int64? interval,
    $fixnum.Int64? resend,
    $fixnum.Int64? nc,
    $fixnum.Int64? sndWnd,
    $fixnum.Int64? rcvWnd,
    $fixnum.Int64? mtu,
    $core.bool? ackNoDelay,
  }) {
    final result = KCPSettings._();
    if (noDelay != null) result.noDelay = noDelay;
    if (interval != null) result.interval = interval;
    if (resend != null) result.resend = resend;
    if (nc != null) result.nc = nc;
    if (sndWnd != null) result.sndWnd = sndWnd;
    if (rcvWnd != null) result.rcvWnd = rcvWnd;
    if (mtu != null) result.mtu = mtu;
    if (ackNoDelay != null) result.ackNoDelay = ackNoDelay;
    return result;
  }

  KCPSettings._();

  factory KCPSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      KCPSettings()..mergeFromBuffer(data, registry);
  factory KCPSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      KCPSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'KCPSettings',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: KCPSettings.$_createMessage)
    ..aInt64(1, _omitFieldNames ? '' : 'noDelay')
    ..aInt64(2, _omitFieldNames ? '' : 'interval')
    ..aInt64(3, _omitFieldNames ? '' : 'resend')
    ..aInt64(4, _omitFieldNames ? '' : 'nc')
    ..aInt64(5, _omitFieldNames ? '' : 'sndWnd')
    ..aInt64(6, _omitFieldNames ? '' : 'rcvWnd')
    ..aInt64(7, _omitFieldNames ? '' : 'mtu')
    ..aOB(8, _omitFieldNames ? '' : 'ackNoDelay')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  KCPSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  KCPSettings copyWith(void Function(KCPSettings) updates) =>
      super.copyWith((message) => updates(message as KCPSettings))
          as KCPSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use KCPSettings() / KCPSettings.new instead')
  static KCPSettings create() => KCPSettings._();
  static $pb.GeneratedMessage $_createMessage() => KCPSettings._();
  @$core.override
  KCPSettings createEmptyInstance() => KCPSettings._();
  @$core.pragma('dart2js:noInline')
  static KCPSettings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<KCPSettings>(
          KCPSettings.$_createMessage);
  static KCPSettings? _defaultInstance;

  @$pb.TagNumber(1)
  $fixnum.Int64 get noDelay => $_getI64(0);
  @$pb.TagNumber(1)
  set noDelay($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNoDelay() => $_has(0);
  @$pb.TagNumber(1)
  void clearNoDelay() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get interval => $_getI64(1);
  @$pb.TagNumber(2)
  set interval($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInterval() => $_has(1);
  @$pb.TagNumber(2)
  void clearInterval() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get resend => $_getI64(2);
  @$pb.TagNumber(3)
  set resend($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasResend() => $_has(2);
  @$pb.TagNumber(3)
  void clearResend() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get nc => $_getI64(3);
  @$pb.TagNumber(4)
  set nc($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasNc() => $_has(3);
  @$pb.TagNumber(4)
  void clearNc() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get sndWnd => $_getI64(4);
  @$pb.TagNumber(5)
  set sndWnd($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSndWnd() => $_has(4);
  @$pb.TagNumber(5)
  void clearSndWnd() => $_clearField(5);

  @$pb.TagNumber(6)
  $fixnum.Int64 get rcvWnd => $_getI64(5);
  @$pb.TagNumber(6)
  set rcvWnd($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasRcvWnd() => $_has(5);
  @$pb.TagNumber(6)
  void clearRcvWnd() => $_clearField(6);

  @$pb.TagNumber(7)
  $fixnum.Int64 get mtu => $_getI64(6);
  @$pb.TagNumber(7)
  set mtu($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasMtu() => $_has(6);
  @$pb.TagNumber(7)
  void clearMtu() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.bool get ackNoDelay => $_getBF(7);
  @$pb.TagNumber(8)
  set ackNoDelay($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAckNoDelay() => $_has(7);
  @$pb.TagNumber(8)
  void clearAckNoDelay() => $_clearField(8);
}

class WireGuardSettings extends $pb.GeneratedMessage {
  factory WireGuardSettings({
    $core.String? config,
  }) {
    final result = WireGuardSettings._();
    if (config != null) result.config = config;
    return result;
  }

  WireGuardSettings._();

  factory WireGuardSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WireGuardSettings()..mergeFromBuffer(data, registry);
  factory WireGuardSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WireGuardSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WireGuardSettings',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: WireGuardSettings.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'config')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WireGuardSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WireGuardSettings copyWith(void Function(WireGuardSettings) updates) =>
      super.copyWith((message) => updates(message as WireGuardSettings))
          as WireGuardSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use WireGuardSettings() / WireGuardSettings.new instead')
  static WireGuardSettings create() => WireGuardSettings._();
  static $pb.GeneratedMessage $_createMessage() => WireGuardSettings._();
  @$core.override
  WireGuardSettings createEmptyInstance() => WireGuardSettings._();
  @$core.pragma('dart2js:noInline')
  static WireGuardSettings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WireGuardSettings>(
          WireGuardSettings.$_createMessage);
  static WireGuardSettings? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get config => $_getSZ(0);
  @$pb.TagNumber(1)
  set config($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasConfig() => $_has(0);
  @$pb.TagNumber(1)
  void clearConfig() => $_clearField(1);
}

class Obfuscation extends $pb.GeneratedMessage {
  factory Obfuscation({
    $core.String? mode,
    $core.String? key,
    $fixnum.Int64? timingMs,
  }) {
    final result = Obfuscation._();
    if (mode != null) result.mode = mode;
    if (key != null) result.key = key;
    if (timingMs != null) result.timingMs = timingMs;
    return result;
  }

  Obfuscation._();

  factory Obfuscation.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Obfuscation()..mergeFromBuffer(data, registry);
  factory Obfuscation.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Obfuscation()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Obfuscation',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: Obfuscation.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'mode')
    ..aOS(2, _omitFieldNames ? '' : 'key')
    ..aInt64(3, _omitFieldNames ? '' : 'timingMs')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Obfuscation clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Obfuscation copyWith(void Function(Obfuscation) updates) =>
      super.copyWith((message) => updates(message as Obfuscation))
          as Obfuscation;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Obfuscation() / Obfuscation.new instead')
  static Obfuscation create() => Obfuscation._();
  static $pb.GeneratedMessage $_createMessage() => Obfuscation._();
  @$core.override
  Obfuscation createEmptyInstance() => Obfuscation._();
  @$core.pragma('dart2js:noInline')
  static Obfuscation getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Obfuscation>(
          Obfuscation.$_createMessage);
  static Obfuscation? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get mode => $_getSZ(0);
  @$pb.TagNumber(1)
  set mode($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearMode() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get key => $_getSZ(1);
  @$pb.TagNumber(2)
  set key($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKey() => $_has(1);
  @$pb.TagNumber(2)
  void clearKey() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get timingMs => $_getI64(2);
  @$pb.TagNumber(3)
  set timingMs($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTimingMs() => $_has(2);
  @$pb.TagNumber(3)
  void clearTimingMs() => $_clearField(3);
}

class Snapshot extends $pb.GeneratedMessage {
  factory Snapshot({
    State? state,
    $fixnum.Int64? generation,
    $core.String? profileId,
    $core.String? mixedAddress,
    $core.String? error,
    Latency? latency,
    ProxyMode? activeMode,
  }) {
    final result = Snapshot._();
    if (state != null) result.state = state;
    if (generation != null) result.generation = generation;
    if (profileId != null) result.profileId = profileId;
    if (mixedAddress != null) result.mixedAddress = mixedAddress;
    if (error != null) result.error = error;
    if (latency != null) result.latency = latency;
    if (activeMode != null) result.activeMode = activeMode;
    return result;
  }

  Snapshot._();

  factory Snapshot.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Snapshot()..mergeFromBuffer(data, registry);
  factory Snapshot.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Snapshot()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Snapshot',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: Snapshot.$_createMessage)
    ..aE<State>(1, _omitFieldNames ? '' : 'state', enumValues: State.values)
    ..a<$fixnum.Int64>(
        2, _omitFieldNames ? '' : 'generation', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, _omitFieldNames ? '' : 'profileId')
    ..aOS(4, _omitFieldNames ? '' : 'mixedAddress')
    ..aOS(5, _omitFieldNames ? '' : 'error')
    ..aOM<Latency>(6, _omitFieldNames ? '' : 'latency',
        subBuilder: Latency.$_createMessage)
    ..aE<ProxyMode>(7, _omitFieldNames ? '' : 'activeMode',
        enumValues: ProxyMode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Snapshot clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Snapshot copyWith(void Function(Snapshot) updates) =>
      super.copyWith((message) => updates(message as Snapshot)) as Snapshot;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Snapshot() / Snapshot.new instead')
  static Snapshot create() => Snapshot._();
  static $pb.GeneratedMessage $_createMessage() => Snapshot._();
  @$core.override
  Snapshot createEmptyInstance() => Snapshot._();
  @$core.pragma('dart2js:noInline')
  static Snapshot getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Snapshot>(Snapshot.$_createMessage);
  static Snapshot? _defaultInstance;

  @$pb.TagNumber(1)
  State get state => $_getN(0);
  @$pb.TagNumber(1)
  set state(State value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get generation => $_getI64(1);
  @$pb.TagNumber(2)
  set generation($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasGeneration() => $_has(1);
  @$pb.TagNumber(2)
  void clearGeneration() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get profileId => $_getSZ(2);
  @$pb.TagNumber(3)
  set profileId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProfileId() => $_has(2);
  @$pb.TagNumber(3)
  void clearProfileId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get mixedAddress => $_getSZ(3);
  @$pb.TagNumber(4)
  set mixedAddress($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMixedAddress() => $_has(3);
  @$pb.TagNumber(4)
  void clearMixedAddress() => $_clearField(4);

  /// Очищенное сообщение controller, без исходной ошибки engine/storage.
  @$pb.TagNumber(5)
  $core.String get error => $_getSZ(4);
  @$pb.TagNumber(5)
  set error($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasError() => $_has(4);
  @$pb.TagNumber(5)
  void clearError() => $_clearField(5);

  @$pb.TagNumber(6)
  Latency get latency => $_getN(5);
  @$pb.TagNumber(6)
  set latency(Latency value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasLatency() => $_has(5);
  @$pb.TagNumber(6)
  void clearLatency() => $_clearField(6);
  @$pb.TagNumber(6)
  Latency ensureLatency() => $_ensure(5);

  @$pb.TagNumber(7)
  ProxyMode get activeMode => $_getN(6);
  @$pb.TagNumber(7)
  set activeMode(ProxyMode value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasActiveMode() => $_has(6);
  @$pb.TagNumber(7)
  void clearActiveMode() => $_clearField(7);
}

class Latency extends $pb.GeneratedMessage {
  factory Latency({
    $core.bool? checking,
    $fixnum.Int64? delayMs,
    $fixnum.Int64? checkedAt,
    $fixnum.Int64? successAt,
    $core.String? error,
  }) {
    final result = Latency._();
    if (checking != null) result.checking = checking;
    if (delayMs != null) result.delayMs = delayMs;
    if (checkedAt != null) result.checkedAt = checkedAt;
    if (successAt != null) result.successAt = successAt;
    if (error != null) result.error = error;
    return result;
  }

  Latency._();

  factory Latency.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Latency()..mergeFromBuffer(data, registry);
  factory Latency.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Latency()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Latency',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: Latency.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'checking')
    ..aInt64(2, _omitFieldNames ? '' : 'delayMs')
    ..aInt64(3, _omitFieldNames ? '' : 'checkedAt')
    ..aInt64(4, _omitFieldNames ? '' : 'successAt')
    ..aOS(5, _omitFieldNames ? '' : 'error')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Latency clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Latency copyWith(void Function(Latency) updates) =>
      super.copyWith((message) => updates(message as Latency)) as Latency;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Latency() / Latency.new instead')
  static Latency create() => Latency._();
  static $pb.GeneratedMessage $_createMessage() => Latency._();
  @$core.override
  Latency createEmptyInstance() => Latency._();
  @$core.pragma('dart2js:noInline')
  static Latency getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Latency>(Latency.$_createMessage);
  static Latency? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get checking => $_getBF(0);
  @$pb.TagNumber(1)
  set checking($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChecking() => $_has(0);
  @$pb.TagNumber(1)
  void clearChecking() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get delayMs => $_getI64(1);
  @$pb.TagNumber(2)
  set delayMs($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDelayMs() => $_has(1);
  @$pb.TagNumber(2)
  void clearDelayMs() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get checkedAt => $_getI64(2);
  @$pb.TagNumber(3)
  set checkedAt($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCheckedAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearCheckedAt() => $_clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get successAt => $_getI64(3);
  @$pb.TagNumber(4)
  set successAt($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSuccessAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearSuccessAt() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get error => $_getSZ(4);
  @$pb.TagNumber(5)
  set error($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasError() => $_has(4);
  @$pb.TagNumber(5)
  void clearError() => $_clearField(5);
}

/// Transfer uses user-selected local paths; archive contents never enter summaries.
class ExportProfilesRequest extends $pb.GeneratedMessage {
  factory ExportProfilesRequest({
    $core.Iterable<$core.String>? ids,
    $core.String? path,
  }) {
    final result = ExportProfilesRequest._();
    if (ids != null) result.ids.addAll(ids);
    if (path != null) result.path = path;
    return result;
  }

  ExportProfilesRequest._();

  factory ExportProfilesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ExportProfilesRequest()..mergeFromBuffer(data, registry);
  factory ExportProfilesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ExportProfilesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ExportProfilesRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ExportProfilesRequest.$_createMessage)
    ..pPS(1, _omitFieldNames ? '' : 'ids')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ExportProfilesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ExportProfilesRequest copyWith(
          void Function(ExportProfilesRequest) updates) =>
      super.copyWith((message) => updates(message as ExportProfilesRequest))
          as ExportProfilesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ExportProfilesRequest() / ExportProfilesRequest.new instead')
  static ExportProfilesRequest create() => ExportProfilesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ExportProfilesRequest._();
  @$core.override
  ExportProfilesRequest createEmptyInstance() => ExportProfilesRequest._();
  @$core.pragma('dart2js:noInline')
  static ExportProfilesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ExportProfilesRequest>(
          ExportProfilesRequest.$_createMessage);
  static ExportProfilesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<$core.String> get ids => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);
}

class ProfileFileRequest extends $pb.GeneratedMessage {
  factory ProfileFileRequest({
    $core.String? path,
  }) {
    final result = ProfileFileRequest._();
    if (path != null) result.path = path;
    return result;
  }

  ProfileFileRequest._();

  factory ProfileFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileFileRequest()..mergeFromBuffer(data, registry);
  factory ProfileFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileFileRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileFileRequest copyWith(void Function(ProfileFileRequest) updates) =>
      super.copyWith((message) => updates(message as ProfileFileRequest))
          as ProfileFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileFileRequest() / ProfileFileRequest.new instead')
  static ProfileFileRequest create() => ProfileFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => ProfileFileRequest._();
  @$core.override
  ProfileFileRequest createEmptyInstance() => ProfileFileRequest._();
  @$core.pragma('dart2js:noInline')
  static ProfileFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProfileFileRequest>(
          ProfileFileRequest.$_createMessage);
  static ProfileFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class ProfileFileEntry extends $pb.GeneratedMessage {
  factory ProfileFileEntry({
    ProfileSummary? profile,
    ProfileSummary? existing,
  }) {
    final result = ProfileFileEntry._();
    if (profile != null) result.profile = profile;
    if (existing != null) result.existing = existing;
    return result;
  }

  ProfileFileEntry._();

  factory ProfileFileEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileFileEntry()..mergeFromBuffer(data, registry);
  factory ProfileFileEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileFileEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileFileEntry',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileFileEntry.$_createMessage)
    ..aOM<ProfileSummary>(1, _omitFieldNames ? '' : 'profile',
        subBuilder: ProfileSummary.$_createMessage)
    ..aOM<ProfileSummary>(2, _omitFieldNames ? '' : 'existing',
        subBuilder: ProfileSummary.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileFileEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileFileEntry copyWith(void Function(ProfileFileEntry) updates) =>
      super.copyWith((message) => updates(message as ProfileFileEntry))
          as ProfileFileEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileFileEntry() / ProfileFileEntry.new instead')
  static ProfileFileEntry create() => ProfileFileEntry._();
  static $pb.GeneratedMessage $_createMessage() => ProfileFileEntry._();
  @$core.override
  ProfileFileEntry createEmptyInstance() => ProfileFileEntry._();
  @$core.pragma('dart2js:noInline')
  static ProfileFileEntry getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ProfileFileEntry>(
          ProfileFileEntry.$_createMessage);
  static ProfileFileEntry? _defaultInstance;

  @$pb.TagNumber(1)
  ProfileSummary get profile => $_getN(0);
  @$pb.TagNumber(1)
  set profile(ProfileSummary value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProfile() => $_has(0);
  @$pb.TagNumber(1)
  void clearProfile() => $_clearField(1);
  @$pb.TagNumber(1)
  ProfileSummary ensureProfile() => $_ensure(0);

  @$pb.TagNumber(2)
  ProfileSummary get existing => $_getN(1);
  @$pb.TagNumber(2)
  set existing(ProfileSummary value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasExisting() => $_has(1);
  @$pb.TagNumber(2)
  void clearExisting() => $_clearField(2);
  @$pb.TagNumber(2)
  ProfileSummary ensureExisting() => $_ensure(1);
}

class ProfileFilePreview extends $pb.GeneratedMessage {
  factory ProfileFilePreview({
    $core.String? token,
    $core.Iterable<ProfileFileEntry>? entries,
  }) {
    final result = ProfileFilePreview._();
    if (token != null) result.token = token;
    if (entries != null) result.entries.addAll(entries);
    return result;
  }

  ProfileFilePreview._();

  factory ProfileFilePreview.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileFilePreview()..mergeFromBuffer(data, registry);
  factory ProfileFilePreview.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileFilePreview()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileFilePreview',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileFilePreview.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..pPM<ProfileFileEntry>(2, _omitFieldNames ? '' : 'entries',
        subBuilder: ProfileFileEntry.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileFilePreview clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileFilePreview copyWith(void Function(ProfileFilePreview) updates) =>
      super.copyWith((message) => updates(message as ProfileFilePreview))
          as ProfileFilePreview;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ProfileFilePreview() / ProfileFilePreview.new instead')
  static ProfileFilePreview create() => ProfileFilePreview._();
  static $pb.GeneratedMessage $_createMessage() => ProfileFilePreview._();
  @$core.override
  ProfileFilePreview createEmptyInstance() => ProfileFilePreview._();
  @$core.pragma('dart2js:noInline')
  static ProfileFilePreview getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProfileFilePreview>(
          ProfileFilePreview.$_createMessage);
  static ProfileFilePreview? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<ProfileFileEntry> get entries => $_getList(1);
}

class ProfileImportChoice extends $pb.GeneratedMessage {
  factory ProfileImportChoice({
    $core.int? index,
    ProfileImportAction? action,
  }) {
    final result = ProfileImportChoice._();
    if (index != null) result.index = index;
    if (action != null) result.action = action;
    return result;
  }

  ProfileImportChoice._();

  factory ProfileImportChoice.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileImportChoice()..mergeFromBuffer(data, registry);
  factory ProfileImportChoice.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileImportChoice()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileImportChoice',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileImportChoice.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'index', fieldType: $pb.PbFieldType.OU3)
    ..aE<ProfileImportAction>(2, _omitFieldNames ? '' : 'action',
        enumValues: ProfileImportAction.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileImportChoice clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileImportChoice copyWith(void Function(ProfileImportChoice) updates) =>
      super.copyWith((message) => updates(message as ProfileImportChoice))
          as ProfileImportChoice;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ProfileImportChoice() / ProfileImportChoice.new instead')
  static ProfileImportChoice create() => ProfileImportChoice._();
  static $pb.GeneratedMessage $_createMessage() => ProfileImportChoice._();
  @$core.override
  ProfileImportChoice createEmptyInstance() => ProfileImportChoice._();
  @$core.pragma('dart2js:noInline')
  static ProfileImportChoice getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProfileImportChoice>(
          ProfileImportChoice.$_createMessage);
  static ProfileImportChoice? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get index => $_getIZ(0);
  @$pb.TagNumber(1)
  set index($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIndex() => $_has(0);
  @$pb.TagNumber(1)
  void clearIndex() => $_clearField(1);

  @$pb.TagNumber(2)
  ProfileImportAction get action => $_getN(1);
  @$pb.TagNumber(2)
  set action(ProfileImportAction value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasAction() => $_has(1);
  @$pb.TagNumber(2)
  void clearAction() => $_clearField(2);
}

class ImportProfileFileRequest extends $pb.GeneratedMessage {
  factory ImportProfileFileRequest({
    $core.String? token,
    $core.Iterable<ProfileImportChoice>? choices,
  }) {
    final result = ImportProfileFileRequest._();
    if (token != null) result.token = token;
    if (choices != null) result.choices.addAll(choices);
    return result;
  }

  ImportProfileFileRequest._();

  factory ImportProfileFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ImportProfileFileRequest()..mergeFromBuffer(data, registry);
  factory ImportProfileFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ImportProfileFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ImportProfileFileRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ImportProfileFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..pPM<ProfileImportChoice>(2, _omitFieldNames ? '' : 'choices',
        subBuilder: ProfileImportChoice.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ImportProfileFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ImportProfileFileRequest copyWith(
          void Function(ImportProfileFileRequest) updates) =>
      super.copyWith((message) => updates(message as ImportProfileFileRequest))
          as ImportProfileFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ImportProfileFileRequest() / ImportProfileFileRequest.new instead')
  static ImportProfileFileRequest create() => ImportProfileFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => ImportProfileFileRequest._();
  @$core.override
  ImportProfileFileRequest createEmptyInstance() =>
      ImportProfileFileRequest._();
  @$core.pragma('dart2js:noInline')
  static ImportProfileFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ImportProfileFileRequest>(
          ImportProfileFileRequest.$_createMessage);
  static ImportProfileFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<ProfileImportChoice> get choices => $_getList(1);
}

/// Status is imported, skipped, changed or failed. Never contains an underlying error.
class ProfileImportResult extends $pb.GeneratedMessage {
  factory ProfileImportResult({
    $core.int? index,
    $core.String? status,
  }) {
    final result = ProfileImportResult._();
    if (index != null) result.index = index;
    if (status != null) result.status = status;
    return result;
  }

  ProfileImportResult._();

  factory ProfileImportResult.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileImportResult()..mergeFromBuffer(data, registry);
  factory ProfileImportResult.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ProfileImportResult()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileImportResult',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ProfileImportResult.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'index', fieldType: $pb.PbFieldType.OU3)
    ..aOS(2, _omitFieldNames ? '' : 'status')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileImportResult clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileImportResult copyWith(void Function(ProfileImportResult) updates) =>
      super.copyWith((message) => updates(message as ProfileImportResult))
          as ProfileImportResult;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ProfileImportResult() / ProfileImportResult.new instead')
  static ProfileImportResult create() => ProfileImportResult._();
  static $pb.GeneratedMessage $_createMessage() => ProfileImportResult._();
  @$core.override
  ProfileImportResult createEmptyInstance() => ProfileImportResult._();
  @$core.pragma('dart2js:noInline')
  static ProfileImportResult getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProfileImportResult>(
          ProfileImportResult.$_createMessage);
  static ProfileImportResult? _defaultInstance;

  @$pb.TagNumber(1)
  $core.int get index => $_getIZ(0);
  @$pb.TagNumber(1)
  set index($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIndex() => $_has(0);
  @$pb.TagNumber(1)
  void clearIndex() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get status => $_getSZ(1);
  @$pb.TagNumber(2)
  set status($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);
}

class ImportProfileFileResponse extends $pb.GeneratedMessage {
  factory ImportProfileFileResponse({
    $core.Iterable<ProfileImportResult>? results,
  }) {
    final result = ImportProfileFileResponse._();
    if (results != null) result.results.addAll(results);
    return result;
  }

  ImportProfileFileResponse._();

  factory ImportProfileFileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ImportProfileFileResponse()..mergeFromBuffer(data, registry);
  factory ImportProfileFileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ImportProfileFileResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ImportProfileFileResponse',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'wireglow.control'),
      createEmptyInstance: ImportProfileFileResponse.$_createMessage)
    ..pPM<ProfileImportResult>(1, _omitFieldNames ? '' : 'results',
        subBuilder: ProfileImportResult.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ImportProfileFileResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ImportProfileFileResponse copyWith(
          void Function(ImportProfileFileResponse) updates) =>
      super.copyWith((message) => updates(message as ImportProfileFileResponse))
          as ImportProfileFileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ImportProfileFileResponse() / ImportProfileFileResponse.new instead')
  static ImportProfileFileResponse create() => ImportProfileFileResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ImportProfileFileResponse._();
  @$core.override
  ImportProfileFileResponse createEmptyInstance() =>
      ImportProfileFileResponse._();
  @$core.pragma('dart2js:noInline')
  static ImportProfileFileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ImportProfileFileResponse>(
          ImportProfileFileResponse.$_createMessage);
  static ImportProfileFileResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ProfileImportResult> get results => $_getList(0);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
