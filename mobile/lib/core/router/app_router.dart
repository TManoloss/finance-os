import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/features/auth/presentation/login_screen.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_screen.dart';
import 'package:finance_os/core/layout/main_layout.dart';

final appRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      name: 'login',
      builder: (context, state) => const LoginScreen(),
    ),
    StatefulShellRoute.indexedStack(
      builder: (context, state, navigationShell) {
        return MainLayout(navigationShell: navigationShell);
      },
      branches: [
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/dashboard',
              name: 'dashboard',
              builder: (context, state) => const DashboardScreen(),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/transactions',
              name: 'transactions',
              builder: (context, state) => const Scaffold(body: Center(child: Text('TRANSACTIONS_MODULE'))),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/cards',
              name: 'cards',
              builder: (context, state) => const Scaffold(body: Center(child: Text('CREDIT_MODULE'))),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/chat',
              name: 'chat',
              builder: (context, state) => const Scaffold(body: Center(child: Text('PIERRE_AI_MODULE'))),
            ),
          ],
        ),
      ],
    ),
  ],
);
