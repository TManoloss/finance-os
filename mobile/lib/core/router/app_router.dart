import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/auth/presentation/login_screen.dart';
import 'package:finance_os/features/auth/presentation/pluggy_setup_screen.dart';
import 'package:finance_os/features/auth/presentation/auth_provider.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_screen.dart';
import 'package:finance_os/features/transactions/presentation/transactions_screen.dart';
import 'package:finance_os/features/chat/presentation/chat_screen.dart';
import 'package:finance_os/core/layout/main_layout.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/dashboard',
    refreshListenable: null, // We could use a Listenable here, but ref.watch already triggers rebuild
    redirect: (context, state) {
      final isAuthenticated = authState.isAuthenticated;
      final isLoggingIn = state.matchedLocation == '/login';

      if (!isAuthenticated) {
        return isLoggingIn ? null : '/login';
      }

      // If authenticated
      final user = authState.user;
      final isPluggyConfigured = user?.isPluggyConfigured ?? false;
      final isSettingUpPluggy = state.matchedLocation == '/pluggy-setup';

      if (!isPluggyConfigured) {
        return isSettingUpPluggy ? null : '/pluggy-setup';
      }

      if (isLoggingIn || (isSettingUpPluggy && isPluggyConfigured)) {
        return '/dashboard';
      }

      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/pluggy-setup',
        name: 'pluggy-setup',
        builder: (context, state) => const PluggySetupScreen(),
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
                builder: (context, state) => const TransactionsScreen(),
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
                builder: (context, state) => const ChatScreen(),
              ),
            ],
          ),
        ],
      ),
    ],
  );
});
