import 'checked_popup_menu.dart';

import 'dart:async';
import 'dart:ui' show AppExitResponse;

import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, TargetPlatform;
import 'package:flutter/services.dart';

import 'app_info.dart';
import 'l10n/app_localizations.dart';
import 'ui_error.dart';
import 'ui_error_text.dart';
import 'language_settings.dart';
import 'theme_settings.dart';
import 'about_client.dart';
import 'client_settings_tab.dart';
import 'connect_model.dart';
import 'controller/controller_repository.dart' hide State;

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final model = ConnectModel(ControllerRepository());
  final languages = await LanguageSettings.load();
  final appearance = await ThemeSettings.load();
  runApp(
    WireglowApp(model: model, languages: languages, appearance: appearance),
  );
  unawaited(model.start());
}

ThemeData desktopTheme(Brightness brightness) {
  final dark = brightness == Brightness.dark;
  final scheme = ColorScheme.fromSeed(
    seedColor: const Color(0xffb76a12),
    brightness: brightness,
    primary: dark ? const Color(0xffffb547) : const Color(0xff955000),
    onPrimary: dark ? const Color(0xff30200b) : const Color(0xfffff9ef),
    surface: dark ? const Color(0xff181715) : const Color(0xfffaf8f4),
    onSurface: dark ? const Color(0xfff0eae0) : const Color(0xff29251f),
    onSurfaceVariant: dark ? const Color(0xffbcb2a3) : const Color(0xff6f6557),
    outline: dark ? const Color(0xff857969) : const Color(0xff8d806e),
    outlineVariant: dark ? const Color(0xff423a30) : const Color(0xffded5c7),
    primaryContainer: dark ? const Color(0xff3b2c18) : const Color(0xfff3e2c8),
    onPrimaryContainer: dark
        ? const Color(0xffffdba1)
        : const Color(0xff5c3407),
    secondary: dark ? const Color(0xffd7bd95) : const Color(0xff775b35),
    secondaryContainer: dark
        ? const Color(0xff393125)
        : const Color(0xffeee2ce),
    onSecondaryContainer: dark
        ? const Color(0xffead9bd)
        : const Color(0xff49371e),
    surfaceContainerLowest: dark ? const Color(0xff12110f) : Colors.white,
    surfaceContainerLow: dark
        ? const Color(0xff211f1b)
        : const Color(0xfff1ece3),
    surfaceContainer: dark ? const Color(0xff28251f) : const Color(0xffeee8de),
    surfaceContainerHigh: dark
        ? const Color(0xff302b24)
        : const Color(0xffe9e1d5),
    surfaceContainerHighest: dark
        ? const Color(0xff393228)
        : const Color(0xffe3d9ca),
  );
  return ThemeData(
    useMaterial3: true,
    colorScheme: scheme,
    scaffoldBackgroundColor: scheme.surface,
    visualDensity: VisualDensity.standard,
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(9)),
      ),
    ),
    inputDecorationTheme: const InputDecorationTheme(
      border: OutlineInputBorder(),
    ),
    textTheme: const TextTheme(bodyMedium: TextStyle(fontSize: 14)),
  );
}

class WireglowApp extends StatefulWidget {
  const WireglowApp({
    super.key,
    required this.model,
    this.languages,
    this.appearance,
  });
  final ConnectModel model;
  final LanguageSettings? languages;
  final ThemeSettings? appearance;
  @override
  State<WireglowApp> createState() => _WireglowAppState();
}

class _WireglowAppState extends State<WireglowApp> {
  late final AppLifecycleListener _lifecycle;
  late final LanguageSettings languages;
  late final ThemeSettings appearance;
  @override
  void initState() {
    super.initState();
    languages = widget.languages ?? LanguageSettings();
    appearance = widget.appearance ?? ThemeSettings();
    _lifecycle = AppLifecycleListener(
      onExitRequested: () async {
        await widget.model.close();
        return AppExitResponse.exit;
      },
    );
  }

  @override
  void dispose() {
    _lifecycle.dispose();
    widget.model.dispose();
    if (widget.languages == null) languages.dispose();
    if (widget.appearance == null) appearance.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: Listenable.merge([languages, appearance]),
    builder: (context, _) => MaterialApp(
      title: appName,
      debugShowCheckedModeBanner: false,
      theme: desktopTheme(Brightness.light),
      darkTheme: desktopTheme(Brightness.dark),
      themeMode: appearance.mode,
      locale: languages.locale,
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      localeListResolutionCallback: LanguageSettings.resolve,
      home: ConnectScreen(
        model: widget.model,
        languages: languages,
        appearance: appearance,
      ),
    ),
  );
}

