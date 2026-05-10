import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/core/router/app_router.dart';

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
    return MaterialApp.router(
      title: 'FINANCE_OS',
      debugShowCheckedModeBanner: false,
      theme: BlueprintTheme.light,
      routerConfig: appRouter,
    );
  }
}
