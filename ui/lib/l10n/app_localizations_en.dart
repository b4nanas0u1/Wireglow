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
  String get clientVersion => 'Client version';

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
      'Controller did not start. Check systemd --user and make sure no other client is running.';

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
}
