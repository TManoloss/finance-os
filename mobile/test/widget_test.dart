import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  GoogleFonts.config.allowRuntimeFetching = false;

  test('uses the dark premium color scheme', () {
    final theme = BlueprintTheme.dark;

    expect(theme.brightness, Brightness.dark);
    expect(theme.scaffoldBackgroundColor, BlueprintTheme.background);
    expect(theme.colorScheme.primary, BlueprintTheme.accentPurple);
  });
}
