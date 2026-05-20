import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class BlueprintTheme {
  // Cores do Visual Dark solicitado no Next.js
  static const Color background = Color(0xFF0A0A0F);
  static const Color surface = Color(0xFF111118);
  static const Color elevated = Color(0xFF1A1A24);
  static const Color border = Color(0xFF2A2A3A);
  static const Color textPrimary = Color(0xFFF0F0F5);
  static const Color textSecondary = Color(0xFF8888A0);
  static const Color accentPurple = Color(0xFF7C6FFF);
  static const Color accentTeal = Color(0xFF4ECDC4);
  static const Color danger = Color(0xFFFF6B6B);
  static const Color warning = Color(0xFFFFD93D);
  static const Color success = Color(0xFF6BCB77);

  static ThemeData get dark {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      scaffoldBackgroundColor: background,
      colorScheme: const ColorScheme.dark(
        primary: accentPurple,
        secondary: accentTeal,
        surface: surface,
        onSurface: textPrimary,
        error: danger,
      ),
      textTheme: GoogleFonts.interTextTheme(
        ThemeData.dark().textTheme,
      ).apply(
        bodyColor: textPrimary,
        displayColor: textPrimary,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: background,
        elevation: 0,
        centerTitle: false,
        titleTextStyle: TextStyle(
          color: textPrimary,
          fontWeight: FontWeight.w900,
          fontSize: 18,
          letterSpacing: -0.5,
        ),
        iconTheme: IconThemeData(color: textPrimary),
      ),
      bottomNavigationBarTheme: const BottomNavigationBarThemeData(
        backgroundColor: surface,
        selectedItemColor: accentPurple,
        unselectedItemColor: textSecondary,
        selectedLabelStyle: TextStyle(fontWeight: FontWeight.w900, fontSize: 10),
        unselectedLabelStyle: TextStyle(fontWeight: FontWeight.bold, fontSize: 10),
        type: BottomNavigationBarType.fixed,
        elevation: 0,
      ),
      cardTheme: CardThemeData(
        color: surface,
        elevation: 0,
        shape: RoundedRectangleBorder(
          side: const BorderSide(color: border, width: 1),
          borderRadius: BorderRadius.circular(16),
        ),
      ),
      chipTheme: ChipThemeData(
        backgroundColor: elevated,
        selectedColor: accentPurple,
        secondarySelectedColor: accentPurple,
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        labelStyle: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
        secondaryLabelStyle: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
        brightness: Brightness.dark,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: const BorderSide(color: border, width: 1),
        ),
      ),
    );
  }
}
