import 'package:flutter/material.dart';

import 'connect_model.dart';
import 'controller/controller_repository.dart' hide State;
import 'language_settings.dart';
import 'l10n/app_localizations.dart';
import 'ui_error_text.dart';

class ClientSettingsTab extends StatefulWidget {
  const ClientSettingsTab({super.key, required this.model, this.languages});
  final ConnectModel model;
  final LanguageSettings? languages;

  @override
  State<ClientSettingsTab> createState() => _ClientSettingsTabState();
}

class _ClientSettingsTabState extends State<ClientSettingsTab> {
  var _form = GlobalKey<FormState>();
  late final TextEditingController _proxy;
  late final TextEditingController _relay;
  late final TextEditingController _api;
  bool _pending = false;
  Object? _error;
  late String _savedProxy;
  late String _savedRelay;
  late int _savedAPI;

  @override
  void initState() {
    super.initState();
    _proxy = TextEditingController(text: widget.model.proxyAddress);
    _relay = TextEditingController(text: widget.model.relayAddress);
    _api = TextEditingController(text: widget.model.apiPort.toString());
    _savedProxy = widget.model.proxyAddress;
    _savedRelay = widget.model.relayAddress;
    _savedAPI = widget.model.apiPort;
  }

  @override
  void didUpdateWidget(covariant ClientSettingsTab oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_proxy.text == _savedProxy) _proxy.text = widget.model.proxyAddress;
    if (_relay.text == _savedRelay) _relay.text = widget.model.relayAddress;
    if (_api.text == _savedAPI.toString()) {
      _api.text = widget.model.apiPort.toString();
    }
    _savedProxy = widget.model.proxyAddress;
    _savedRelay = widget.model.relayAddress;
    _savedAPI = widget.model.apiPort;
  }

  @override
  void dispose() {
    _proxy.dispose();
    _relay.dispose();
    _api.dispose();
    super.dispose();
  }

  int? port(String? value) {
    final match = RegExp(r'^127\.0\.0\.1:([0-9]+)$')
        .firstMatch(value?.trim() ?? '');
    return match == null ? null : int.tryParse(match[1]!);
  }

  String normalizedAddress(String value) {
    final number = port(value);
    return number == null ? value.trim() : '127.0.0.1:$number';
  }

  bool get addressesChanged =>
      normalizedAddress(_proxy.text) !=
          normalizedAddress(widget.model.proxyAddress) ||
      normalizedAddress(_relay.text) !=
          normalizedAddress(widget.model.relayAddress) ||
      apiPort(_api.text) != widget.model.apiPort;

  int? apiPort(String? value) {
    final text = value?.trim() ?? '';
    return RegExp(r'^[0-9]+$').hasMatch(text) ? int.tryParse(text) : null;
  }

  bool collides(int number) =>
      [
        port(_proxy.text),
        port(_relay.text),
        apiPort(_api.text),
      ].where((p) => p == number).length >
      1;

  String? validate(String? value) {
    final l = AppLocalizations.of(context)!;
    final number = port(value);
    if (number == null || number < 1 || number > 65535) {
      return l.invalidLocalAddress;
    }
    return collides(number) ? l.collidingPorts : null;
  }

  String? validateAPI(String? value) {
    final l = AppLocalizations.of(context)!;
    final number = apiPort(value);
    if (number == null || number < 1 || number > 65535) return l.invalidAPIPort;
    return collides(number) ? l.collidingPorts : null;
  }

  Future<void> save() async {
    if (!addressesChanged || !_form.currentState!.validate()) return;
    await perform(
      () => widget.model.saveAddresses(
        '127.0.0.1:${port(_proxy.text)}',
        '127.0.0.1:${port(_relay.text)}',
        apiPort(_api.text)!,
      ),
      reset: false,
    );
  }

  Future<void> reset() async {
    final l = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l.confirmResetSettings),
        content: Text(l.resetSettingsDescription),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l.reset),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    await perform(() async {
      await widget.model.resetSettings();
      await widget.languages?.select('system');
    }, reset: true);
  }

  Future<void> perform(
    Future<void> Function() action, {
    required bool reset,
  }) async {
    setState(() {
      _pending = true;
      _error = null;
    });
    try {
      await action();
      if (!mounted) return;
      _proxy.text = widget.model.proxyAddress;
      _relay.text = widget.model.relayAddress;
      _api.text = widget.model.apiPort.toString();
      // Recreate validation state without invoking the immediate preference callbacks.
      _form = GlobalKey<FormState>();
      final l = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(reset ? l.settingsReset : l.settingsSaved)),
      );
    } catch (e) {
      if (mounted) setState(() => _error = ConnectModel.message(e));
    } finally {
      if (mounted) setState(() => _pending = false);
    }
  }

  Widget block(
    BuildContext context,
    String title,
    List<Widget> children, {
    bool nested = false,
  }) => Container(
    width: double.infinity,
    padding: const EdgeInsets.all(20),
    decoration: BoxDecoration(
      border: Border.all(color: Theme.of(context).colorScheme.outlineVariant),
      borderRadius: BorderRadius.circular(12),
      color: nested ? Theme.of(context).colorScheme.surfaceContainerLow : null,
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: nested
              ? Theme.of(context).textTheme.titleMedium
              : Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 20),
        ...children,
      ],
    ),
  );

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context)!;
    final m = widget.model;
    final enabled =
        m.available &&
        !m.commandPending &&
        !m.connecting &&
        !m.stopping &&
        !_pending;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Align(
        alignment: Alignment.topLeft,
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 560),
          child: Form(
            key: _form,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                block(context, l.generalSettings, [
                  DropdownButtonFormField<ProxyMode>(
                    key: ValueKey(m.preferredMode),
                    initialValue: m.preferredMode,
                    isExpanded: true,
                    decoration: InputDecoration(labelText: l.proxyMode),
                    items: [
                      DropdownMenuItem(
                        value: ProxyMode.PROXY_MODE_LOCAL,
                        child: Text(l.localProxyMode),
                      ),
                      DropdownMenuItem(
                        value: ProxyMode.PROXY_MODE_SYSTEM,
                        enabled: m.systemProxyUnavailableReason.isEmpty,
                        child: Text(l.systemProxyMode),
                      ),
                    ],
                    onChanged: enabled ? (mode) => m.selectMode(mode!) : null,
                  ),
                  if (m.systemProxyUnavailableReason.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text(m.systemProxyUnavailableReason),
                  ],
                  if (widget.languages != null) ...[
                    const SizedBox(height: 16),
                    DropdownButtonFormField<String>(
                      key: ValueKey(widget.languages!.preference),
                      initialValue: widget.languages!.preference,
                      isExpanded: true,
                      decoration: InputDecoration(labelText: l.language),
                      items: [
                        DropdownMenuItem(
                          value: 'system',
                          child: Text(l.systemLanguage),
                        ),
                        const DropdownMenuItem(
                          value: 'en',
                          child: Text('English'),
                        ),
                        const DropdownMenuItem(
                          value: 'ru',
                          child: Text('Русский'),
                        ),
                      ],
                      onChanged: enabled
                          ? (value) async {
                              try {
                                await widget.languages!.select(value!);
                              } catch (_) {
                                if (mounted) {
                                  setState(() => _error = l.settingsFailed);
                                }
                              }
                            }
                          : null,
                    ),
                  ],
                ]),
                const SizedBox(height: 20),
                block(context, l.networkSettings, [
                  Text(l.addressesHelp),
                  const SizedBox(height: 16),
                  TextFormField(
                    key: const Key('proxy-address'),
                    controller: _proxy,
                    onChanged: (_) => setState(() {}),
                    enabled: enabled,
                    validator: validate,
                    decoration: InputDecoration(labelText: l.proxyAddressLabel),
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    key: const Key('api-port'),
                    controller: _api,
                    onChanged: (_) => setState(() {}),
                    enabled: enabled,
                    validator: validateAPI,
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(labelText: l.apiPortLabel),
                  ),
                  if (m.busySession) ...[
                    const SizedBox(height: 12),
                    Text(l.activeSettingsHelp),
                  ],
                ]),
                const SizedBox(height: 20),
                block(context, l.relaySettings, [
                  block(context, 'FreeTurn', [
                    TextFormField(
                      key: const Key('relay-address'),
                      controller: _relay,
                      onChanged: (_) => setState(() {}),
                      enabled: enabled,
                      validator: validate,
                      decoration: InputDecoration(
                        labelText: l.relayAddressLabel,
                      ),
                    ),
                  ], nested: true),
                ]),
                if (_error != null || m.error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    uiErrorText(context, _error ?? m.error!),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ],
                const SizedBox(height: 24),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  children: [
                    FilledButton(
                      onPressed: enabled && addressesChanged ? save : null,
                      child: Text(l.saveSettings),
                    ),
                    OutlinedButton(
                      onPressed: enabled ? reset : null,
                      child: Text(l.resetSettings),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
