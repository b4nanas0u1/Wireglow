import 'dart:io';

import 'package:flutter/material.dart';

import 'app_info.dart';
import 'l10n/app_localizations.dart';

class AboutClient extends StatelessWidget {
  const AboutClient({super.key});

  Future<void> _openRepository(BuildContext context, String repository) async {
    final url = 'https://github.com/$repository';
    try {
      final result = Platform.isWindows
          ? await Process.run('rundll32', ['url.dll,FileProtocolHandler', url])
          : await Process.run(Platform.isMacOS ? 'open' : 'xdg-open', [url]);
      if (result.exitCode != 0) throw const ProcessException('', []);
    } on ProcessException {
      if (!context.mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.browserFailed)),
      );
    }
  }

  Widget _field(BuildContext context, String label, String value) => Padding(
    padding: const EdgeInsets.only(bottom: 20),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: TextStyle(
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        const SizedBox(height: 4),
        SelectableText(value),
      ],
    ),
  );

  Widget _engine(
    BuildContext context,
    String name,
    String version,
    String repository,
  ) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      _field(context, name, version),
      if (repository.isNotEmpty) ...[
        SelectableText('https://github.com/$repository'),
        const SizedBox(height: 8),
        TextButton.icon(
          onPressed: () => _openRepository(context, repository),
          icon: const Icon(Icons.open_in_new, size: 18),
          label: Text(AppLocalizations.of(context)!.openGithub),
        ),
      ] else
        Text(AppLocalizations.of(context)!.repositoryMissing),
      const SizedBox(height: 24),
    ],
  );

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.all(24),
    child: Align(
      alignment: Alignment.topLeft,
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 640),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(appName, style: Theme.of(context).textTheme.headlineSmall),
            const SizedBox(height: 24),
            _field(
              context,
              AppLocalizations.of(context)!.clientVersion,
              appVersion,
            ),
            _field(
              context,
              AppLocalizations.of(context)!.commit,
              appCommit.isEmpty
                  ? AppLocalizations.of(context)!.unknown
                  : appCommit,
            ),
            _field(
              context,
              AppLocalizations.of(context)!.buildTime,
              appBuildTime.isEmpty
                  ? AppLocalizations.of(context)!.unknown
                  : appBuildTime,
            ),
            const Divider(),
            const SizedBox(height: 24),
            Text(
              AppLocalizations.of(context)!.engines,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 24),
            _engine(
              context,
              'mihomo',
              mihomoVersion.isEmpty
                  ? AppLocalizations.of(context)!.unknown
                  : mihomoVersion,
              mihomoRepository,
            ),
            _engine(
              context,
              'free-turn',
              freeTurnVersion.isEmpty
                  ? AppLocalizations.of(context)!.unknown
                  : freeTurnVersion,
              freeTurnRepository,
            ),
          ],
        ),
      ),
    ),
  );
}
