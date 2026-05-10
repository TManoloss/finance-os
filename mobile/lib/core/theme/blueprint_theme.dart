import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class BlueprintTheme {
  static const Color background = Color(0xFFF4F1EA);
  static const Color border = Color(0xFF000000);
  static const Color elevated = Color(0xFFE8E5DE);
  static const Color accent = Color(0xFF0000FF);
  static const Color danger = Color(0xFFD00000);
  static const Color success = Color(0xFF008000);

  static ThemeData get light {
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: background,
      colorScheme: ColorScheme.fromSeed(
        seedColor: accent,
        surface: background,
        onSurface: border,
        primary: accent,
        error: danger,
      ),
      textTheme: GoogleFonts.courierPrimeTextTheme().apply(
        bodyColor: border,
        displayColor: border,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: elevated,
        elevation: 0,
        shape: Border(bottom: BorderSide(color: border, width: 2)),
        titleTextStyle: TextStyle(color: border, fontWeight: FontWeight.bold, fontSize: 18),
      ),
      cardTheme: CardTheme(
        color: background,
        elevation: 0,
        shape: RoundedRectangleBorder(
          side: const BorderSide(color: border, width: 2),
          borderRadius: BorderRadius.zero,
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: border,
          foregroundColor: Colors.white,
          shape: const RoundedRectangleBorder(borderRadius: BorderRadius.zero),
          padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 24),
          textStyle: const TextStyle(fontWeight: FontWeight.bold, letterSpacing: 1.2),
        ),
      ),
    );
  }
}
