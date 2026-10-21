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

import 'package:protobuf/protobuf.dart' as $pb;

class ProxyMode extends $pb.ProtobufEnum {
  static const ProxyMode PROXY_MODE_UNSPECIFIED =
      ProxyMode._(0, _omitEnumNames ? '' : 'PROXY_MODE_UNSPECIFIED');
  static const ProxyMode PROXY_MODE_LOCAL =
      ProxyMode._(1, _omitEnumNames ? '' : 'PROXY_MODE_LOCAL');
  static const ProxyMode PROXY_MODE_SYSTEM =
      ProxyMode._(2, _omitEnumNames ? '' : 'PROXY_MODE_SYSTEM');
  static const ProxyMode PROXY_MODE_TUN =
      ProxyMode._(3, _omitEnumNames ? '' : 'PROXY_MODE_TUN');

  static const $core.List<ProxyMode> values = <ProxyMode>[
    PROXY_MODE_UNSPECIFIED,
    PROXY_MODE_LOCAL,
    PROXY_MODE_SYSTEM,
    PROXY_MODE_TUN,
  ];

  static final $core.List<ProxyMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static ProxyMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ProxyMode._(super.value, super.name);
}

class State extends $pb.ProtobufEnum {
  static const State STATE_UNSPECIFIED =
      State._(0, _omitEnumNames ? '' : 'STATE_UNSPECIFIED');
  static const State STATE_IDLE =
      State._(1, _omitEnumNames ? '' : 'STATE_IDLE');
  static const State STATE_VALIDATING =
      State._(2, _omitEnumNames ? '' : 'STATE_VALIDATING');
  static const State STATE_STARTING_RELAY =
      State._(3, _omitEnumNames ? '' : 'STATE_STARTING_RELAY');
  static const State STATE_STARTING_MIHOMO =
      State._(4, _omitEnumNames ? '' : 'STATE_STARTING_MIHOMO');
  static const State STATE_PROBING =
      State._(5, _omitEnumNames ? '' : 'STATE_PROBING');
  static const State STATE_ACTIVATING_NETWORK =
      State._(6, _omitEnumNames ? '' : 'STATE_ACTIVATING_NETWORK');
  static const State STATE_CONNECTED =
      State._(7, _omitEnumNames ? '' : 'STATE_CONNECTED');
  static const State STATE_STOPPING =
      State._(8, _omitEnumNames ? '' : 'STATE_STOPPING');
  static const State STATE_FAILED =
      State._(9, _omitEnumNames ? '' : 'STATE_FAILED');

  static const $core.List<State> values = <State>[
    STATE_UNSPECIFIED,
    STATE_IDLE,
    STATE_VALIDATING,
    STATE_STARTING_RELAY,
    STATE_STARTING_MIHOMO,
    STATE_PROBING,
    STATE_ACTIVATING_NETWORK,
    STATE_CONNECTED,
    STATE_STOPPING,
    STATE_FAILED,
  ];

  static final $core.List<State?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 9);
  static State? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const State._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