class ConnectScreen extends StatelessWidget {
  const ConnectScreen({
    super.key,
    required this.model,
    this.languages,
    this.appearance,
  });
  final ConnectModel model;
  final LanguageSettings? languages;
  final ThemeSettings? appearance;

  static const preferenceMenuConstraints = BoxConstraints.tightFor(width: 176);
  String value(BuildContext context, String s) =>
      s.isEmpty ? AppLocalizations.of(context)!.notSet : s;
  String status(BuildContext context) {
    if (model.switchingMode) {
      return AppLocalizations.of(context)!.switchingProxyMode;
    }
    if (model.loading) return AppLocalizations.of(context)!.starting;
    if (!model.available) return AppLocalizations.of(context)!.unknownState;
    if (model.stopping) return AppLocalizations.of(context)!.disconnecting;
    if (model.connected) return AppLocalizations.of(context)!.active;
    if (model.connecting) return AppLocalizations.of(context)!.connecting;
    if (model.snapshot.state.name == 'STATE_FAILED') {
      return AppLocalizations.of(context)!.failed;
    }
    return model.selected == null
        ? AppLocalizations.of(context)!.addToConnect
        : AppLocalizations.of(context)!.pressToConnect;
  }

  String actionLabel(BuildContext context) => model.connected
      ? AppLocalizations.of(context)!.disconnect
      : model.connecting
      ? AppLocalizations.of(context)!.cancelConnection
      : model.snapshot.state.name == 'STATE_FAILED'
      ? AppLocalizations.of(context)!.retry
      : AppLocalizations.of(context)!.connect;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: model,
    builder: (context, _) {
      final colors = Theme.of(context).colorScheme;
      return DefaultTabController(
        length: 3,
        child: Scaffold(
          body: SafeArea(
            child: Column(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 18,
                  ),
                  decoration: BoxDecoration(
                    border: Border(
                      bottom: BorderSide(color: colors.outlineVariant),
                    ),
                  ),
                  child: Row(
                    children: [
                      Container(
                        width: 30,
                        height: 30,
                        alignment: Alignment.center,
                        decoration: BoxDecoration(
                          color: colors.primary,
                          borderRadius: BorderRadius.circular(9),
                        ),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(7),
                          child: Image.asset(
                            'linux/runner/resources/app_icon.png',
                            width: 30,
                            height: 30,
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      const Text(appName, style: TextStyle(fontSize: 16)),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: proxyModeSelector(context),
                        ),
                      ),
                      if (languages != null) ...[
                        const SizedBox(width: 8),
                        SizedBox.square(
                          dimension: 36,
                          child: languageSelector(context),
                        ),
                      ],
                      if (appearance != null) ...[
                        const SizedBox(width: 8),
                        SizedBox.square(
                          dimension: 36,
                          child: themeSelector(context),
                        ),
                      ],
                    ],
                  ),
                ),
                TabBar(
                  isScrollable: true,
                  tabAlignment: TabAlignment.start,
                  tabs: [
                    Tab(text: AppLocalizations.of(context)!.connections),
                    Tab(text: AppLocalizations.of(context)!.clientSettings),
                    Tab(text: AppLocalizations.of(context)!.about),
                  ],
                ),
                Expanded(
                  child: TabBarView(
                    children: [
                      LayoutBuilder(
                        builder: (context, constraints) {
                          if (constraints.maxWidth < 640) {
                            return SingleChildScrollView(
                              child: Column(
                                children: [sidebar(context), content(context)],
                              ),
                            );
                          }
                          return Row(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              SizedBox(
                                width: 260,
                                child: ColoredBox(
                                  color: colors.surfaceContainerLow,
                                  child: SingleChildScrollView(
                                    child: sidebar(context),
                                  ),
                                ),
                              ),
                              Expanded(
                                child: SingleChildScrollView(
                                  child: content(context),
                                ),
                              ),
                            ],
                          );
                        },
                      ),
                      ClientSettingsTab(model: model, languages: languages),
                      const AboutClient(),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    },
  );
  Widget proxyModeSelector(BuildContext context) {
    final l = AppLocalizations.of(context)!;
    final displayedMode = model.connected
        ? model.snapshot.activeMode
        : model.preferredMode;
    String label(ProxyMode mode) => mode == ProxyMode.PROXY_MODE_TUN
        ? l.tunMode
        : mode == ProxyMode.PROXY_MODE_SYSTEM
        ? l.systemProxyMode
        : mode == ProxyMode.PROXY_MODE_LOCAL
        ? l.localProxyMode
        : l.proxyModeUnknown;
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 260),
      child: CheckedPopupMenuButton<ProxyMode>(
        key: const Key('proxy-mode-selector'),
        check: model.refreshModeAvailability,
        tooltip: label(displayedMode),
        enabled:
            model.available &&
            !model.commandPending &&
            !model.connecting &&
            !model.stopping,
        initialValue: model.preferredMode,
        onSelected: model.selectMode,
        constraints: const BoxConstraints(minWidth: 144, maxWidth: 200),
        itemBuilder: (context) => [
          for (final mode in [
            ProxyMode.PROXY_MODE_LOCAL,
            ProxyMode.PROXY_MODE_SYSTEM,
            if (defaultTargetPlatform == TargetPlatform.linux ||
                defaultTargetPlatform == TargetPlatform.windows)
              ProxyMode.PROXY_MODE_TUN,
          ])
            PopupMenuItem<ProxyMode>(
              value: mode,
              height: 32,
              padding: const EdgeInsets.symmetric(horizontal: 10),
              enabled: model.modeEnabled(mode),
              child: Semantics(
                checked: mode == displayedMode,
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      Icons.check,
                      size: 16,
                      color: mode == displayedMode
                          ? Theme.of(context).colorScheme.primary
                          : Colors.transparent,
                    ),
                    const SizedBox(width: 8),
                    Flexible(
                      child: Tooltip(
                        message: model.modeUnavailableReason(mode),
                        child: Text(
                          label(mode),
                          style: const TextStyle(fontSize: 13),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
        ],
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (model.switchingMode)
                const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              else
                const Icon(Icons.settings_ethernet, size: 18),
              const SizedBox(width: 8),
              Flexible(
                child: Text(
                  label(displayedMode),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 13),
                ),
              ),
              const SizedBox(width: 4),
              const Icon(Icons.expand_more, size: 16),
            ],
          ),
        ),
      ),
    );
  }

  Widget languageSelector(BuildContext context) {
    final l = AppLocalizations.of(context)!;
    PopupMenuItem<String> item(String value, String label) => PopupMenuItem(
      value: value,
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Semantics(
        checked: languages!.preference == value,
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.check,
              size: 16,
              color: languages!.preference == value
                  ? Theme.of(context).colorScheme.primary
                  : Colors.transparent,
            ),
            const SizedBox(width: 8),
            Flexible(child: Text(label, style: const TextStyle(fontSize: 13))),
          ],
        ),
      ),
    );
    return PopupMenuButton<String>(
      key: const Key('language-selector'),
      tooltip: l.language,
      initialValue: languages!.preference,
      icon: const Icon(Icons.public, size: 20),
      padding: EdgeInsets.zero,
      menuPadding: const EdgeInsets.symmetric(vertical: 4),
      constraints: preferenceMenuConstraints,
      style: IconButton.styleFrom(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
        side: BorderSide(color: Theme.of(context).colorScheme.outlineVariant),
      ),
      itemBuilder: (context) => [
        item('system', l.systemLanguage),
        item('en', 'English'),
        item('ru', 'Русский'),
      ],
      onSelected: (value) async {
        try {
          await languages!.select(value);
        } catch (_) {
          if (!context.mounted) return;
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(AppLocalizations.of(context)!.settingsFailed),
            ),
          );
        }
      },
    );
  }

  Widget themeSelector(BuildContext context) {
    final l = AppLocalizations.of(context)!;
    final colors = Theme.of(context).colorScheme;
    return PopupMenuButton<ThemeMode>(
      key: const Key('theme-selector'),
      tooltip: l.colorTheme,
      initialValue: appearance!.mode,
      icon: Icon(switch (appearance!.mode) {
        ThemeMode.system => Icons.desktop_windows_outlined,
        ThemeMode.dark => Icons.dark_mode_outlined,
        ThemeMode.light => Icons.light_mode_outlined,
      }, size: 20),
      padding: EdgeInsets.zero,
      menuPadding: const EdgeInsets.symmetric(vertical: 4),
      constraints: preferenceMenuConstraints,
      style: IconButton.styleFrom(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
        side: BorderSide(color: colors.outlineVariant),
      ),
      itemBuilder: (context) => [
        for (final mode in [ThemeMode.system, ThemeMode.dark, ThemeMode.light])
          PopupMenuItem<ThemeMode>(
            value: mode,
            height: 32,
            padding: const EdgeInsets.symmetric(horizontal: 10),
            child: Semantics(
              checked: appearance!.mode == mode,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.check,
                    size: 16,
                    color: appearance!.mode == mode
                        ? colors.primary
                        : Colors.transparent,
                  ),
                  const SizedBox(width: 8),
                  Flexible(
                    child: Text(switch (mode) {
                      ThemeMode.system => l.systemTheme,
                      ThemeMode.dark => l.darkTheme,
                      ThemeMode.light => l.lightTheme,
                    }, style: const TextStyle(fontSize: 13)),
                  ),
                ],
              ),
            ),
          ),
      ],
      onSelected: (mode) async {
        try {
          await appearance!.select(mode);
        } catch (_) {
          if (!context.mounted) return;
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(AppLocalizations.of(context)!.settingsFailed),
            ),
          );
        }
      },
    );
  }

  Widget sidebar(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Container(
      color: colors.surfaceContainerLow,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(AppLocalizations.of(context)!.profiles),
              Text('${model.profiles.length}'),
            ],
          ),
          const SizedBox(height: 16),
          OutlinedButton.icon(
            onPressed: model.available && !model.busy
                ? () => addDialog(context)
                : null,
            icon: const Icon(Icons.add, size: 18),
            label: Text(AppLocalizations.of(context)!.addByLink),
          ),
          const SizedBox(height: 16),
          for (final p in model.profiles)
            Padding(
              padding: const EdgeInsets.only(bottom: 6),
              child: Container(
                decoration: BoxDecoration(
                  color: p.id == model.selected?.id
                      ? colors.primaryContainer
                      : Colors.transparent,
                  borderRadius: BorderRadius.circular(9),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: TextButton(
                        style: TextButton.styleFrom(
                          alignment: Alignment.centerLeft,
                          foregroundColor: colors.onSurface,
                          padding: const EdgeInsets.all(12),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(9),
                          ),
                        ),
                        onPressed: model.available && !model.busy
                            ? () => model.select(p)
                            : null,
                        child: Row(
                          children: [
                            const Icon(Icons.bookmark_border, size: 18),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    p.name.isEmpty
                                        ? AppLocalizations.of(context)!.unnamed
                                        : p.name,
                                  ),
                                  Text(
                                    '${value(context, p.provider)}${p.id == model.selected?.id && model.connected ? ' · ${AppLocalizations.of(context)!.connected}' : ''}',
                                    style: TextStyle(
                                      fontSize: 12,
                                      color: colors.onSurfaceVariant,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    PopupMenuButton<String>(
                      key: ValueKey('profile-actions-${p.id}'),
                      tooltip: AppLocalizations.of(context)!.profileActions,
                      enabled:
                          model.available &&
                          !model.commandPending &&
                          !model.connecting &&
                          !model.stopping,
                      icon: const Icon(Icons.more_horiz, size: 18),
                      onSelected: (action) {
                        if (action == 'edit') {
                          editDialog(context, p);
                        } else if (action == 'delete') {
                          deleteDialog(context, p);
                        }
                      },
                      itemBuilder: (context) => [
                        PopupMenuItem(
                          value: 'edit',
                          child: Row(
                            children: [
                              const Icon(Icons.edit_outlined, size: 18),
                              const SizedBox(width: 10),
                              Text(AppLocalizations.of(context)!.editProfile),
                            ],
                          ),
                        ),
                        PopupMenuItem(
                          value: 'delete',
                          enabled: !model.busy,
                          child: Row(
                            children: [
                              Icon(
                                Icons.delete_outline,
                                size: 18,
                                color: Theme.of(context).colorScheme.error,
                              ),
                              const SizedBox(width: 10),
                              Text(
                                AppLocalizations.of(context)!.delete,
                                style: TextStyle(
                                  color: Theme.of(context).colorScheme.error,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget latencyIndicator(BuildContext context) {
    final latency = model.snapshot.latency;
    final checking = latency.checking || model.latencyPending;
    String timeLabel(int milliseconds) {
      final time = DateTime.fromMillisecondsSinceEpoch(milliseconds).toLocal();
      return "${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}:${time.second.toString().padLeft(2, '0')}";
    }

    final success = latency.successAt.toInt();
    final text = checking
        ? AppLocalizations.of(context)!.checkingLatency
        : latency.error.isNotEmpty
        ? AppLocalizations.of(context)!.checkFailed
        : success > 0
        ? AppLocalizations.of(context)!.latency(latency.delayMs.toString())
        : AppLocalizations.of(context)!.latencyNotMeasured;
    return Padding(
      padding: const EdgeInsets.only(top: 10),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Flexible(
                child: Tooltip(
                  message: AppLocalizations.of(context)!.latencyHelp,
                  child: Text(text, textAlign: TextAlign.center),
                ),
              ),
              IconButton(
                tooltip: AppLocalizations.of(context)!.checkLatency,
                onPressed: checking ? null : model.checkLatency,
                icon: checking
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.refresh, size: 18),
              ),
            ],
          ),
          if (success > 0)
            Text(
              AppLocalizations.of(context)!.lastSuccess(timeLabel(success)) +
                  (latency.error.isNotEmpty
                      ? " · ${AppLocalizations.of(context)!.latency(latency.delayMs.toString())}"
                      : ""),
              style: TextStyle(
                fontSize: 12,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
        ],
      ),
    );
  }

  Widget content(BuildContext context) {
    final p = model.selected;
    final colors = Theme.of(context).colorScheme;
    final failed = model.snapshot.state.name == 'STATE_FAILED';
    final amber = colors.primary;
    const activeAmber = Color(0xffffb547);
    final buttonColor = model.connected
        ? const Color(0xff211f1b)
        : failed
        ? colors.errorContainer
        : model.busy
        ? colors.primaryContainer
        : colors.primary;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 40, vertical: 32),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            p == null
                ? AppLocalizations.of(context)!.noConnections
                : p.name.isEmpty
                ? AppLocalizations.of(context)!.unnamed
                : p.name,
            style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w500),
          ),
          if (!model.modeAvailable)
            Padding(
              padding: const EdgeInsets.only(top: 10),
              child: Text(
                model.systemProxyUnavailableReason,
                textAlign: TextAlign.center,
                style: TextStyle(color: colors.error),
              ),
            ),
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 32),
            child: Column(
              children: [
                Tooltip(
                  message: actionLabel(context),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    width: 96,
                    height: 96,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      boxShadow: model.connected
                          ? [
                              BoxShadow(
                                color: activeAmber.withValues(alpha: 0.25),
                                blurRadius: 18,
                                spreadRadius: 1,
                              ),
                            ]
                          : [],
                    ),
                    child: FilledButton(
                      onPressed:
                          model.available &&
                              p != null &&
                              !model.stopping &&
                              !model.commandPending &&
                              (model.busySession || model.modeAvailable)
                          ? model.action
                          : null,
                      style: FilledButton.styleFrom(
                        backgroundColor: buttonColor,
                        elevation: 0,
                        surfaceTintColor: Colors.transparent,
                        side: BorderSide(
                          color: model.connected
                              ? activeAmber
                              : Colors.transparent,
                          width: 2,
                        ),
                        foregroundColor: model.connected
                            ? activeAmber
                            : failed
                            ? colors.onErrorContainer
                            : model.busy
                            ? colors.primary
                            : colors.onPrimary,
                        shape: const CircleBorder(),
                        padding: EdgeInsets.zero,
                      ),
                      child: model.stopping || model.loading
                          ? const SizedBox(
                              width: 30,
                              height: 30,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : Icon(
                              model.connecting
                                  ? Icons.stop
                                  : failed
                                  ? Icons.refresh
                                  : Icons.power_settings_new,
                              size: 34,
                              semanticLabel: actionLabel(context),
                            ),
                    ),
                  ),
                ),
                const SizedBox(height: 14),
                Semantics(
                  liveRegion: true,
                  child: Text(
                    status(context),
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: failed
                          ? colors.error
                          : model.connected
                          ? amber
                          : colors.onSurface,
                    ),
                  ),
                ),
                if (model.connected) latencyIndicator(context),
                if (p != null) ...[
                  const SizedBox(height: 14),
                  SelectableText(
                    '${p.provider.isEmpty ? '…' : p.provider} — ${p.peer.isEmpty ? '…' : p.peer}${p.provider == 'vk' ? ' — ${p.streams}/${p.streamsPerCredential}' : ''}',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 12,
                      color: colors.onSurfaceVariant,
                    ),
                  ),
                ],
                if (model.error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 16),
                    child: Text(
                      uiErrorText(context, model.error!),
                      textAlign: TextAlign.center,
                      style: TextStyle(color: colors.error),
                    ),
                  ),
                if (model.error == null &&
                    model.snapshot.error.isNotEmpty &&
                    !failed)
                  Padding(
                    padding: const EdgeInsets.only(top: 16),
                    child: Text(
                      model.snapshot.error,
                      textAlign: TextAlign.center,
                      style: TextStyle(color: colors.error),
                    ),
                  ),
                if (failed) ...[
                  const SizedBox(height: 12),
                  Text(
                    AppLocalizations.of(context)!.retryHelp,
                    textAlign: TextAlign.center,
                  ),
                  ExpansionTile(
                    title: Text(AppLocalizations.of(context)!.errorDetails),
                    children: [
                      Padding(
                        padding: const EdgeInsets.all(12),
                        child: SelectableText(
                          model.snapshot.error.isEmpty
                              ? AppLocalizations.of(context)!.noErrorDetails
                              : model.snapshot.error,
                        ),
                      ),
                    ],
                  ),
                ],
                if (model.connected && model.snapshot.mixedAddress.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: Wrap(
                      crossAxisAlignment: WrapCrossAlignment.center,
                      alignment: WrapAlignment.center,
                      children: [
                        SelectableText(
                          AppLocalizations.of(context)!
                              .localProxy(model.snapshot.mixedAddress),
                          style: const TextStyle(fontSize: 12),
                        ),
                        IconButton(
                          tooltip: AppLocalizations.of(context)!
                              .copyProxyAddress,
                          onPressed: () => Clipboard.setData(
                            ClipboardData(text: model.snapshot.mixedAddress),
                          ),
                          icon: const Icon(Icons.copy, size: 16),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
          if (p != null)
            ExpansionTile(
              key: ValueKey(p.id),
              tilePadding: EdgeInsets.zero,
              title: Text(
                AppLocalizations.of(context)!.properties,
                style: TextStyle(fontSize: 14, color: colors.primary),
              ),
              children: [properties(context, p)],
            ),
          const Divider(),
          const SizedBox(height: 12),
          Text(
            model.connected &&
                    model.snapshot.activeMode ==
                        ProxyMode.PROXY_MODE_UNSPECIFIED
                ? AppLocalizations.of(context)!.proxyModeUnknown
                : (model.busySession
                          ? model.snapshot.activeMode
                          : model.preferredMode) ==
                      ProxyMode.PROXY_MODE_SYSTEM
                ? AppLocalizations.of(context)!.systemProxy
                : (model.busySession
                          ? model.snapshot.activeMode
                          : model.preferredMode) ==
                      ProxyMode.PROXY_MODE_TUN
                ? (defaultTargetPlatform == TargetPlatform.windows
                      ? AppLocalizations.of(context)!.tunHelpWindows
                      : AppLocalizations.of(context)!.tunHelp)
                : AppLocalizations.of(context)!.localProxyHelp,
            style: TextStyle(fontSize: 12, color: colors.onSurfaceVariant),
          ),
        ],
      ),
    );
  }

  Widget properties(BuildContext context, ProfileSummary p) {
    final fields = <String, String>{
      AppLocalizations.of(context)!.provider: value(context, p.provider),
      'Peer': value(context, p.peer),
      AppLocalizations.of(context)!.turnTransport: p.provider == 'direct'
          ? AppLocalizations.of(context)!.notApplicable
          : value(context, p.turnTransport.toUpperCase()),
      AppLocalizations.of(context)!.relayMode: value(
        context,
        p.relayMode.toUpperCase(),
      ),
      if (p.provider == 'vk')
        AppLocalizations.of(context)!.streams: '${p.streams}',
      if (p.provider == 'vk')
        AppLocalizations.of(context)!.streamsPerCredential:
            '${p.streamsPerCredential}',
      if (p.obfuscationMode.isNotEmpty && p.obfuscationMode != 'none')
        AppLocalizations.of(context)!.obfuscation: p.obfuscationMode,
    };
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 12),
      child: LayoutBuilder(
        builder: (context, box) => Wrap(
          spacing: 24,
          runSpacing: 14,
          children: [
            for (final e in fields.entries)
              SizedBox(
                width: box.maxWidth < 380
                    ? box.maxWidth
                    : (box.maxWidth - 24) / 2,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      e.key,
                      style: TextStyle(
                        fontSize: 12,
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 3),
                    SelectableText(e.value),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }

  Future<void> deleteDialog(BuildContext context, ProfileSummary p) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => ListenableBuilder(
        listenable: model,
        builder: (context, _) => AlertDialog(
          title: Text(AppLocalizations.of(context)!.confirmDelete),
          content: Text(
            AppLocalizations.of(context)!.deleteDescription(
              p.name.isEmpty ? AppLocalizations.of(context)!.unnamed : p.name,
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: Text(AppLocalizations.of(context)!.cancel),
            ),
            FilledButton(
              style: FilledButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.error,
                foregroundColor: Theme.of(context).colorScheme.onError,
              ),
              onPressed: model.available && !model.busy
                  ? () => Navigator.pop(context, true)
                  : null,
              child: Text(AppLocalizations.of(context)!.delete),
            ),
          ],
        ),
      ),
    );
    if (confirmed == true) await model.deleteProfile(p.id);
  }

  Future<void> addDialog(BuildContext context) => showDialog<void>(
    context: context,
    barrierDismissible: false,
    builder: (_) => AddProfileDialog(model: model),
  );

  Future<void> editDialog(BuildContext context, ProfileSummary p) async {
    final saved = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (_) => EditProfileDialog(model: model, summary: p),
    );
    if (saved == true && context.mounted && model.connected) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Изменения будут применены при следующем подключении.'),
        ),
      );
    }
  }
}

