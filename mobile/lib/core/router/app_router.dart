import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/auth/presentation/login_screen.dart';
import 'package:finance_os/features/auth/presentation/register_screen.dart';
import 'package:finance_os/features/auth/presentation/pluggy_setup_screen.dart';
import 'package:finance_os/features/auth/presentation/auth_provider.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_screen.dart';
import 'package:finance_os/features/transactions/presentation/transactions_screen.dart';
import 'package:finance_os/features/cards/presentation/cards_screen.dart';
import 'package:finance_os/features/reports/presentation/reports_screen.dart';
import 'package:finance_os/features/chat/presentation/chat_screen.dart';
import 'package:finance_os/features/settings/presentation/settings_screen.dart';
import 'package:finance_os/features/health/presentation/health_screen.dart';
import 'package:finance_os/features/merchants/presentation/merchants_screen.dart';
import 'package:finance_os/features/simulator/presentation/simulator_screen.dart';
import 'package:finance_os/features/reports/presentation/replay_screen.dart';
import 'package:finance_os/core/layout/main_layout.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/dashboard',
    redirect: (context, state) {
      final isAuthenticated = authState.isAuthenticated;
      final isLoggingIn = state.matchedLocation == '/login';
      final isRegistering = state.matchedLocation == '/register';

      if (!isAuthenticated) {
        return (isLoggingIn || isRegistering) ? null : '/login';
      }

      // If authenticated
      final user = authState.user;
      final isPluggyConfigured = user?.isPluggyConfigured ?? false;
      final isSettingUpPluggy = state.matchedLocation == '/pluggy-setup';

      if (!isPluggyConfigured) {
        return isSettingUpPluggy ? null : '/pluggy-setup';
      }

      if (isLoggingIn || isRegistering || (isSettingUpPluggy && isPluggyConfigured)) {
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
        path: '/register',
        name: 'register',
        builder: (context, state) => const RegisterScreen(),
      ),
      GoRoute(
        path: '/pluggy-setup',
        name: 'pluggy-setup',
        builder: (context, state) => const PluggySetupScreen(),
      ),
      // Rotas protegidas que não são parte da shell (settings, health, merchants, simulator, replay)
      GoRoute(
        path: '/settings',
        name: 'settings',
        builder: (context, state) => const SettingsScreen(),
      ),
      GoRoute(
        path: '/health',
        name: 'health',
        builder: (context, state) => const HealthScreen(),
      ),
      GoRoute(
        path: '/merchants',
        name: 'merchants',
        builder: (context, state) => const MerchantsScreen(),
      ),
      GoRoute(
        path: '/simulator',
        name: 'simulator',
        builder: (context, state) => const SimulatorScreen(),
      ),
      GoRoute(
        path: '/replay/:month',
        name: 'replay',
        builder: (context, state) => ReplayScreen(month: state.pathParameters['month'] ?? ''),
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
                builder: (context, state) => const CardsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/reports',
                name: 'reports',
                builder: (context, state) => const ReportsScreen(),
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
