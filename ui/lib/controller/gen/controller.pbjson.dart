// This is a generated file - do not edit.
//
// Generated from controller.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use proxyModeDescriptor instead')
const ProxyMode$json = {
  '1': 'ProxyMode',
  '2': [
    {'1': 'PROXY_MODE_UNSPECIFIED', '2': 0},
    {'1': 'PROXY_MODE_LOCAL', '2': 1},
    {'1': 'PROXY_MODE_SYSTEM', '2': 2},
    {'1': 'PROXY_MODE_TUN', '2': 3},
  ],
};

/// Descriptor for `ProxyMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List proxyModeDescriptor = $convert.base64Decode(
    'CglQcm94eU1vZGUSGgoWUFJPWFlfTU9ERV9VTlNQRUNJRklFRBAAEhQKEFBST1hZX01PREVfTE'
    '9DQUwQARIVChFQUk9YWV9NT0RFX1NZU1RFTRACEhIKDlBST1hZX01PREVfVFVOEAM=');

@$core.Deprecated('Use stateDescriptor instead')
const State$json = {
  '1': 'State',
  '2': [
    {'1': 'STATE_UNSPECIFIED', '2': 0},
    {'1': 'STATE_IDLE', '2': 1},
    {'1': 'STATE_VALIDATING', '2': 2},
    {'1': 'STATE_STARTING_RELAY', '2': 3},
    {'1': 'STATE_STARTING_MIHOMO', '2': 4},
    {'1': 'STATE_PROBING', '2': 5},
    {'1': 'STATE_ACTIVATING_NETWORK', '2': 6},
    {'1': 'STATE_CONNECTED', '2': 7},
    {'1': 'STATE_STOPPING', '2': 8},
    {'1': 'STATE_FAILED', '2': 9},
  ],
};

/// Descriptor for `State`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List stateDescriptor = $convert.base64Decode(
    'CgVTdGF0ZRIVChFTVEFURV9VTlNQRUNJRklFRBAAEg4KClNUQVRFX0lETEUQARIUChBTVEFURV'
    '9WQUxJREFUSU5HEAISGAoUU1RBVEVfU1RBUlRJTkdfUkVMQVkQAxIZChVTVEFURV9TVEFSVElO'
    'R19NSUhPTU8QBBIRCg1TVEFURV9QUk9CSU5HEAUSHAoYU1RBVEVfQUNUSVZBVElOR19ORVRXT1'
    'JLEAYSEwoPU1RBVEVfQ09OTkVDVEVEEAcSEgoOU1RBVEVfU1RPUFBJTkcQCBIQCgxTVEFURV9G'
    'QUlMRUQQCQ==');

@$core.Deprecated('Use profileImportActionDescriptor instead')
const ProfileImportAction$json = {
  '1': 'ProfileImportAction',
  '2': [
    {'1': 'PROFILE_IMPORT_ACTION_UNSPECIFIED', '2': 0},
    {'1': 'PROFILE_IMPORT_ACTION_ADD', '2': 1},
    {'1': 'PROFILE_IMPORT_ACTION_SKIP', '2': 2},
    {'1': 'PROFILE_IMPORT_ACTION_COPY', '2': 3},
    {'1': 'PROFILE_IMPORT_ACTION_REPLACE', '2': 4},
  ],
};

/// Descriptor for `ProfileImportAction`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List profileImportActionDescriptor = $convert.base64Decode(
    'ChNQcm9maWxlSW1wb3J0QWN0aW9uEiUKIVBST0ZJTEVfSU1QT1JUX0FDVElPTl9VTlNQRUNJRk'
    'lFRBAAEh0KGVBST0ZJTEVfSU1QT1JUX0FDVElPTl9BREQQARIeChpQUk9GSUxFX0lNUE9SVF9B'
    'Q1RJT05fU0tJUBACEh4KGlBST0ZJTEVfSU1QT1JUX0FDVElPTl9DT1BZEAMSIQodUFJPRklMRV'
    '9JTVBPUlRfQUNUSU9OX1JFUExBQ0UQBA==');

