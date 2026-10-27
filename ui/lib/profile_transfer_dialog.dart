import 'package:file_selector/file_selector.dart';
import 'package:flutter/material.dart';
import 'package:grpc/grpc.dart';

import 'connect_model.dart';
import 'controller/controller_repository.dart' hide State;
import 'l10n/app_localizations.dart';

/// The picker supplies paths only. Parsing, credentials and writes stay in Go.
class ProfileTransferFiles {
  const ProfileTransferFiles();
  static const types = [
    XTypeGroup(label: 'Wireglow profiles', extensions: ['json']),
  ];
  Future<String?> open() async =>
      (await openFile(acceptedTypeGroups: types))?.path;
  Future<String?> save() async => (await getSaveLocation(
    acceptedTypeGroups: types,
    suggestedName: 'wireglow-profiles.json',
  ))?.path;
}

class ProfileTransferDialog extends StatefulWidget {
  const ProfileTransferDialog({
    super.key,
    required this.model,
    required this.exporting,
    this.files = const ProfileTransferFiles(),
  });
  final ConnectModel model;
  final bool exporting;
  final ProfileTransferFiles files;
  @override
  State<ProfileTransferDialog> createState() => _ProfileTransferDialogState();
}

class _ProfileTransferDialogState extends State<ProfileTransferDialog> {
  final selected = <String>{};
  final choices = <int, ProfileImportAction>{};
  ProfileFilePreview? preview;
  ImportProfileFileResponse? result;
  String? fileName;
  String? error;
  bool pending = false;
  late final List<ProfileSummary> profiles;

  @override
  void initState() {
    super.initState();
    profiles = List.of(widget.model.profiles);
    selected.addAll(profiles.map((p) => p.id));
  }

  String failure(Object e) {
    final l = AppLocalizations.of(context)!;
    if (e is GrpcError) {
      if (e.code == StatusCode.invalidArgument) return l.profileArchiveInvalid;
      if (e.code == StatusCode.failedPrecondition) {
        return l.profilePreviewExpired;
      }
      if (e.code == StatusCode.unavailable) return l.lostController;
    }
    return l.profileTransferFailed;
  }

  Future<void> chooseFile() async {
    setState(() {
      pending = true;
      error = null;
    });
    try {
      final path = await widget.files.open();
      if (path == null || !mounted) return;
      // Clear the old selection before requesting another preview.
      setState(() {
        preview = null;
        choices.clear();
        result = null;
      });
      final value = await widget.model.repository.previewProfileFile(path);
      if (!mounted) return;
      setState(() {
        preview = value;
        fileName = path.split(RegExp(r'[/\\]')).last;
        for (var i = 0; i < value.entries.length; i++) {
          choices[i] = value.entries[i].hasExisting()
              ? ProfileImportAction.PROFILE_IMPORT_ACTION_SKIP
              : ProfileImportAction.PROFILE_IMPORT_ACTION_ADD;
        }
      });
    } catch (e) {
      if (mounted) setState(() => error = failure(e));
    } finally {
      if (mounted) setState(() => pending = false);
    }
  }

  Future<void> export() async {
    setState(() {
      pending = true;
      error = null;
    });
    try {
      final path = await widget.files.save();
      if (path == null || !mounted) return;
      await widget.model.repository.exportProfiles(selected.toList(), path);
      if (!mounted) return;
      Navigator.pop(
        context,
        AppLocalizations.of(context)!.profilesExported(selected.length),
      );
    } catch (e) {
      if (mounted) setState(() => error = failure(e));
    } finally {
      if (mounted) setState(() => pending = false);
    }
  }

  Future<void> import() async {
    setState(() {
      pending = true;
      error = null;
    });
    try {
      final response = await widget.model.importProfileFile(
        preview!.token,
        choices.entries
            .map((e) => ProfileImportChoice(index: e.key, action: e.value))
            .toList(),
      );
      if (!mounted) return;
      final failed = response.results.any(
        (r) => r.status == 'failed' || r.status == 'changed',
      );
      if (failed) {
        setState(() => result = response);
      } else {
        Navigator.pop(
          context,
          AppLocalizations.of(context)!.profilesImported(
            response.results.where((r) => r.status == 'imported').length,
          ),
        );
      }
    } catch (e) {
      // A lost reply may have followed a successful write. A fresh preview is required.
      if (mounted) {
        setState(() {
          error = failure(e);
          preview = null;
          choices.clear();
        });
      }
    } finally {
      if (mounted) setState(() => pending = false);
    }
  }

