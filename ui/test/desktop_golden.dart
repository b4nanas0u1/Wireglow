import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

// Text rasterization differs between Windows and Linux Flutter engines.
// Keep exact pixel comparisons against a reference for the current host.
Matcher matchesDesktopGolden(String name) => matchesGoldenFile(
  Platform.isWindows ? 'goldens/windows/$name.png' : 'goldens/$name.png',
);