@$core.Deprecated('Use helloResponseDescriptor instead')
const HelloResponse$json = {
  '1': 'HelloResponse',
  '2': [
    {'1': 'build_id', '3': 1, '4': 1, '5': 9, '10': 'buildId'},
    {'1': 'epoch', '3': 2, '4': 1, '5': 9, '10': 'epoch'},
  ],
};

/// Descriptor for `HelloResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List helloResponseDescriptor = $convert.base64Decode(
    'Cg1IZWxsb1Jlc3BvbnNlEhkKCGJ1aWxkX2lkGAEgASgJUgdidWlsZElkEhQKBWVwb2NoGAIgAS'
    'gJUgVlcG9jaA==');

@$core.Deprecated('Use connectRequestDescriptor instead')
const ConnectRequest$json = {
  '1': 'ConnectRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'mode',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.wireglow.control.ProxyMode',
      '10': 'mode'
    },
  ],
};

/// Descriptor for `ConnectRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List connectRequestDescriptor = $convert.base64Decode(
    'Cg5Db25uZWN0UmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSLwoEbW9kZRgCIAEoDjIbLndpcmVnbG'
    '93LmNvbnRyb2wuUHJveHlNb2RlUgRtb2Rl');

@$core.Deprecated('Use proxyModeRequestDescriptor instead')
const ProxyModeRequest$json = {
  '1': 'ProxyModeRequest',
  '2': [
    {
      '1': 'mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.wireglow.control.ProxyMode',
      '10': 'mode'
    },
  ],
};

/// Descriptor for `ProxyModeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List proxyModeRequestDescriptor = $convert.base64Decode(
    'ChBQcm94eU1vZGVSZXF1ZXN0Ei8KBG1vZGUYASABKA4yGy53aXJlZ2xvdy5jb250cm9sLlByb3'
    'h5TW9kZVIEbW9kZQ==');

@$core.Deprecated('Use clientSettingsDescriptor instead')
const ClientSettings$json = {
  '1': 'ClientSettings',
  '2': [
    {
      '1': 'preferred_mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.wireglow.control.ProxyMode',
      '10': 'preferredMode'
    },
    {'1': 'proxy_address', '3': 2, '4': 1, '5': 9, '10': 'proxyAddress'},
    {'1': 'relay_address', '3': 3, '4': 1, '5': 9, '10': 'relayAddress'},
    {'1': 'api_port', '3': 4, '4': 1, '5': 13, '10': 'apiPort'},
  ],
};

/// Descriptor for `ClientSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List clientSettingsDescriptor = $convert.base64Decode(
    'Cg5DbGllbnRTZXR0aW5ncxJCCg5wcmVmZXJyZWRfbW9kZRgBIAEoDjIbLndpcmVnbG93LmNvbn'
    'Ryb2wuUHJveHlNb2RlUg1wcmVmZXJyZWRNb2RlEiMKDXByb3h5X2FkZHJlc3MYAiABKAlSDHBy'
    'b3h5QWRkcmVzcxIjCg1yZWxheV9hZGRyZXNzGAMgASgJUgxyZWxheUFkZHJlc3MSGQoIYXBpX3'
    'BvcnQYBCABKA1SB2FwaVBvcnQ=');

@$core.Deprecated('Use capabilitiesDescriptor instead')
const Capabilities$json = {
  '1': 'Capabilities',
  '2': [
    {
      '1': 'system_proxy_unavailable_reason',
      '3': 1,
      '4': 1,
      '5': 9,
      '10': 'systemProxyUnavailableReason'
    },
    {
      '1': 'tun_unavailable_reason',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'tunUnavailableReason'
    },
  ],
};

/// Descriptor for `Capabilities`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List capabilitiesDescriptor = $convert.base64Decode(
    'CgxDYXBhYmlsaXRpZXMSRQofc3lzdGVtX3Byb3h5X3VuYXZhaWxhYmxlX3JlYXNvbhgBIAEoCV'
    'Icc3lzdGVtUHJveHlVbmF2YWlsYWJsZVJlYXNvbhI0ChZ0dW5fdW5hdmFpbGFibGVfcmVhc29u'
    'GAIgASgJUhR0dW5VbmF2YWlsYWJsZVJlYXNvbg==');