class EditProfileDialog extends StatefulWidget {
  const EditProfileDialog({
    super.key,
    required this.model,
    required this.summary,
  });
  final ConnectModel model;
  final ProfileSummary summary;
  @override
  State<EditProfileDialog> createState() => _EditProfileDialogState();
}

class _EditProfileDialogState extends State<EditProfileDialog> {
  final name = TextEditingController();
  final streams = TextEditingController(text: '1');
  final perAccount = TextEditingController(text: '1');
  final links = <TextEditingController>[];
  bool loading = true;
  bool pending = false;
  Object? error;
  bool get vk => widget.summary.provider == 'vk';
  @override
  void initState() {
    super.initState();
    unawaited(_load());
  }

  Future<void> _load() async {
    try {
      final profile = await widget.model.repository.getProfile(
        widget.summary.id,
      );
      if (!mounted) return;
      final c = profile.connection;
      name.text = c.name;
      streams.text = (c.streams.toInt() < 1 ? 1 : c.streams.toInt()).toString();
      perAccount.text =
          (c.streamsPerCredential.toInt() < 1
                  ? 1
                  : c.streamsPerCredential.toInt())
              .toString();
      for (final link in (c.hasVk() ? c.vk.callLinks : const <String>[])) {
        links.add(TextEditingController(text: link));
      }
      if (vk && links.isEmpty) links.add(TextEditingController());
    } catch (e) {
      error = ConnectModel.message(e);
    }
    if (mounted) setState(() => loading = false);
  }

