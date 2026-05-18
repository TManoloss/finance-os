import 'package:flutter/material.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/core/router/app_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

void main() {
  runApp(
    const ProviderScope(
      child: FinanceOS(),
    ),
  );
}

class FinanceOS extends ConsumerWidget {
  const FinanceOS({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'FINANCE_OS',
      debugShowCheckedModeBanner: false,
      theme: BlueprintTheme.dark, // Alterado para dark
      routerConfig: router,
    );
  }
}
