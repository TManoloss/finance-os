import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons/lucide_icons.dart';
import '../theme/blueprint_theme.dart';

class MainLayout extends StatelessWidget {
  final StatefulNavigationShell navigationShell;

  const MainLayout({super.key, required this.navigationShell});

  void _goBranch(int index) {
    navigationShell.goBranch(
      index,
      initialLocation: index == navigationShell.currentIndex,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: Container(
        decoration: const BoxDecoration(
          border: Border(top: BorderSide(color: BlueprintTheme.border, width: 1)),
          color: BlueprintTheme.surface,
        ),
        child: BottomNavigationBar(
          currentIndex: navigationShell.currentIndex,
          onTap: _goBranch,
          items: const [
            BottomNavigationBarItem(
              icon: Icon(LucideIcons.layoutDashboard), 
              activeIcon: Icon(LucideIcons.layoutDashboard, color: BlueprintTheme.accentPurple),
              label: 'HOME'
            ),
            BottomNavigationBarItem(
              icon: Icon(LucideIcons.arrowLeftRight), 
              activeIcon: Icon(LucideIcons.arrowLeftRight, color: BlueprintTheme.accentPurple),
              label: 'EXTRATO'
            ),
            BottomNavigationBarItem(
              icon: Icon(LucideIcons.creditCard), 
              activeIcon: Icon(LucideIcons.creditCard, color: BlueprintTheme.accentPurple),
              label: 'CARTÕES'
            ),
            BottomNavigationBarItem(
              icon: Icon(LucideIcons.barChart3), 
              activeIcon: Icon(LucideIcons.barChart3, color: BlueprintTheme.accentPurple),
              label: 'RELATÓRIOS'
            ),
            BottomNavigationBarItem(
              icon: Icon(LucideIcons.terminal), 
              activeIcon: Icon(LucideIcons.terminal, color: BlueprintTheme.accentPurple),
              label: 'PIERRE'
            ),
          ],
        ),
      ),
    );
  }
}