  @override
  void dispose() {
    name.dispose();
    streams.dispose();
    perAccount.dispose();
    for (final c in links) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> save() async {
    final s = int.tryParse(streams.text.trim()),
        a = int.tryParse(perAccount.text.trim());
    if (name.text.trim().isEmpty ||
        s == null ||
        s < 1 ||
        a == null ||
        a < 1 ||
        (vk && links.every((c) => c.text.trim().isEmpty))) {
      setState(
        () => error =
            'Укажите имя, минимум 1 поток и хотя бы одну ссылку звонка для VK.',
      );
      return;
    }
    setState(() {
      pending = true;
      error = null;
    });
    try {
      await widget.model.editProfile(
        id: widget.summary.id,
        name: name.text.trim(),
        callLinks: links
            .map((c) => c.text.trim())
            .where((s) => s.isNotEmpty)
            .toList(),
        streams: s,
        streamsPerCredential: a,
      );
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        setState(() {
          pending = false;
          error = ConnectModel.message(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Редактор профиля'),
    content: SizedBox(
      width: 460,
      child: loading
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextField(
                    controller: name,
                    decoration: const InputDecoration(labelText: 'Имя'),
                  ),
                  const SizedBox(height: 12),
                  if (vk) ...[
                    const Text('Ссылки звонков'),
                    for (var i = 0; i < links.length; i++)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Row(
                          children: [
                            Expanded(
                              child: TextField(
                                controller: links[i],
                                decoration: InputDecoration(
                                  labelText: 'Ссылка ${i + 1}',
                                ),
                              ),
                            ),
                            IconButton(
                              onPressed: links.length > 1
                                  ? () => setState(() {
                                      links.removeAt(i).dispose();
                                    })
                                  : null,
                              icon: const Icon(Icons.remove_circle_outline),
                            ),
                          ],
                        ),
                      ),
                    TextButton.icon(
                      onPressed: () =>
                          setState(() => links.add(TextEditingController())),
                      icon: const Icon(Icons.add),
                      label: const Text('Добавить ссылку'),
                    ),
                  ],
                  TextField(
                    controller: streams,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Потоки'),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: perAccount,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(
                      labelText: 'Потоки на аккаунт',
                    ),
                  ),
                  if (error != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 12),
                      child: Text(
                        error.toString(),
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.error,
                        ),
                      ),
                    ),
                ],
              ),
            ),
    ),
    actions: [
      TextButton(
        onPressed: pending ? null : () => Navigator.pop(context),
        child: const Text('Отмена'),
      ),
      FilledButton(
        onPressed: loading || pending ? null : save,
        child: const Text('Сохранить'),
      ),
    ],
  );
}