@$core.Deprecated('Use profileRequestDescriptor instead')
const ProfileRequest$json = {
  '1': 'ProfileRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `ProfileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileRequestDescriptor =
    $convert.base64Decode('Cg5Qcm9maWxlUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQ=');

@$core.Deprecated('Use profileEditsDescriptor instead')
const ProfileEdits$json = {
  '1': 'ProfileEdits',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'call_links', '3': 3, '4': 3, '5': 9, '10': 'callLinks'},
    {'1': 'streams', '3': 4, '4': 1, '5': 3, '10': 'streams'},
    {
      '1': 'streams_per_credential',
      '3': 5,
      '4': 1,
      '5': 3,
      '10': 'streamsPerCredential'
    },
    {
      '1': 'advanced',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.ProfileAdvancedEdits',
      '10': 'advanced'
    },
  ],
};

/// Descriptor for `ProfileEdits`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileEditsDescriptor = $convert.base64Decode(
    'CgxQcm9maWxlRWRpdHMSDgoCaWQYASABKAlSAmlkEhIKBG5hbWUYAiABKAlSBG5hbWUSHQoKY2'
    'FsbF9saW5rcxgDIAMoCVIJY2FsbExpbmtzEhgKB3N0cmVhbXMYBCABKANSB3N0cmVhbXMSNAoW'
    'c3RyZWFtc19wZXJfY3JlZGVudGlhbBgFIAEoA1IUc3RyZWFtc1BlckNyZWRlbnRpYWwSQgoIYW'
    'R2YW5jZWQYBiABKAsyJi53aXJlZ2xvdy5jb250cm9sLlByb2ZpbGVBZHZhbmNlZEVkaXRzUghh'
    'ZHZhbmNlZA==');

@$core.Deprecated('Use profileAdvancedEditsDescriptor instead')
const ProfileAdvancedEdits$json = {
  '1': 'ProfileAdvancedEdits',
  '2': [
    {'1': 'turn_transport', '3': 1, '4': 1, '5': 9, '10': 'turnTransport'},
    {
      '1': 'dns',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.DNSSettings',
      '10': 'dns'
    },
    {'1': 'manual_captcha', '3': 3, '4': 1, '5': 8, '10': 'manualCaptcha'},
    {'1': 'vk_platform', '3': 4, '4': 1, '5': 9, '10': 'vkPlatform'},
    {'1': 'turn_address', '3': 5, '4': 1, '5': 9, '10': 'turnAddress'},
    {'1': 'pacing_ms', '3': 6, '4': 1, '5': 3, '10': 'pacingMs'},
  ],
};

/// Descriptor for `ProfileAdvancedEdits`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileAdvancedEditsDescriptor = $convert.base64Decode(
    'ChRQcm9maWxlQWR2YW5jZWRFZGl0cxIlCg50dXJuX3RyYW5zcG9ydBgBIAEoCVINdHVyblRyYW'
    '5zcG9ydBIvCgNkbnMYAiABKAsyHS53aXJlZ2xvdy5jb250cm9sLkROU1NldHRpbmdzUgNkbnMS'
    'JQoObWFudWFsX2NhcHRjaGEYAyABKAhSDW1hbnVhbENhcHRjaGESHwoLdmtfcGxhdGZvcm0YBC'
    'ABKAlSCnZrUGxhdGZvcm0SIQoMdHVybl9hZGRyZXNzGAUgASgJUgt0dXJuQWRkcmVzcxIbCglw'
    'YWNpbmdfbXMYBiABKANSCHBhY2luZ01z');

@$core.Deprecated('Use importProfileRequestDescriptor instead')
const ImportProfileRequest$json = {
  '1': 'ImportProfileRequest',
  '2': [
    {'1': 'freeturn_link', '3': 1, '4': 1, '5': 9, '10': 'freeturnLink'},
  ],
  '9': [
    {'1': 2, '2': 3},
  ],
};

