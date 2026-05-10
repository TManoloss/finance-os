import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/features/auth/presentation/login_screen.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_screen.dart';

final appRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      name: 'login',
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: '/dashboard',
      name: 'dashboard',
      builder: (context, state) => const DashboardScreen(),
    ),
  ],
);
