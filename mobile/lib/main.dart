import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

void main() {
  runApp(
    const ProviderScope(
      child: FinanceOSApp(),
    ),
  );
}

class FinanceOSApp extends StatelessWidget {
  const FinanceOSApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'FINANCE_OS',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        textTheme: GoogleFonts.courierPrimeTextTheme(),
      ),
      home: const Scaffold(
        body: Center(child: Text('FINANCE_OS_BOOT...')),
      ),
    );
  }
}
