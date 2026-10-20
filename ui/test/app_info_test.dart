import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:wireglow_ui/about_client.dart';
import 'package:wireglow_ui/app_info.dart';
import 'package:wireglow_ui/l10n/app_localizations.dart';

void main() {
  testWidgets('About shows the build version or dev without metadata', (
    tester,
  ) async {
    const expectedVersion = String.fromEnvironment(
      'APP_VERSION',
      defaultValue: 'dev',
    );
    expect(appVersion, expectedVersion);
    await tester.pumpWidget(
      const MaterialApp(
        locale: Locale('en'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home: Scaffold(body: AboutClient()),
      ),
    );
    expect(find.text(expectedVersion), findsOneWidget);
  });
}