/// Descriptor for `ImportProfileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List importProfileRequestDescriptor = $convert.base64Decode(
    'ChRJbXBvcnRQcm9maWxlUmVxdWVzdBIjCg1mcmVldHVybl9saW5rGAEgASgJUgxmcmVldHVybk'
    'xpbmtKBAgCEAM=');

@$core.Deprecated('Use profileSummaryDescriptor instead')
const ProfileSummary$json = {
  '1': 'ProfileSummary',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'provider', '3': 3, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'peer', '3': 4, '4': 1, '5': 9, '10': 'peer'},
    {'1': 'turn_transport', '3': 5, '4': 1, '5': 9, '10': 'turnTransport'},
    {'1': 'relay_mode', '3': 6, '4': 1, '5': 9, '10': 'relayMode'},
    {'1': 'streams', '3': 7, '4': 1, '5': 3, '10': 'streams'},
    {
      '1': 'streams_per_credential',
      '3': 8,
      '4': 1,
      '5': 3,
      '10': 'streamsPerCredential'
    },
    {'1': 'obfuscation_mode', '3': 9, '4': 1, '5': 9, '10': 'obfuscationMode'},
  ],
};

/// Descriptor for `ProfileSummary`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileSummaryDescriptor = $convert.base64Decode(
    'Cg5Qcm9maWxlU3VtbWFyeRIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIaCg'
    'hwcm92aWRlchgDIAEoCVIIcHJvdmlkZXISEgoEcGVlchgEIAEoCVIEcGVlchIlCg50dXJuX3Ry'
    'YW5zcG9ydBgFIAEoCVINdHVyblRyYW5zcG9ydBIdCgpyZWxheV9tb2RlGAYgASgJUglyZWxheU'
    '1vZGUSGAoHc3RyZWFtcxgHIAEoA1IHc3RyZWFtcxI0ChZzdHJlYW1zX3Blcl9jcmVkZW50aWFs'
    'GAggASgDUhRzdHJlYW1zUGVyQ3JlZGVudGlhbBIpChBvYmZ1c2NhdGlvbl9tb2RlGAkgASgJUg'
    '9vYmZ1c2NhdGlvbk1vZGU=');

@$core.Deprecated('Use listProfilesResponseDescriptor instead')
const ListProfilesResponse$json = {
  '1': 'ListProfilesResponse',
  '2': [
    {
      '1': 'profiles',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.wireglow.control.ProfileSummary',
      '10': 'profiles'
    },
  ],
};

/// Descriptor for `ListProfilesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProfilesResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0UHJvZmlsZXNSZXNwb25zZRI8Cghwcm9maWxlcxgBIAMoCzIgLndpcmVnbG93LmNvbn'
    'Ryb2wuUHJvZmlsZVN1bW1hcnlSCHByb2ZpbGVz');

@$core.Deprecated('Use profileDescriptor instead')
const Profile$json = {
  '1': 'Profile',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {
      '1': 'connection',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.Connection',
      '10': 'connection'
    },
  ],
  '9': [
    {'1': 3, '2': 4},
  ],
};

/// Descriptor for `Profile`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileDescriptor = $convert.base64Decode(
    'CgdQcm9maWxlEg4KAmlkGAEgASgJUgJpZBI8Cgpjb25uZWN0aW9uGAIgASgLMhwud2lyZWdsb3'
    'cuY29udHJvbC5Db25uZWN0aW9uUgpjb25uZWN0aW9uSgQIAxAE');

@$core.Deprecated('Use connectionDescriptor instead')
const Connection$json = {
  '1': 'Connection',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'client_id', '3': 2, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'provider', '3': 3, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'turn_transport', '3': 4, '4': 1, '5': 9, '10': 'turnTransport'},
    {'1': 'relay_mode', '3': 5, '4': 1, '5': 9, '10': 'relayMode'},
    {
      '1': 'vk',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.VKSettings',
      '10': 'vk'
    },
    {
      '1': 'dns',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.DNSSettings',
      '10': 'dns'
    },
    {'1': 'streams', '3': 8, '4': 1, '5': 3, '10': 'streams'},
    {
      '1': 'streams_per_credential',
      '3': 9,
      '4': 1,
      '5': 3,
      '10': 'streamsPerCredential'
    },
    {'1': 'bond', '3': 10, '4': 1, '5': 8, '10': 'bond'},
    {
      '1': 'kcp',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.KCPSettings',
      '10': 'kcp'
    },
    {
      '1': 'wg',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.WireGuardSettings',
      '10': 'wg'
    },
    {'1': 'peer', '3': 13, '4': 1, '5': 9, '10': 'peer'},
    {
      '1': 'obfuscation',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.Obfuscation',
      '10': 'obfuscation'
    },
  ],
};

