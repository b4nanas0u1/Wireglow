import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_ru.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('ru'),
  ];

  /// No description provided for @notSet.
  ///
  /// In en, this message translates to:
  /// **'Not set'**
  String get notSet;

  /// No description provided for @starting.
  ///
  /// In en, this message translates to:
  /// **'Starting controller…'**
  String get starting;

  /// No description provided for @unknownState.
  ///
  /// In en, this message translates to:
  /// **'Connection state unknown'**
  String get unknownState;

  /// No description provided for @disconnecting.
  ///
  /// In en, this message translates to:
  /// **'Disconnecting…'**
  String get disconnecting;

  /// No description provided for @active.
  ///
  /// In en, this message translates to:
  /// **'Connection active'**
  String get active;

  /// No description provided for @connecting.
  ///
  /// In en, this message translates to:
  /// **'Connecting…'**
  String get connecting;

  /// No description provided for @failed.
  ///
  /// In en, this message translates to:
  /// **'Connection failed'**
  String get failed;

  /// No description provided for @addToConnect.
  ///
  /// In en, this message translates to:
  /// **'Add a link to connect'**
  String get addToConnect;

  /// No description provided for @pressToConnect.
  ///
  /// In en, this message translates to:
  /// **'Press the button to connect'**
  String get pressToConnect;

  /// No description provided for @disconnect.
  ///
  /// In en, this message translates to:
  /// **'Disconnect'**
  String get disconnect;

  /// No description provided for @cancelConnection.
  ///
  /// In en, this message translates to:
  /// **'Cancel connection'**
  String get cancelConnection;

  /// No description provided for @retry.
  ///
  /// In en, this message translates to:
  /// **'Try again'**
  String get retry;

  /// No description provided for @connect.
  ///
  /// In en, this message translates to:
  /// **'Connect'**
  String get connect;

  /// No description provided for @connections.
  ///
  /// In en, this message translates to:
  /// **'Connections'**
  String get connections;

  /// No description provided for @about.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get about;

  /// No description provided for @profiles.
  ///
  /// In en, this message translates to:
  /// **'Profiles'**
  String get profiles;

  /// No description provided for @addByLink.
  ///
  /// In en, this message translates to:
  /// **'Add from link'**
  String get addByLink;

  /// No description provided for @unnamed.
  ///
  /// In en, this message translates to:
  /// **'Unnamed'**
  String get unnamed;

  /// No description provided for @connected.
  ///
  /// In en, this message translates to:
  /// **'Connected'**
  String get connected;

  /// No description provided for @savedProfile.
  ///
  /// In en, this message translates to:
  /// **'Saved profile'**
  String get savedProfile;

  /// No description provided for @deleteProfile.
  ///
  /// In en, this message translates to:
  /// **'Delete profile'**
  String get deleteProfile;

  /// No description provided for @checkingLatency.
  ///
  /// In en, this message translates to:
  /// **'Checking latency…'**
  String get checkingLatency;

  /// No description provided for @checkFailed.
  ///
  /// In en, this message translates to:
  /// **'Check failed'**
  String get checkFailed;

  /// No description provided for @latencyNotMeasured.
  ///
  /// In en, this message translates to:
  /// **'Latency not measured yet'**
  String get latencyNotMeasured;

  /// No description provided for @latencyHelp.
  ///
  /// In en, this message translates to:
  /// **'HTTPS request time through the connection. Checked every 30 seconds.'**
  String get latencyHelp;

  /// No description provided for @checkLatency.
  ///
  /// In en, this message translates to:
  /// **'Check latency'**
  String get checkLatency;

  /// No description provided for @noConnections.
  ///
  /// In en, this message translates to:
  /// **'No connections'**
  String get noConnections;

  /// No description provided for @retryHelp.
  ///
  /// In en, this message translates to:
  /// **'Check connection settings and network access, then try again.'**
  String get retryHelp;

  /// No description provided for @errorDetails.
  ///
  /// In en, this message translates to:
  /// **'Error details'**
  String get errorDetails;

  /// No description provided for @noErrorDetails.
  ///
  /// In en, this message translates to:
  /// **'Controller did not provide details.'**
  String get noErrorDetails;

  /// No description provided for @properties.
  ///
  /// In en, this message translates to:
  /// **'Connection properties'**
  String get properties;

  /// No description provided for @systemProxy.
  ///
  /// In en, this message translates to:
  /// **'System proxy: used by applications that follow system proxy settings.'**
  String get systemProxy;

  /// No description provided for @localProxyHelp.
  ///
  /// In en, this message translates to:
  /// **'Local proxy: enter the displayed address in your application. System settings are unchanged.'**
  String get localProxyHelp;

  /// No description provided for @provider.
  ///
  /// In en, this message translates to:
  /// **'Provider'**
  String get provider;

  /// No description provided for @turnTransport.
  ///
  /// In en, this message translates to:
  /// **'TURN transport'**
  String get turnTransport;

  /// No description provided for @notApplicable.
  ///
  /// In en, this message translates to:
  /// **'Not applicable'**
  String get notApplicable;

  /// No description provided for @relayMode.
  ///
  /// In en, this message translates to:
  /// **'Relay mode'**
  String get relayMode;

  /// No description provided for @streams.
  ///
  /// In en, this message translates to:
  /// **'Streams'**
  String get streams;

  /// No description provided for @streamsPerCredential.
  ///
  /// In en, this message translates to:
  /// **'Streams per credential'**
  String get streamsPerCredential;

  /// No description provided for @obfuscation.
  ///
  /// In en, this message translates to:
  /// **'Obfuscation'**
  String get obfuscation;

  /// No description provided for @confirmDelete.
  ///
  /// In en, this message translates to:
  /// **'Delete profile?'**
  String get confirmDelete;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @delete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get delete;

  /// No description provided for @insertLink.
  ///
  /// In en, this message translates to:
  /// **'Paste a connection link.'**
  String get insertLink;

  /// No description provided for @addProfile.
  ///
  /// In en, this message translates to:
  /// **'Add profile'**
  String get addProfile;

  /// No description provided for @closeAdd.
  ///
  /// In en, this message translates to:
  /// **'Close add profile'**
  String get closeAdd;

  /// No description provided for @pasteHelp.
  ///
  /// In en, this message translates to:
  /// **'Paste the link you received.'**
  String get pasteHelp;

  /// No description provided for @connectionLink.
  ///
  /// In en, this message translates to:
  /// **'Connection link'**
  String get connectionLink;

  /// No description provided for @paste.
  ///
  /// In en, this message translates to:
  /// **'Paste'**
  String get paste;

  /// No description provided for @nameFromLink.
  ///
  /// In en, this message translates to:
  /// **'The name comes from the link.'**
  String get nameFromLink;

  /// No description provided for @add.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get add;

  /// No description provided for @browserFailed.
  ///
  /// In en, this message translates to:
  /// **'Could not open the link in your browser'**
  String get browserFailed;

  /// No description provided for @openGithub.
  ///
  /// In en, this message translates to:
  /// **'Open on GitHub'**
  String get openGithub;

  /// No description provided for @repositoryMissing.
  ///
  /// In en, this message translates to:
  /// **'Repository was not specified at build time'**
  String get repositoryMissing;

  /// No description provided for @clientVersion.
  ///
  /// In en, this message translates to:
  /// **'Client version'**
  String get clientVersion;

  /// No description provided for @commit.
  ///
  /// In en, this message translates to:
  /// **'Commit'**
  String get commit;

  /// No description provided for @tag.
  ///
  /// In en, this message translates to:
  /// **'Tag'**
  String get tag;

  /// No description provided for @buildTime.
  ///
  /// In en, this message translates to:
  /// **'Build time'**
  String get buildTime;

  /// No description provided for @engines.
  ///
  /// In en, this message translates to:
  /// **'Engines'**
  String get engines;

  /// No description provided for @unknown.
  ///
  /// In en, this message translates to:
  /// **'unknown'**
  String get unknown;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// No description provided for @systemLanguage.
  ///
  /// In en, this message translates to:
  /// **'System'**
  String get systemLanguage;

  /// No description provided for @lostController.
  ///
  /// In en, this message translates to:
  /// **'Connection to controller lost. Connection state unknown.'**
  String get lostController;

  /// No description provided for @waitController.
  ///
  /// In en, this message translates to:
  /// **'Disconnect first or wait for the controller.'**
  String get waitController;

  /// No description provided for @startFailed.
  ///
  /// In en, this message translates to:
  /// **'Could not start controller.'**
  String get startFailed;

  /// No description provided for @incompleteBundle.
  ///
  /// In en, this message translates to:
  /// **'Application files are missing. Extract the complete application folder again.'**
  String get incompleteBundle;

  /// No description provided for @controllerFailed.
  ///
  /// In en, this message translates to:
  /// **'Controller did not start. Check systemd --user and make sure no other client is running.'**
  String get controllerFailed;

  /// No description provided for @settingsFailed.
  ///
  /// In en, this message translates to:
  /// **'Could not save the language preference.'**
  String get settingsFailed;

  /// No description provided for @latency.
  ///
  /// In en, this message translates to:
  /// **'Latency: {delay} ms'**
  String latency(String delay);

  /// No description provided for @lastSuccess.
  ///
  /// In en, this message translates to:
  /// **'Last success: {time}'**
  String lastSuccess(String time);

  /// No description provided for @localProxy.
  ///
  /// In en, this message translates to:
  /// **'Local HTTP/SOCKS proxy: {address}'**
  String localProxy(String address);

  /// No description provided for @deleteDescription.
  ///
  /// In en, this message translates to:
  /// **'Profile “{name}” will be deleted. You will need its connection link to add it again.'**
  String deleteDescription(String name);

  /// No description provided for @proxyMode.
  ///
  /// In en, this message translates to:
  /// **'Proxy mode'**
  String get proxyMode;

  /// No description provided for @localProxyMode.
  ///
  /// In en, this message translates to:
  /// **'Local proxy'**
  String get localProxyMode;

  /// No description provided for @systemProxyMode.
  ///
  /// In en, this message translates to:
  /// **'System proxy'**
  String get systemProxyMode;

  /// No description provided for @copyProxyAddress.
  ///
  /// In en, this message translates to:
  /// **'Copy proxy address'**
  String get copyProxyAddress;

  /// No description provided for @switchingProxyMode.
  ///
  /// In en, this message translates to:
  /// **'Switching proxy mode…'**
  String get switchingProxyMode;

  /// No description provided for @proxyModeUnknown.
  ///
  /// In en, this message translates to:
  /// **'Proxy mode unknown'**
  String get proxyModeUnknown;

  /// No description provided for @clientSettings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get clientSettings;

  /// No description provided for @localAddresses.
  ///
  /// In en, this message translates to:
  /// **'Local addresses'**
  String get localAddresses;

  /// No description provided for @proxyAddressLabel.
  ///
  /// In en, this message translates to:
  /// **'HTTP/SOCKS proxy'**
  String get proxyAddressLabel;

  /// No description provided for @relayAddressLabel.
  ///
  /// In en, this message translates to:
  /// **'FreeTurn relay'**
  String get relayAddressLabel;

  /// No description provided for @addressesHelp.
  ///
  /// In en, this message translates to:
  /// **'Address format: 127.0.0.1:port. All ports must be between 1 and 65535 and must differ. Changes apply to all profiles on the next connection.'**
  String get addressesHelp;

  /// No description provided for @activeSettingsHelp.
  ///
  /// In en, this message translates to:
  /// **'The active connection keeps its current addresses until you reconnect.'**
  String get activeSettingsHelp;

  /// No description provided for @invalidLocalAddress.
  ///
  /// In en, this message translates to:
  /// **'Enter 127.0.0.1 and a port from 1 to 65535.'**
  String get invalidLocalAddress;

  /// No description provided for @collidingPorts.
  ///
  /// In en, this message translates to:
  /// **'Ports must differ.'**
  String get collidingPorts;

  /// No description provided for @saveSettings.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get saveSettings;

  /// No description provided for @settingsSaved.
  ///
  /// In en, this message translates to:
  /// **'Settings saved'**
  String get settingsSaved;

  /// No description provided for @resetSettings.
  ///
  /// In en, this message translates to:
  /// **'Reset to defaults'**
  String get resetSettings;

  /// No description provided for @confirmResetSettings.
  ///
  /// In en, this message translates to:
  /// **'Reset client settings?'**
  String get confirmResetSettings;

  /// No description provided for @resetSettingsDescription.
  ///
  /// In en, this message translates to:
  /// **'Restore default addresses, local proxy mode and system language. Saved profiles will be kept. The active connection keeps its addresses until you reconnect.'**
  String get resetSettingsDescription;

  /// No description provided for @reset.
  ///
  /// In en, this message translates to:
  /// **'Reset'**
  String get reset;

  /// No description provided for @settingsReset.
  ///
  /// In en, this message translates to:
  /// **'Default settings restored'**
  String get settingsReset;

  /// No description provided for @generalSettings.
  ///
  /// In en, this message translates to:
  /// **'General'**
  String get generalSettings;

  /// No description provided for @networkSettings.
  ///
  /// In en, this message translates to:
  /// **'Network'**
  String get networkSettings;

  /// No description provided for @relaySettings.
  ///
  /// In en, this message translates to:
  /// **'Relay'**
  String get relaySettings;

  /// No description provided for @apiPortLabel.
  ///
  /// In en, this message translates to:
  /// **'mihomo API port'**
  String get apiPortLabel;

  /// No description provided for @invalidAPIPort.
  ///
  /// In en, this message translates to:
  /// **'Enter a port between 1 and 65535.'**
  String get invalidAPIPort;

  /// No description provided for @profileActions.
  ///
  /// In en, this message translates to:
  /// **'Profile actions'**
  String get profileActions;

  /// No description provided for @editProfile.
  ///
  /// In en, this message translates to:
  /// **'Edit'**
  String get editProfile;

  /// No description provided for @colorTheme.
  ///
  /// In en, this message translates to:
  /// **'Color theme'**
  String get colorTheme;

  /// No description provided for @systemTheme.
  ///
  /// In en, this message translates to:
  /// **'System'**
  String get systemTheme;

  /// No description provided for @darkTheme.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get darkTheme;

  /// No description provided for @lightTheme.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get lightTheme;

  /// No description provided for @tunMode.
  ///
  /// In en, this message translates to:
  /// **'TUN'**
  String get tunMode;

  /// No description provided for @tunHelp.
  ///
  /// In en, this message translates to:
  /// **'TUN captures IP traffic. Enabling it reconnects the session and uses 1.1.1.1 as direct bootstrap DNS. Other active VPNs block activation.'**
  String get tunHelp;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'ru'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'ru':
      return AppLocalizationsRu();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
