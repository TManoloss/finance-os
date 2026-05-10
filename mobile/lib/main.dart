import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/theme/blueprint_theme.dart';

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
      theme: BlueprintTheme.light,
      home: const Scaffold(
        body: Center(child: Text('FINANCE_OS_BOOT...')),
      ),
    );
  }
}
