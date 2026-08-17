import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';

/// The mobile visual language: dark, dense and intentionally quiet.
/// Kept under the existing name so feature code can migrate without churn.
class BlueprintTheme {
  static const Color background = Color(0xFF000000);
  static const Color surface = Color(0xFF121212);
  static const Color elevated = Color(0xFF1B1B1D);
  static const Color border = Color(0xFF2A2A2E);
  static const Color textPrimary = Color(0xFFF4F4F5);
  static const Color textSecondary = Color(0xFF919198);
  static const Color accentPurple = Color(0xFF8B7CFF);
  static const Color accentTeal = Color(0xFF35C98B);
  static const Color danger = Color(0xFFFF6B6B);
  static const Color warning = Color(0xFFF5A524);
  static const Color success = Color(0xFF35C98B);

  static ThemeData get dark => ThemeData(
        useMaterial3: true,
        brightness: Brightness.dark,
        scaffoldBackgroundColor: background,
        colorScheme: const ColorScheme.dark(
          primary: accentPurple,
          secondary: accentTeal,
          surface: surface,
          onSurface: textPrimary,
          error: danger,
          outline: border,
        ),
        textTheme: GoogleFonts.interTextTheme(ThemeData.dark().textTheme).apply(
          bodyColor: textPrimary,
          displayColor: textPrimary,
        ),
        appBarTheme: const AppBarTheme(
          backgroundColor: background,
          elevation: 0,
          surfaceTintColor: Colors.transparent,
          systemOverlayStyle: SystemUiOverlayStyle.light,
          titleTextStyle: TextStyle(fontSize: 17, fontWeight: FontWeight.w700, color: textPrimary),
          iconTheme: IconThemeData(color: textPrimary),
        ),
        cardTheme: CardThemeData(
          color: surface,
          elevation: 0,
          margin: EdgeInsets.zero,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        ),
        inputDecorationTheme: InputDecorationTheme(
          filled: true,
          fillColor: elevated,
          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 15),
          hintStyle: const TextStyle(color: textSecondary),
          border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
          enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
          focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: accentPurple)),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: textPrimary,
            foregroundColor: background,
            minimumSize: const Size.fromHeight(48),
            elevation: 0,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            textStyle: const TextStyle(fontWeight: FontWeight.w700),
          ),
        ),
        dividerTheme: const DividerThemeData(color: border, thickness: 1),
        snackBarTheme: SnackBarThemeData(
          backgroundColor: elevated,
          contentTextStyle: const TextStyle(color: textPrimary),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          behavior: SnackBarBehavior.floating,
        ),
      );

  // Compatibility for the existing feature tree while it is being redesigned.
  static ThemeData get light => dark;
}

BoxDecoration neoBrutalCard({
  Color backgroundColor = BlueprintTheme.surface,
  double shadowOffset = 0,
  Color? borderColor,
}) => BoxDecoration(
  color: backgroundColor,
  borderRadius: BorderRadius.circular(16),
  border: Border.all(color: borderColor ?? BlueprintTheme.border),
);

BoxDecoration neoBrutalElevated({Color? borderColor}) => neoBrutalCard(
  backgroundColor: BlueprintTheme.elevated,
  borderColor: borderColor,
);

TextStyle terminalLabel({Color? color, double fontSize = 10}) => TextStyle(
  fontSize: fontSize,
  fontWeight: FontWeight.w600,
  letterSpacing: .2,
  color: color ?? BlueprintTheme.textSecondary,
);

TextStyle moneyStyle({Color? color, double fontSize = 20}) => TextStyle(
  fontSize: fontSize,
  fontWeight: FontWeight.w700,
  letterSpacing: -.8,
  color: color ?? BlueprintTheme.textPrimary,
);