  int get importCount => choices.values
      .where((a) => a != ProfileImportAction.PROFILE_IMPORT_ACTION_SKIP)
      .length;

  Widget row(Widget title, Widget subtitle, Widget control) => LayoutBuilder(
    builder: (context, box) {
      final text = Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [title, const SizedBox(height: 4), subtitle],
      );
      if (box.maxWidth < 380) {
        return Padding(
          padding: const EdgeInsets.symmetric(vertical: 10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [text, const SizedBox(height: 8), control],
          ),
        );
      }
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 10),
        child: Row(
          children: [
            Expanded(child: text),
            const SizedBox(width: 16),
            control,
          ],
        ),
      );
    },
  );

  Widget importContents(AppLocalizations l) {
    final value = preview;
    if (value == null) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(l.importProfilesHelp),
          const SizedBox(height: 24),
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              border: Border.all(
                color: Theme.of(context).colorScheme.outlineVariant,
              ),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              children: [
                const Icon(Icons.insert_drive_file_outlined, size: 28),
                const SizedBox(height: 12),
                Text(l.profileFileType),
                const SizedBox(height: 16),
                OutlinedButton.icon(
                  key: const Key('choose-profile-file'),
                  onPressed: pending ? null : chooseFile,
                  icon: const Icon(Icons.folder_open_outlined),
                  label: Text(l.chooseProfileFile),
                ),
              ],
            ),
          ),
        ],
      );
    }
    if (result != null) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(l.profileImportPartial),
          const SizedBox(height: 12),
          for (final item in result!.results)
            ListTile(
              contentPadding: EdgeInsets.zero,
              title: Text(value.entries[item.index].profile.name),
              subtitle: Text(switch (item.status) {
                'imported' => l.profileImportDone,
                'skipped' => l.profileImportSkipped,
                'changed' => l.profileImportChanged,
                _ => l.profileImportFailed,
              }),
            ),
        ],
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(Icons.insert_drive_file_outlined),
            const SizedBox(width: 12),
            Expanded(child: Text(fileName ?? l.profileFileType)),
            TextButton(
              onPressed: pending ? null : chooseFile,
              child: Text(l.changeProfileFile),
            ),
          ],
        ),
        const Divider(),
        Text(
          l.profileCount(value.entries.length),
          style: Theme.of(context).textTheme.bodySmall,
        ),
        for (var i = 0; i < value.entries.length; i++)
          row(
            Text(
              value.entries[i].profile.name.isEmpty
                  ? l.unnamed
                  : value.entries[i].profile.name,
            ),
            Text(
              value.entries[i].hasExisting()
                  ? l.profileAlreadyExists(value.entries[i].existing.name)
                  : value.entries[i].profile.provider,
              style: Theme.of(context).textTheme.bodySmall,
            ),
            value.entries[i].hasExisting()
                ? DropdownButton<ProfileImportAction>(
                    key: Key('profile-import-action-$i'),
                    value: choices[i],
                    items: [
                      DropdownMenuItem(
                        value: ProfileImportAction.PROFILE_IMPORT_ACTION_SKIP,
                        child: Text(l.skipExistingProfile),
                      ),
                      DropdownMenuItem(
                        value: ProfileImportAction.PROFILE_IMPORT_ACTION_COPY,
                        child: Text(l.copyImportedProfile),
                      ),
                      DropdownMenuItem(
                        value:
                            ProfileImportAction.PROFILE_IMPORT_ACTION_REPLACE,
                        child: Text(l.replaceExistingProfile),
                      ),
                    ],
                    onChanged: pending
                        ? null
                        : (v) => setState(() => choices[i] = v!),
                  )
                : Checkbox(
                    key: Key('profile-import-selected-$i'),
                    semanticLabel: value.entries[i].profile.name,
                    value:
                        choices[i] !=
                        ProfileImportAction.PROFILE_IMPORT_ACTION_SKIP,
                    onChanged: pending
                        ? null
                        : (v) => setState(
                            () => choices[i] = v!
                                ? ProfileImportAction.PROFILE_IMPORT_ACTION_ADD
                                : ProfileImportAction
                                      .PROFILE_IMPORT_ACTION_SKIP,
                          ),
                  ),
          ),
        const Divider(),
        Text(
          l.profileImportReviewHelp,
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context)!;
    return PopScope(
      canPop: !pending,
      child: AlertDialog(
        title: Row(
          children: [
            Expanded(
              child: Text(
                widget.exporting ? l.exportProfiles : l.importProfiles,
              ),
            ),
            IconButton(
              tooltip: l.cancel,
              onPressed: pending ? null : () => Navigator.pop(context),
              icon: const Icon(Icons.close),
            ),
          ],
        ),
        content: SizedBox(
          width: 512,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (widget.exporting) ...[
                  Text(l.exportProfilesHelp),
                  const SizedBox(height: 12),
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    controlAffinity: ListTileControlAffinity.leading,
                    title: Text(l.selectAllProfiles),
                    subtitle: Text(l.profileCount(selected.length)),
                    value: selected.length == profiles.length,
                    onChanged: pending
                        ? null
                        : (v) => setState(() {
                            selected.clear();
                            if (v!) selected.addAll(profiles.map((p) => p.id));
                          }),
                  ),
                  const Divider(),
                  for (final p in profiles)
                    CheckboxListTile(
                      key: Key('export-profile-${p.id}'),
                      contentPadding: EdgeInsets.zero,
                      controlAffinity: ListTileControlAffinity.leading,
                      title: Text(p.name.isEmpty ? l.unnamed : p.name),
                      subtitle: Text(p.provider),
                      value: selected.contains(p.id),
                      onChanged: pending
                          ? null
                          : (v) => setState(() {
                              v! ? selected.add(p.id) : selected.remove(p.id);
                            }),
                    ),
                  const Divider(),
                  const SizedBox(height: 8),
                  Text(
                    l.profileExportSecretsHelp,
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ] else
                  importContents(l),
                if (pending) ...[
                  const SizedBox(height: 16),
                  const LinearProgressIndicator(),
                ],
                if (error != null) ...[
                  const SizedBox(height: 12),
                  Semantics(
                    liveRegion: true,
                    child: Text(
                      error!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
        actions: [
          if (!widget.exporting && result != null)
            TextButton(
              onPressed: pending ? null : chooseFile,
              child: Text(l.chooseProfileFile),
            ),
          TextButton(
            onPressed: pending ? null : () => Navigator.pop(context),
            child: Text(result == null ? l.cancel : l.closeProfileTransfer),
          ),
          if (widget.exporting)
            FilledButton(
              key: const Key('confirm-profile-export'),
              onPressed: pending || selected.isEmpty ? null : export,
              child: Text(l.exportProfileCount(selected.length)),
            )
          else if (preview != null && result == null)
            FilledButton(
              key: const Key('confirm-profile-import'),
              onPressed: pending || importCount == 0 ? null : import,
              child: Text(l.importProfileCount(importCount)),
            ),
        ],
      ),
    );
  }
}

class ProfileTransferSettings extends StatelessWidget {
  const ProfileTransferSettings({
    super.key,
    required this.model,
    this.files = const ProfileTransferFiles(),
  });
  final ConnectModel model;
  final ProfileTransferFiles files;
  Future<void> open(BuildContext context, bool exporting) async {
    final message = await showDialog<String>(
      context: context,
      barrierDismissible: false,
      builder: (_) => ProfileTransferDialog(
        model: model,
        exporting: exporting,
        files: files,
      ),
    );
    if (message != null && context.mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l = AppLocalizations.of(context)!;
    return ListenableBuilder(
      listenable: model,
      builder: (context, _) {
        final enabled =
            model.available &&
            !model.commandPending &&
            !model.connecting &&
            !model.stopping;
        Widget action(String title, String help, bool exporting) =>
            LayoutBuilder(
              builder: (context, box) {
                final text = Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title),
                    const SizedBox(height: 4),
                    Text(help, style: Theme.of(context).textTheme.bodySmall),
                  ],
                );
                final button = OutlinedButton.icon(
                  key: Key(exporting ? 'export-profiles' : 'import-profiles'),
                  onPressed:
                      enabled && (!exporting || model.profiles.isNotEmpty)
                      ? () => open(context, exporting)
                      : null,
                  icon: Icon(
                    exporting
                        ? Icons.file_upload_outlined
                        : Icons.file_download_outlined,
                    size: 18,
                  ),
                  label: Text(
                    exporting ? l.exportProfilesAction : l.importProfilesAction,
                  ),
                );
                if (box.maxWidth < 420) {
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [text, const SizedBox(height: 12), button],
                  );
                }
                return Row(
                  children: [
                    Expanded(child: text),
                    const SizedBox(width: 16),
                    button,
                  ],
                );
              },
            );
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(l.profileTransferHelp),
            const SizedBox(height: 20),
            action(l.importProfiles, l.importProfilesSettingsHelp, false),
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 8),
              child: Divider(),
            ),
            action(l.exportProfiles, l.exportProfilesSettingsHelp, true),
          ],
        );
      },
    );
  }
}
