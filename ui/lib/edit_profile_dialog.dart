import 'dart:async';

import 'package:fixnum/fixnum.dart';

import 'package:flutter/material.dart';

import 'connect_model.dart';
import 'controller/controller_repository.dart' hide State;
import 'l10n/app_localizations.dart';

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
  final form = GlobalKey<FormState>();
  final name = TextEditingController();
  final streams = TextEditingController();
  final credentials = TextEditingController();
  final turnAddress = TextEditingController();
  final dnsServers = TextEditingController();
  final pacing = TextEditingController();
  final links = <TextEditingController>[];
  bool loading = true,
      loaded = false,
      pending = false,
      advanced = false,
      manualCaptcha = false;
  String transport = '', dnsMode = '', platform = '', obfuscation = '';
  Object? error;
  bool get vk => widget.summary.provider == 'vk';
  bool get hasObfuscation => obfuscation.isNotEmpty && obfuscation != 'none';
  AppLocalizations get l => AppLocalizations.of(context)!;

  @override
  void initState() {
    super.initState();
    unawaited(load());
  }

  Future<void> load() async {
    try {
      final profile = await widget.model.repository.getProfile(
        widget.summary.id,
      );
      if (!mounted) return;
      final c = profile.connection;
      name.text = c.name;
      streams.text = c.streams.toInt() == 0 ? '' : '${c.streams}';
      credentials.text = c.streamsPerCredential.toInt() == 0
          ? ''
          : '${c.streamsPerCredential}';
      transport = c.turnTransport;
      dnsMode = c.dns.mode;
      dnsServers.text = c.dns.servers.join(', ');
      platform = c.vk.platform;
      manualCaptcha = c.vk.manualCaptcha;
      turnAddress.text = c.vk.turnAddress;
      obfuscation = c.obfuscation.mode;
      pacing.text = '${c.obfuscation.timingMs}';
      for (final link in c.vk.callLinks) {
        links.add(TextEditingController(text: link));
      }
      if (vk && links.isEmpty) links.add(TextEditingController());
      loaded = true;
    } catch (e) {
      error = ConnectModel.message(e);
    }
    if (mounted) setState(() => loading = false);
  }

  @override
  void dispose() {
    for (final c in [
      name,
      streams,
      credentials,
      turnAddress,
      dnsServers,
      pacing,
      ...links,
    ]) {
      c.dispose();
    }
    super.dispose();
  }

  String? countError(String? value) {
    if (value == null || value.trim().isEmpty) return null;
    final n = int.tryParse(value.trim());
    return n == null || n < 1 || n > 1024 ? l.editorInvalidStreams : null;
  }

  Future<void> save() async {
    if (!(form.currentState?.validate() ?? false)) return;
    final calls = links
        .map((c) => c.text.trim())
        .where((s) => s.isNotEmpty)
        .toList();
    if (vk && (calls.isEmpty || calls.any((s) => s.contains(',')))) {
      setState(() => error = l.editorInvalidLinks);
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
        callLinks: calls,
        streams: int.tryParse(streams.text.trim()) ?? 0,
        streamsPerCredential: int.tryParse(credentials.text.trim()) ?? 0,
        advanced: ProfileAdvancedEdits(
          turnTransport: transport,
          dns: DNSSettings(
            mode: dnsMode,
            servers: dnsServers.text
                .split(',')
                .map((s) => s.trim())
                .where((s) => s.isNotEmpty),
          ),
          manualCaptcha: manualCaptcha,
          vkPlatform: platform,
          turnAddress: turnAddress.text.trim(),
          pacingMs: Int64(int.tryParse(pacing.text.trim()) ?? 0),
        ),
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

  Widget section(String title, List<Widget> children) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      Text(title, style: Theme.of(context).textTheme.titleSmall),
      const SizedBox(height: 8),
      ...children,
    ],
  );

  Widget help(String text) => Padding(
    padding: const EdgeInsets.only(top: 6),
    child: Text(
      text,
      style: Theme.of(context).textTheme.bodySmall
          ?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant),
    ),
  );

  Widget divider() => const Padding(
    padding: EdgeInsets.symmetric(vertical: 10),
    child: Divider(height: 1),
  );

  Widget pair(Widget a, Widget b) => LayoutBuilder(
    builder: (context, box) => box.maxWidth < 440
        ? Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [a, const SizedBox(height: 14), b],
          )
        : Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(child: a),
              const SizedBox(width: 16),
              Expanded(child: b),
            ],
          ),
  );

  Widget choice(
    String label,
    String value,
    Map<String, String> options,
    ValueChanged<String> change,
  ) => DropdownButtonFormField<String>(
    initialValue: value,
    isExpanded: true,
    decoration: InputDecoration(labelText: label),
    items: {...options, if (!options.containsKey(value)) value: value}.entries
        .map(
          (e) => DropdownMenuItem(
            value: e.key,
            child: Text(e.value, overflow: TextOverflow.ellipsis),
          ),
        )
        .toList(),
    onChanged: pending
        ? null
        : (v) {
            if (v != null) setState(() => change(v));
          },
  );

  @override
  Widget build(BuildContext context) => PopScope(
    canPop: !pending,
    child: AlertDialog(
      title: Text(l.editorTitle),
      contentPadding: const EdgeInsets.fromLTRB(24, 12, 24, 12),
      content: SizedBox(
        width: 560,
        child: loading
            ? const Center(child: CircularProgressIndicator())
            : SingleChildScrollView(
                child: Form(
                  key: form,
                  child: AbsorbPointer(
                    absorbing: pending,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        section(l.editorProfile, [
                          TextFormField(
                            controller: name,
                            maxLength: 80,
                            decoration: InputDecoration(
                              isDense: true,
                              labelText: l.editorName,
                              counterText: '',
                            ),
                            validator: (v) => v == null || v.trim().isEmpty
                                ? l.editorNameRequired
                                : null,
                          ),
                        ]),
                        if (vk) ...[
                          divider(),
                          section(l.editorVK, [
                            for (var i = 0; i < links.length; i++)
                              Padding(
                                padding: const EdgeInsets.only(top: 8),
                                child: Row(
                                  children: [
                                    Expanded(
                                      child: TextFormField(
                                        controller: links[i],
                                        decoration: InputDecoration(
                                          isDense: true,
                                          labelText: l.editorCallLink(i + 1),
                                        ),
                                      ),
                                    ),
                                    IconButton(
                                      tooltip: l.editorRemoveLink,
                                      onPressed: links.length > 1
                                          ? () => setState(() {
                                              links.removeAt(i).dispose();
                                            })
                                          : null,
                                      icon: const Icon(
                                        Icons.remove_circle_outline,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            Align(
                              alignment: Alignment.centerLeft,
                              child: TextButton.icon(
                                onPressed: () => setState(
                                  () => links.add(TextEditingController()),
                                ),
                                icon: const Icon(Icons.add, size: 18),
                                label: Text(l.editorAddLink),
                              ),
                            ),
                          ]),
                        ],
                        divider(),
                        section(l.editorPerformance, [
                          pair(
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                TextFormField(
                                  controller: streams,
                                  keyboardType: TextInputType.number,
                                  decoration: InputDecoration(
                                    labelText: vk ? l.editorStreams : l.streams,
                                    floatingLabelBehavior:
                                        FloatingLabelBehavior.always,
                                    isDense: true,
                                    hintText: l.editorDefault(vk ? '12' : '1'),
                                  ),
                                  validator: countError,
                                ),
                                if (vk) help(l.editorStreamsHelp),
                              ],
                            ),
                            ifCredentialField(),
                          ),
                        ]),
                        divider(),
                        ExpansionTile(
                          key: const Key('profile-advanced'),
                          minTileHeight: 40,
                          tilePadding: EdgeInsets.zero,
                          childrenPadding: EdgeInsets.zero,
                          initiallyExpanded: advanced,
                          onExpansionChanged: (v) =>
                              setState(() => advanced = v),
                          shape: const Border(),
                          collapsedShape: const Border(),
                          title: Text(
                            l.editorAdvanced,
                            style: Theme.of(context).textTheme.titleSmall,
                          ),
                          children: [
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                const SizedBox(height: 10),
                                if (vk) ...[
                                  section(l.editorVKAuth, [
                                    choice(l.editorPlatform, platform, {
                                      '': l.editorDefault('Desktop'),
                                      'desktop': 'Desktop',
                                      'mobile': 'Mobile',
                                    }, (v) => platform = v),
                                    CheckboxListTile(
                                      contentPadding: EdgeInsets.zero,
                                      title: Text(l.editorCaptcha),
                                      subtitle: Text(l.editorCaptchaHelp),
                                      value: manualCaptcha,
                                      onChanged: (v) => setState(
                                        () => manualCaptcha = v ?? false,
                                      ),
                                    ),
                                  ]),
                                  divider(),
                                  section('TURN', [
                                    choice(l.editorTransport, transport, {
                                      '': l.editorDefault('TCP'),
                                      'tcp': 'TCP',
                                      'udp': 'UDP',
                                    }, (v) => transport = v),
                                    if (transport == 'udp')
                                      help(l.editorUDPHelp),
                                    const SizedBox(height: 16),
                                    TextFormField(
                                      controller: turnAddress,
                                      decoration: InputDecoration(
                                        labelText: l.editorTurnAddress,
                                        hintText: l.editorTurnHint,
                                      ),
                                    ),
                                    help(l.editorTurnHelp),
                                  ]),
                                  divider(),
                                ],
                                section('DNS', [
                                  choice(l.editorDNSMode, dnsMode, {
                                    '': l.editorDefault(l.editorDNSAuto),
                                    'auto': l.editorDNSAuto,
                                    'plain': l.editorDNSPlain,
                                    'doh': 'DoH',
                                  }, (v) => dnsMode = v),
                                  const SizedBox(height: 16),
                                  TextFormField(
                                    controller: dnsServers,
                                    enabled: dnsMode != 'doh',
                                    decoration: InputDecoration(
                                      labelText: l.editorDNSServers,
                                      hintText: '1.1.1.1, 9.9.9.9:53',
                                    ),
                                  ),
                                  help(
                                    dnsMode == 'doh'
                                        ? l.editorDoHHelp
                                        : l.editorDNSHelp,
                                  ),
                                  help(l.editorTunDNSHelp),
                                ]),
                                divider(),
                                section(l.editorObfuscation, [
                                  Text(
                                    hasObfuscation
                                        ? obfuscation
                                        : l.editorObfuscationOff,
                                  ),
                                  const SizedBox(height: 12),
                                  TextFormField(
                                    controller: pacing,
                                    enabled: hasObfuscation,
                                    keyboardType: TextInputType.number,
                                    decoration: InputDecoration(
                                      labelText: l.editorPacing,
                                    ),
                                    validator: (v) {
                                      if (!hasObfuscation) return null;
                                      final n = int.tryParse(v?.trim() ?? '');
                                      return n == null || n < 0 || n > 60000
                                          ? l.editorInvalidPacing
                                          : null;
                                    },
                                  ),
                                  help(
                                    hasObfuscation
                                        ? l.editorPacingHelp
                                        : l.editorPacingDisabled,
                                  ),
                                ]),
                                const SizedBox(height: 14),
                              ],
                            ),
                          ],
                        ),
                        help(l.editorReconnect),
                        if (error != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 12),
                            child: Text(
                              '$error',
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.error,
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
              ),
      ),
      actions: [
        TextButton(
          onPressed: pending ? null : () => Navigator.pop(context),
          child: Text(l.cancel),
        ),
        FilledButton(
          onPressed: !loaded || pending ? null : save,
          child: Text(l.saveSettings),
        ),
      ],
    ),
  );

  Widget ifCredentialField() => vk
      ? Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextFormField(
              controller: credentials,
              keyboardType: TextInputType.number,
              decoration: InputDecoration(
                labelText: l.editorCredentials,
                floatingLabelBehavior: FloatingLabelBehavior.always,
                isDense: true,
                hintText: l.editorDefault('12'),
              ),
              validator: countError,
            ),
            help(l.editorCredentialsHelp),
          ],
        )
      : const SizedBox.shrink();
}
