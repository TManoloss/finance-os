import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/core/router/app_router.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

void main() {
  runApp(
    const ProviderScope(
      child: MyApp(),
    ),
  );
}

class MyApp extends ConsumerWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Finance OS',
      theme: BlueprintTheme.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
