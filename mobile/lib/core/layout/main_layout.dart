import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons/lucide_icons.dart';

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
        decoration: BoxDecoration(
          border: const Border(top: BorderSide(color: Colors.black, width: 2)),
          color: const Color(0xFFE8E5DE),
        ),
        child: BottomNavigationBar(
          currentIndex: navigationShell.currentIndex,
          onTap: _goBranch,
          backgroundColor: Colors.transparent,
          elevation: 0,
          selectedItemColor: const Color(0xFF0000FF), // Accent Blue
          unselectedItemColor: Colors.black,
          showSelectedLabels: true,
          showUnselectedLabels: true,
          selectedLabelStyle: const TextStyle(fontWeight: FontWeight.w900, fontSize: 10),
          unselectedLabelStyle: const TextStyle(fontWeight: FontWeight.bold, fontSize: 10),
          type: BottomNavigationBarType.fixed,
          items: const [
            BottomNavigationBarItem(icon: Icon(LucideIcons.layoutDashboard), label: 'DASHBOARD'),
            BottomNavigationBarItem(icon: Icon(LucideIcons.arrowLeftRight), label: 'TRANSACTIONS'),
            BottomNavigationBarItem(icon: Icon(LucideIcons.creditCard), label: 'CREDIT'),
            BottomNavigationBarItem(icon: Icon(LucideIcons.terminal), label: 'PIERRE_AI'),
          ],
        ),
      ),
    );
  }
}