/// Descriptor for `Connection`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List connectionDescriptor = $convert.base64Decode(
    'CgpDb25uZWN0aW9uEhIKBG5hbWUYASABKAlSBG5hbWUSGwoJY2xpZW50X2lkGAIgASgJUghjbG'
    'llbnRJZBIaCghwcm92aWRlchgDIAEoCVIIcHJvdmlkZXISJQoOdHVybl90cmFuc3BvcnQYBCAB'
    'KAlSDXR1cm5UcmFuc3BvcnQSHQoKcmVsYXlfbW9kZRgFIAEoCVIJcmVsYXlNb2RlEiwKAnZrGA'
    'YgASgLMhwud2lyZWdsb3cuY29udHJvbC5WS1NldHRpbmdzUgJ2axIvCgNkbnMYByABKAsyHS53'
    'aXJlZ2xvdy5jb250cm9sLkROU1NldHRpbmdzUgNkbnMSGAoHc3RyZWFtcxgIIAEoA1IHc3RyZW'
    'FtcxI0ChZzdHJlYW1zX3Blcl9jcmVkZW50aWFsGAkgASgDUhRzdHJlYW1zUGVyQ3JlZGVudGlh'
    'bBISCgRib25kGAogASgIUgRib25kEi8KA2tjcBgLIAEoCzIdLndpcmVnbG93LmNvbnRyb2wuS0'
    'NQU2V0dGluZ3NSA2tjcBIzCgJ3ZxgMIAEoCzIjLndpcmVnbG93LmNvbnRyb2wuV2lyZUd1YXJk'
    'U2V0dGluZ3NSAndnEhIKBHBlZXIYDSABKAlSBHBlZXISPwoLb2JmdXNjYXRpb24YDiABKAsyHS'
    '53aXJlZ2xvdy5jb250cm9sLk9iZnVzY2F0aW9uUgtvYmZ1c2NhdGlvbg==');

@$core.Deprecated('Use vKSettingsDescriptor instead')
const VKSettings$json = {
  '1': 'VKSettings',
  '2': [
    {'1': 'call_links', '3': 1, '4': 3, '5': 9, '10': 'callLinks'},
    {'1': 'manual_captcha', '3': 2, '4': 1, '5': 8, '10': 'manualCaptcha'},
    {'1': 'platform', '3': 3, '4': 1, '5': 9, '10': 'platform'},
    {'1': 'turn_address', '3': 4, '4': 1, '5': 9, '10': 'turnAddress'},
  ],
};

/// Descriptor for `VKSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List vKSettingsDescriptor = $convert.base64Decode(
    'CgpWS1NldHRpbmdzEh0KCmNhbGxfbGlua3MYASADKAlSCWNhbGxMaW5rcxIlCg5tYW51YWxfY2'
    'FwdGNoYRgCIAEoCFINbWFudWFsQ2FwdGNoYRIaCghwbGF0Zm9ybRgDIAEoCVIIcGxhdGZvcm0S'
    'IQoMdHVybl9hZGRyZXNzGAQgASgJUgt0dXJuQWRkcmVzcw==');

@$core.Deprecated('Use dNSSettingsDescriptor instead')
const DNSSettings$json = {
  '1': 'DNSSettings',
  '2': [
    {'1': 'mode', '3': 1, '4': 1, '5': 9, '10': 'mode'},
    {'1': 'servers', '3': 2, '4': 3, '5': 9, '10': 'servers'},
  ],
};

/// Descriptor for `DNSSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List dNSSettingsDescriptor = $convert.base64Decode(
    'CgtETlNTZXR0aW5ncxISCgRtb2RlGAEgASgJUgRtb2RlEhgKB3NlcnZlcnMYAiADKAlSB3Nlcn'
    'ZlcnM=');

