import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class MainLayout extends StatelessWidget {
  final StatefulNavigationShell navigationShell;

  const MainLayout({super.key, required this.navigationShell});

  void _goBranch(int index) => navigationShell.goBranch(
        index,
        initialLocation: index == navigationShell.currentIndex,
      );

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      body: navigationShell,
      extendBody: true,
      bottomNavigationBar: SafeArea(
        minimum: const EdgeInsets.fromLTRB(24, 0, 24, 10),
        child: _FloatingNav(
          currentIndex: navigationShell.currentIndex,
          onSelect: _goBranch,
        ),
      ),
    );
  }
}

class _FloatingNav extends StatelessWidget {
  final int currentIndex;
  final ValueChanged<int> onSelect;

  const _FloatingNav({required this.currentIndex, required this.onSelect});

  @override
  Widget build(BuildContext context) {
    const accent = Color(0xFFFFA58E);
    const items = [
      (LucideIcons.home, 'Home'),
      (LucideIcons.barChart3, 'Portfolio'),
      (LucideIcons.flame, 'Goal'),
      (LucideIcons.history, 'History'),
      (LucideIcons.settings, 'Settings'),
    ];

    return Container(
      height: 58,
      decoration: BoxDecoration(
        color: const Color(0xFF211C1C),
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: accent.withValues(alpha: .72)),
        boxShadow: const [BoxShadow(color: Color(0x66FF9F89), blurRadius: 12, offset: Offset(0, 5))],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: List.generate(items.length, (index) {
          final selected = index == currentIndex;
          // Goal e History não usam o marcador laranja lateral/label da referência.
          final showAccent = selected && index != 2 && index != 3;
          final item = items[index];
          return Expanded(
            child: InkWell(
              borderRadius: BorderRadius.circular(18),
              onTap: () => onSelect(index),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(item.$1, size: 18, color: showAccent ? accent : BlueprintTheme.textSecondary),
                  const SizedBox(height: 2),
                  SizedBox(
                    height: 12,
                    child: Text(
                      showAccent ? item.$2 : '',
                      style: const TextStyle(color: accent, fontSize: 8, fontWeight: FontWeight.w700),
                    ),
                  ),
                ],
              ),
            ),
          );
        }),
      ),
    );
  }
}
