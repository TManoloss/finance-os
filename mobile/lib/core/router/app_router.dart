import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/auth/presentation/login_screen.dart';
import 'package:finance_os/features/auth/presentation/register_screen.dart';
import 'package:finance_os/features/auth/presentation/auth_provider.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_screen.dart';
import 'package:finance_os/features/transactions/presentation/transactions_screen.dart';
import 'package:finance_os/features/cards/presentation/cards_screen.dart';
import 'package:finance_os/features/cards/presentation/pluggy_connect_screen.dart';
import 'package:finance_os/features/reports/presentation/reports_screen.dart';
import 'package:finance_os/features/chat/presentation/chat_screen.dart';
import 'package:finance_os/features/settings/presentation/settings_screen.dart';
import 'package:finance_os/features/health/presentation/health_screen.dart';
import 'package:finance_os/features/merchants/presentation/merchants_screen.dart';
import 'package:finance_os/features/simulator/presentation/simulator_screen.dart';
import 'package:finance_os/features/reports/presentation/replay_screen.dart';
import 'package:finance_os/features/goals/presentation/goals_screen.dart';
import 'package:finance_os/features/more/presentation/more_screen.dart';
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
      if (isLoggingIn ||
          isRegistering ||
          state.matchedLocation == '/pluggy-setup') {
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
      // Rotas protegidas que não são parte da shell (settings, health, merchants, simulator, replay)
      GoRoute(
        path: '/settings',
        name: 'settings',
        builder: (context, state) => const SettingsScreen(),
      ),
      GoRoute(
        path: '/pluggy-connect',
        name: 'pluggy-connect',
        builder: (context, state) =>
            PluggyConnectScreen(connectToken: state.extra! as String),
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
        builder: (context, state) =>
            ReplayScreen(month: state.pathParameters['month'] ?? ''),
      ),
      GoRoute(
        path: '/reports',
        name: 'reports',
        builder: (context, state) => const ReportsScreen(),
      ),
      GoRoute(
        path: '/chat',
        name: 'chat',
        builder: (context, state) => const ChatScreen(),
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
                path: '/cards',
                name: 'cards',
                builder: (context, state) => const CardsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/goals',
                name: 'goals',
                builder: (context, state) => const GoalsScreen(),
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
                path: '/more',
                name: 'more',
                builder: (context, state) => const MoreScreen(),
              ),
            ],
          ),
        ],
      ),
    ],
  );
});