@$core.Deprecated('Use kCPSettingsDescriptor instead')
const KCPSettings$json = {
  '1': 'KCPSettings',
  '2': [
    {'1': 'no_delay', '3': 1, '4': 1, '5': 3, '10': 'noDelay'},
    {'1': 'interval', '3': 2, '4': 1, '5': 3, '10': 'interval'},
    {'1': 'resend', '3': 3, '4': 1, '5': 3, '10': 'resend'},
    {'1': 'nc', '3': 4, '4': 1, '5': 3, '10': 'nc'},
    {'1': 'snd_wnd', '3': 5, '4': 1, '5': 3, '10': 'sndWnd'},
    {'1': 'rcv_wnd', '3': 6, '4': 1, '5': 3, '10': 'rcvWnd'},
    {'1': 'mtu', '3': 7, '4': 1, '5': 3, '10': 'mtu'},
    {'1': 'ack_no_delay', '3': 8, '4': 1, '5': 8, '10': 'ackNoDelay'},
  ],
};

/// Descriptor for `KCPSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List kCPSettingsDescriptor = $convert.base64Decode(
    'CgtLQ1BTZXR0aW5ncxIZCghub19kZWxheRgBIAEoA1IHbm9EZWxheRIaCghpbnRlcnZhbBgCIA'
    'EoA1IIaW50ZXJ2YWwSFgoGcmVzZW5kGAMgASgDUgZyZXNlbmQSDgoCbmMYBCABKANSAm5jEhcK'
    'B3NuZF93bmQYBSABKANSBnNuZFduZBIXCgdyY3Zfd25kGAYgASgDUgZyY3ZXbmQSEAoDbXR1GA'
    'cgASgDUgNtdHUSIAoMYWNrX25vX2RlbGF5GAggASgIUgphY2tOb0RlbGF5');

@$core.Deprecated('Use wireGuardSettingsDescriptor instead')
const WireGuardSettings$json = {
  '1': 'WireGuardSettings',
  '2': [
    {'1': 'config', '3': 1, '4': 1, '5': 9, '10': 'config'},
  ],
};

/// Descriptor for `WireGuardSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List wireGuardSettingsDescriptor = $convert.base64Decode(
    'ChFXaXJlR3VhcmRTZXR0aW5ncxIWCgZjb25maWcYASABKAlSBmNvbmZpZw==');

@$core.Deprecated('Use obfuscationDescriptor instead')
const Obfuscation$json = {
  '1': 'Obfuscation',
  '2': [
    {'1': 'mode', '3': 1, '4': 1, '5': 9, '10': 'mode'},
    {'1': 'key', '3': 2, '4': 1, '5': 9, '10': 'key'},
    {'1': 'timing_ms', '3': 3, '4': 1, '5': 3, '10': 'timingMs'},
  ],
};

/// Descriptor for `Obfuscation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List obfuscationDescriptor = $convert.base64Decode(
    'CgtPYmZ1c2NhdGlvbhISCgRtb2RlGAEgASgJUgRtb2RlEhAKA2tleRgCIAEoCVIDa2V5EhsKCX'
    'RpbWluZ19tcxgDIAEoA1IIdGltaW5nTXM=');

@$core.Deprecated('Use snapshotDescriptor instead')
const Snapshot$json = {
  '1': 'Snapshot',
  '2': [
    {
      '1': 'active_mode',
      '3': 7,
      '4': 1,
      '5': 14,
      '6': '.wireglow.control.ProxyMode',
      '10': 'activeMode'
    },
    {
      '1': 'state',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.wireglow.control.State',
      '10': 'state'
    },
    {'1': 'generation', '3': 2, '4': 1, '5': 4, '10': 'generation'},
    {'1': 'profile_id', '3': 3, '4': 1, '5': 9, '10': 'profileId'},
    {'1': 'mixed_address', '3': 4, '4': 1, '5': 9, '10': 'mixedAddress'},
    {'1': 'error', '3': 5, '4': 1, '5': 9, '10': 'error'},
    {
      '1': 'latency',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.Latency',
      '10': 'latency'
    },
  ],
};

