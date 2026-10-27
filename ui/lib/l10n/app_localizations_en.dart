// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get notSet => 'Not set';

  @override
  String get starting => 'Starting controller…';

  @override
  String get unknownState => 'Connection state unknown';

  @override
  String get disconnecting => 'Disconnecting…';

  @override
  String get active => 'Connection active';

  @override
  String get connecting => 'Connecting…';

  @override
  String get failed => 'Connection failed';

  @override
  String get addToConnect => 'Add a link to connect';

  @override
  String get pressToConnect => 'Press the button to connect';

  @override
  String get disconnect => 'Disconnect';

  @override
  String get cancelConnection => 'Cancel connection';

  @override
  String get retry => 'Try again';

  @override
  String get connect => 'Connect';

  @override
  String get connections => 'Connections';

  @override
  String get about => 'About';

  @override
  String get profiles => 'Profiles';

  @override
  String get addByLink => 'Add from link';

  @override
  String get unnamed => 'Unnamed';

  @override
  String get connected => 'Connected';

  @override
  String get savedProfile => 'Saved profile';

  @override
  String get deleteProfile => 'Delete profile';

  @override
  String get checkingLatency => 'Checking latency…';

  @override
  String get checkFailed => 'Check failed';

  @override
  String get latencyNotMeasured => 'Latency not measured yet';

  @override
  String get latencyHelp =>
      'HTTPS request time through the connection. Checked every 30 seconds.';

  @override
  String get checkLatency => 'Check latency';

  @override
  String get noConnections => 'No connections';

  @override
  String get retryHelp =>
      'Check connection settings and network access, then try again.';

  @override
  String get errorDetails => 'Error details';

  @override
  String get noErrorDetails => 'Controller did not provide details.';

  @override
  String get properties => 'Connection properties';

  @override
  String get systemProxy =>
      'System proxy: used by applications that follow system proxy settings.';

  @override
  String get localProxyHelp =>
      'Local proxy: enter the displayed address in your application. System settings are unchanged.';

  @override
  String get provider => 'Provider';

  @override
  String get turnTransport => 'TURN transport';

  @override
  String get notApplicable => 'Not applicable';

  @override
  String get relayMode => 'Relay mode';

  @override
  String get streams => 'Streams';

  @override
  String get streamsPerCredential => 'Streams per credential';

  @override
  String get obfuscation => 'Obfuscation';

  @override
  String get confirmDelete => 'Delete profile?';

  @override
  String get cancel => 'Cancel';

  @override
  String get delete => 'Delete';

  @override
  String get insertLink => 'Paste a connection link.';

  @override
  String get addProfile => 'Add profile';

  @override
  String get closeAdd => 'Close add profile';

  @override
  String get pasteHelp => 'Paste the link you received.';

  @override
  String get connectionLink => 'Connection link';

  @override
  String get paste => 'Paste';

  @override
  String get nameFromLink => 'The name comes from the link.';

  @override
  String get add => 'Add';

  @override
  String get browserFailed => 'Could not open the link in your browser';

  @override
  String get openGithub => 'Open on GitHub';

  @override
  String get repositoryMissing => 'Repository was not specified at build time';

  @override
  String get clientVersion => 'Version';

  @override
  String get commit => 'Commit';

  @override
  String get tag => 'Tag';

  @override
  String get buildTime => 'Build time';

  @override
  String get engines => 'Engines';

  @override
  String get unknown => 'unknown';

  @override
  String get language => 'Language';

  @override
  String get systemLanguage => 'System';

  @override
  String get lostController =>
      'Connection to controller lost. Connection state unknown.';

  @override
  String get waitController => 'Disconnect first or wait for the controller.';

  @override
  String get startFailed => 'Could not start controller.';

  @override
  String get incompleteBundle =>
      'Application files are missing. Extract the complete application folder again.';

  @override
  String get controllerFailed =>
      'Controller did not start. Check access to the data folder and make sure no other client is running.';

  @override
  String get settingsFailed => 'Could not save the language preference.';

  @override
  String latency(String delay) {
    return 'Latency: $delay ms';
  }

  @override
  String lastSuccess(String time) {
    return 'Last success: $time';
  }

  @override
  String localProxy(String address) {
    return 'Local HTTP/SOCKS proxy: $address';
  }

  @override
  String deleteDescription(String name) {
    return 'Profile “$name” will be deleted. You will need its connection link to add it again.';
  }

  @override
  String get proxyMode => 'Proxy mode';

  @override
  String get localProxyMode => 'Local proxy';

  @override
  String get systemProxyMode => 'System proxy';

  @override
  String get copyProxyAddress => 'Copy proxy address';

  @override
  String get switchingProxyMode => 'Switching proxy mode…';

  @override
  String get proxyModeUnknown => 'Proxy mode unknown';

  @override
  String get clientSettings => 'Settings';

  @override
  String get localAddresses => 'Local addresses';

  @override
  String get proxyAddressLabel => 'HTTP/SOCKS proxy';

  @override
  String get relayAddressLabel => 'FreeTurn relay';

  @override
  String get addressesHelp =>
      'Address format: 127.0.0.1:port. All ports must be between 1 and 65535 and must differ. Changes apply to all profiles on the next connection.';

  @override
  String get activeSettingsHelp =>
      'The active connection keeps its current addresses until you reconnect.';

  @override
  String get invalidLocalAddress =>
      'Enter 127.0.0.1 and a port from 1 to 65535.';

  @override
  String get collidingPorts => 'Ports must differ.';

  @override
  String get saveSettings => 'Save';

  @override
  String get settingsSaved => 'Settings saved';

  @override
  String get resetSettings => 'Reset to defaults';

  @override
  String get confirmResetSettings => 'Reset client settings?';

  @override
  String get resetSettingsDescription =>
      'Restore default addresses, local proxy mode and system language. Saved profiles will be kept. The active connection keeps its addresses until you reconnect.';

  @override
  String get reset => 'Reset';

  @override
  String get settingsReset => 'Default settings restored';

  @override
  String get generalSettings => 'General';

  @override
  String get networkSettings => 'Network';

  @override
  String get relaySettings => 'Relay';

  @override
  String get apiPortLabel => 'mihomo API port';

  @override
  String get invalidAPIPort => 'Enter a port between 1 and 65535.';

  @override
  String get profileActions => 'Profile actions';

  @override
  String get editProfile => 'Edit';

  @override
  String get colorTheme => 'Color theme';

  @override
  String get systemTheme => 'System';

  @override
  String get darkTheme => 'Dark';

  @override
  String get lightTheme => 'Light';

  @override
  String get tunMode => 'TUN';

  @override
  String get tunHelpWindows =>
      'Experimental Windows TUN requires the portable bundle and UAC elevation. It reconnects the session and uses direct HTTPS bootstrap DNS. Other active VPNs block activation. For TURN, select TCP transport; application UDP traffic is supported.';

  @override
  String get tunHelp =>
      'TUN captures IP traffic. Enabling it reconnects the session and uses 1.1.1.1 as direct bootstrap DNS. Other active VPNs block activation.';

  @override
  String get editorTitle => 'Edit profile';

  @override
  String get editorProfile => 'Profile';

  @override
  String get editorName => 'Profile name';

  @override
  String get editorNameRequired => 'Enter a profile name.';

  @override
  String get editorVK => 'VK Calls';

  @override
  String get editorCallLinks => 'Call links';

  @override
  String editorCallLink(int number) {
    return 'Call link $number';
  }

  @override
  String get editorRemoveLink => 'Remove call link';

  @override
  String get editorAddLink => 'Add call link';

  @override
  String get editorPerformance => 'Performance';

  @override
  String get editorStreams => 'Streams per call';

  @override
  String get editorCredentials => 'Streams per TURN credential set';

  @override
  String get editorStreamsHelp => 'Parallel TURN sessions for each call link.';

  @override
  String get editorCredentialsHelp =>
      'Sessions sharing one set of credentials from VK.';

  @override
  String get editorAdvanced => 'Advanced settings';

  @override
  String get editorVKAuth => 'VK authorization';

  @override
  String get editorPlatform => 'Device type';

  @override
  String get editorCaptcha => 'Manual CAPTCHA';

  @override
  String get editorCaptchaHelp =>
      'Solve in a browser instead of automatically.';

  @override
  String get editorTransport => 'Transport';

  @override
  String get editorUDPHelp =>
      'UDP transport is unavailable in Windows TUN mode.';

  @override
  String get editorTurnAddress => 'Node address';

  @override
  String get editorTurnHint => 'Host or host:port';

  @override
  String get editorTurnHelp =>
      'Leave empty to use the node from VK. Without a port, the port from VK is used. For IPv6 with a port, use [address]:port.';

  @override
  String get editorDNSMode => 'DNS mode';

  @override
  String get editorDNSAuto => 'Automatic';

  @override
  String get editorDNSPlain => 'Plain DNS';

  @override
  String get editorDNSServers => 'Custom DNS servers';

  @override
  String get editorDNSHelp =>
      'DNS for FreeTurn connections. Enter IP[:port] addresses separated by commas.';

  @override
  String get editorDoHHelp =>
      'DoH uses the engine’s built-in servers. Custom DoH URLs are not supported.';

  @override
  String get editorTunDNSHelp =>
      'TUN mode uses DNS selected by the application. Profile DNS applies in local and system proxy modes.';

  @override
  String get editorObfuscation => 'Obfuscation';

  @override
  String get editorObfuscationOff => 'Disabled in this profile';

  @override
  String get editorPacing => 'Pacing, ms';

  @override
  String get editorPacingHelp =>
      'Interval between outgoing packets. 0 disables pacing.';

  @override
  String get editorPacingDisabled =>
      'Available when obfuscation is enabled in the imported profile.';

  @override
  String get editorReconnect => 'Changes apply on the next connection.';

  @override
  String editorDefault(String value) {
    return 'Default ($value)';
  }

  @override
  String get editorInvalidStreams =>
      'Enter 1–1024, or leave empty for the default.';

  @override
  String get editorInvalidLinks =>
      'Enter at least one call link, without commas.';

  @override
  String get editorInvalidPacing => 'Enter 0–60000 milliseconds.';

  @override
  String get importProfiles => 'Import profiles';

  @override
  String get exportProfiles => 'Export profiles';

  @override
  String get importProfilesAction => 'Import…';

  @override
  String get exportProfilesAction => 'Export…';

  @override
  String get profileTransferHelp =>
      'Transfer connection profiles between devices or keep a backup.';

  @override
  String get importProfilesSettingsHelp =>
      'Review a Wireglow file and choose what to add.';

  @override
  String get exportProfilesSettingsHelp =>
      'Save all profiles or select individual ones.';

  @override
  String get importProfilesHelp =>
      'Restore profiles from a Wireglow export. Review the file before importing.';

  @override
  String get exportProfilesHelp => 'Choose the profiles to save to a file.';

  @override
  String get profileFileType => 'Wireglow profiles file (.json)';

  @override
  String get chooseProfileFile => 'Choose file…';

  @override
  String get changeProfileFile => 'Change';

  @override
  String get selectAllProfiles => 'Select all';

  @override
  String get profileExportSecretsHelp =>
      'Includes connection settings, keys and call links. Anyone with this file can use these connections.';

  @override
  String get profileImportReviewHelp =>
      'Existing profiles are matched by ID. Replacing a profile applies on the next connection.';

  @override
  String get skipExistingProfile => 'Skip existing';

  @override
  String get copyImportedProfile => 'Create a copy';

  @override
  String get replaceExistingProfile => 'Replace existing';

  @override
  String get profileArchiveInvalid =>
      'Invalid Wireglow file, unsupported version or file too large (maximum 16 MiB and 1000 profiles).';

  @override
  String get profilePreviewExpired =>
      'Preview expired or a session operation is running. Open the file again and review the profiles.';

  @override
  String get profileTransferFailed =>
      'Could not complete the operation. Check file permissions and available disk space, then choose the file again.';

  @override
  String get profileImportPartial =>
      'Some profiles were not imported. Saved profiles are listed below. Choose the file again to retry.';

  @override
  String get profileImportDone => 'Imported';

  @override
  String get profileImportSkipped => 'Skipped';

  @override
  String get profileImportChanged => 'Changed after preview; not imported';

  @override
  String get profileImportFailed => 'Could not save this profile';

  @override
  String get closeProfileTransfer => 'Close';

  @override
  String profileCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count profiles',
      one: '1 profile',
      zero: 'No profiles selected',
    );
    return '$_temp0';
  }

  @override
  String exportProfileCount(int count) {
    return 'Export $count';
  }

  @override
  String importProfileCount(int count) {
    return 'Import $count';
  }

  @override
  String profilesExported(int count) {
    return 'Profiles exported: $count';
  }

  @override
  String profilesImported(int count) {
    return 'Profiles imported: $count';
  }

  @override
  String profileAlreadyExists(String name) {
    return 'Already saved as “$name”';
  }
}
