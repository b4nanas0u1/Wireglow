import 'package:flutter/widgets.dart';

import 'l10n/app_localizations.dart';
import 'ui_error.dart';

String uiErrorText(BuildContext context, Object error) {
  if (error is! UiError) return error.toString();
  final l = AppLocalizations.of(context)!;
  return switch (error.code) {
    UiErrorCode.lostController => l.lostController,
    UiErrorCode.waitController => l.waitController,
    UiErrorCode.startFailed => l.startFailed,
    UiErrorCode.incompleteBundle => l.incompleteBundle,
    UiErrorCode.controllerFailed => l.controllerFailed,
    UiErrorCode.insertLink => l.insertLink,
  };
}