/// Descriptor for `Snapshot`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List snapshotDescriptor = $convert.base64Decode(
    'CghTbmFwc2hvdBI8CgthY3RpdmVfbW9kZRgHIAEoDjIbLndpcmVnbG93LmNvbnRyb2wuUHJveH'
    'lNb2RlUgphY3RpdmVNb2RlEi0KBXN0YXRlGAEgASgOMhcud2lyZWdsb3cuY29udHJvbC5TdGF0'
    'ZVIFc3RhdGUSHgoKZ2VuZXJhdGlvbhgCIAEoBFIKZ2VuZXJhdGlvbhIdCgpwcm9maWxlX2lkGA'
    'MgASgJUglwcm9maWxlSWQSIwoNbWl4ZWRfYWRkcmVzcxgEIAEoCVIMbWl4ZWRBZGRyZXNzEhQK'
    'BWVycm9yGAUgASgJUgVlcnJvchIzCgdsYXRlbmN5GAYgASgLMhkud2lyZWdsb3cuY29udHJvbC'
    '5MYXRlbmN5UgdsYXRlbmN5');

@$core.Deprecated('Use latencyDescriptor instead')
const Latency$json = {
  '1': 'Latency',
  '2': [
    {'1': 'checking', '3': 1, '4': 1, '5': 8, '10': 'checking'},
    {'1': 'delay_ms', '3': 2, '4': 1, '5': 3, '10': 'delayMs'},
    {'1': 'checked_at', '3': 3, '4': 1, '5': 3, '10': 'checkedAt'},
    {'1': 'success_at', '3': 4, '4': 1, '5': 3, '10': 'successAt'},
    {'1': 'error', '3': 5, '4': 1, '5': 9, '10': 'error'},
  ],
};

/// Descriptor for `Latency`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List latencyDescriptor = $convert.base64Decode(
    'CgdMYXRlbmN5EhoKCGNoZWNraW5nGAEgASgIUghjaGVja2luZxIZCghkZWxheV9tcxgCIAEoA1'
    'IHZGVsYXlNcxIdCgpjaGVja2VkX2F0GAMgASgDUgljaGVja2VkQXQSHQoKc3VjY2Vzc19hdBgE'
    'IAEoA1IJc3VjY2Vzc0F0EhQKBWVycm9yGAUgASgJUgVlcnJvcg==');

@$core.Deprecated('Use exportProfilesRequestDescriptor instead')
const ExportProfilesRequest$json = {
  '1': 'ExportProfilesRequest',
  '2': [
    {'1': 'ids', '3': 1, '4': 3, '5': 9, '10': 'ids'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `ExportProfilesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List exportProfilesRequestDescriptor = $convert.base64Decode(
    'ChVFeHBvcnRQcm9maWxlc1JlcXVlc3QSEAoDaWRzGAEgAygJUgNpZHMSEgoEcGF0aBgCIAEoCV'
    'IEcGF0aA==');

@$core.Deprecated('Use profileFileRequestDescriptor instead')
const ProfileFileRequest$json = {
  '1': 'ProfileFileRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `ProfileFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileFileRequestDescriptor = $convert
    .base64Decode('ChJQcm9maWxlRmlsZVJlcXVlc3QSEgoEcGF0aBgBIAEoCVIEcGF0aA==');

@$core.Deprecated('Use profileFileEntryDescriptor instead')
const ProfileFileEntry$json = {
  '1': 'ProfileFileEntry',
  '2': [
    {
      '1': 'profile',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.ProfileSummary',
      '10': 'profile'
    },
    {
      '1': 'existing',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.wireglow.control.ProfileSummary',
      '10': 'existing'
    },
  ],
};

/// Descriptor for `ProfileFileEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileFileEntryDescriptor = $convert.base64Decode(
    'ChBQcm9maWxlRmlsZUVudHJ5EjoKB3Byb2ZpbGUYASABKAsyIC53aXJlZ2xvdy5jb250cm9sLl'
    'Byb2ZpbGVTdW1tYXJ5Ugdwcm9maWxlEjwKCGV4aXN0aW5nGAIgASgLMiAud2lyZWdsb3cuY29u'
    'dHJvbC5Qcm9maWxlU3VtbWFyeVIIZXhpc3Rpbmc=');

@$core.Deprecated('Use profileFilePreviewDescriptor instead')
const ProfileFilePreview$json = {
  '1': 'ProfileFilePreview',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {
      '1': 'entries',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.wireglow.control.ProfileFileEntry',
      '10': 'entries'
    },
  ],
};

/// Descriptor for `ProfileFilePreview`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileFilePreviewDescriptor = $convert.base64Decode(
    'ChJQcm9maWxlRmlsZVByZXZpZXcSFAoFdG9rZW4YASABKAlSBXRva2VuEjwKB2VudHJpZXMYAi'
    'ADKAsyIi53aXJlZ2xvdy5jb250cm9sLlByb2ZpbGVGaWxlRW50cnlSB2VudHJpZXM=');

