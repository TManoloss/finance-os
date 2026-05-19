import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'core/theme/colors.dart';
import 'features/dashboard/presentation/dashboard_screen.dart';
import 'features/reports/presentation/replay_screen.dart';

void main() {
  runApp(
    const ProviderScope(
      child: MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    final GoRouter router = GoRouter(
      initialLocation: '/',
      routes: [
        GoRoute(
          path: '/',
          builder: (context, state) => const DashboardScreen(),
        ),
        GoRoute(
          path: '/replay/:month',
          builder: (context, state) => ReplayScreen(
            month: state.pathParameters['month'] ?? '2026-05',
          ),
        ),
      ],
    );

    return MaterialApp.router(
      title: 'Finance OS',
      theme: ThemeData(
        brightness: Brightness.dark,
        scaffoldBackgroundColor: AppColors.background,
        primaryColor: AppColors.primary,
      ),
      routerConfig: router,
    );
  }
}