class AddProfileDialog extends StatefulWidget {
  const AddProfileDialog({super.key, required this.model});
  final ConnectModel model;
  @override
  State<AddProfileDialog> createState() => _AddProfileDialogState();
}

class _AddProfileDialogState extends State<AddProfileDialog> {
  final link = TextEditingController();
  bool pending = false;
  Object? error;
  @override
  void dispose() {
    link.dispose();
    super.dispose();
  }

  Future<void> save() async {
    if (link.text.trim().isEmpty) {
      setState(() => error = const UiError(UiErrorCode.insertLink));
      return;
    }
    setState(() {
      pending = true;
      error = null;
    });
    try {
      await widget.model.add(link.text);
      if (!mounted) return;
      Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        setState(() {
          pending = false;
          error = ConnectModel.message(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => PopScope(
    canPop: !pending,
    child: AlertDialog(
      title: Row(
        children: [
          Expanded(child: Text(AppLocalizations.of(context)!.addProfile)),
          IconButton(
            tooltip: AppLocalizations.of(context)!.closeAdd,
            onPressed: pending ? null : () => Navigator.pop(context),
            icon: const Icon(Icons.close),
          ),
        ],
      ),
      content: SizedBox(
        width: 440,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(AppLocalizations.of(context)!.pasteHelp),
              const SizedBox(height: 20),
              Row(
                children: [
                  Expanded(
                    child: Text(AppLocalizations.of(context)!.connectionLink),
                  ),
                  TextButton(
                    onPressed: pending
                        ? null
                        : () async {
                            final data = await Clipboard.getData(
                              Clipboard.kTextPlain,
                            );
                            if (mounted && !pending) {
                              setState(() => link.text = data?.text ?? '');
                            }
                          },
                    child: Text(AppLocalizations.of(context)!.paste),
                  ),
                ],
              ),
              TextField(
                controller: link,
                autofocus: true,
                enabled: !pending,
                minLines: 3,
                maxLines: 5,
                autocorrect: false,
                enableSuggestions: false,
                decoration: InputDecoration(hintText: 'freeturn://…'),
              ),
              const SizedBox(height: 10),
              Text(
                AppLocalizations.of(context)!.nameFromLink,
                style: TextStyle(fontSize: 12),
              ),
              if (error != null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Semantics(
                    liveRegion: true,
                    child: Text(
                      uiErrorText(context, error!),
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ),
                ),
              if (pending)
                const Padding(
                  padding: EdgeInsets.only(top: 16),
                  child: LinearProgressIndicator(),
                ),
            ],
          ),
        ),
      ),
      actions: [
        FilledButton(
          onPressed: pending ? null : save,
          child: Text(AppLocalizations.of(context)!.add),
        ),
      ],
    ),
  );
}