@$core.Deprecated('Use profileImportChoiceDescriptor instead')
const ProfileImportChoice$json = {
  '1': 'ProfileImportChoice',
  '2': [
    {'1': 'index', '3': 1, '4': 1, '5': 13, '10': 'index'},
    {
      '1': 'action',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.wireglow.control.ProfileImportAction',
      '10': 'action'
    },
  ],
};

/// Descriptor for `ProfileImportChoice`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileImportChoiceDescriptor = $convert.base64Decode(
    'ChNQcm9maWxlSW1wb3J0Q2hvaWNlEhQKBWluZGV4GAEgASgNUgVpbmRleBI9CgZhY3Rpb24YAi'
    'ABKA4yJS53aXJlZ2xvdy5jb250cm9sLlByb2ZpbGVJbXBvcnRBY3Rpb25SBmFjdGlvbg==');

@$core.Deprecated('Use importProfileFileRequestDescriptor instead')
const ImportProfileFileRequest$json = {
  '1': 'ImportProfileFileRequest',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {
      '1': 'choices',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.wireglow.control.ProfileImportChoice',
      '10': 'choices'
    },
  ],
};

/// Descriptor for `ImportProfileFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List importProfileFileRequestDescriptor = $convert.base64Decode(
    'ChhJbXBvcnRQcm9maWxlRmlsZVJlcXVlc3QSFAoFdG9rZW4YASABKAlSBXRva2VuEj8KB2Nob2'
    'ljZXMYAiADKAsyJS53aXJlZ2xvdy5jb250cm9sLlByb2ZpbGVJbXBvcnRDaG9pY2VSB2Nob2lj'
    'ZXM=');

@$core.Deprecated('Use profileImportResultDescriptor instead')
const ProfileImportResult$json = {
  '1': 'ProfileImportResult',
  '2': [
    {'1': 'index', '3': 1, '4': 1, '5': 13, '10': 'index'},
    {'1': 'status', '3': 2, '4': 1, '5': 9, '10': 'status'},
  ],
};

/// Descriptor for `ProfileImportResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileImportResultDescriptor = $convert.base64Decode(
    'ChNQcm9maWxlSW1wb3J0UmVzdWx0EhQKBWluZGV4GAEgASgNUgVpbmRleBIWCgZzdGF0dXMYAi'
    'ABKAlSBnN0YXR1cw==');

@$core.Deprecated('Use importProfileFileResponseDescriptor instead')
const ImportProfileFileResponse$json = {
  '1': 'ImportProfileFileResponse',
  '2': [
    {
      '1': 'results',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.wireglow.control.ProfileImportResult',
      '10': 'results'
    },
  ],
};

/// Descriptor for `ImportProfileFileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List importProfileFileResponseDescriptor =
    $convert.base64Decode(
        'ChlJbXBvcnRQcm9maWxlRmlsZVJlc3BvbnNlEj8KB3Jlc3VsdHMYASADKAsyJS53aXJlZ2xvdy'
        '5jb250cm9sLlByb2ZpbGVJbXBvcnRSZXN1bHRSB3Jlc3VsdHM=');
