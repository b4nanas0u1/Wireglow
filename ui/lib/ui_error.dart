enum UiErrorCode {
  lostController,
  waitController,
  startFailed,
  incompleteBundle,
  controllerFailed,
  insertLink,
}

class UiError implements Exception {
  const UiError(this.code);
  final UiErrorCode code;
}
