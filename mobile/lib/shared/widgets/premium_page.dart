import 'package:flutter/material.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class PremiumPage extends StatelessWidget {
  final String title;
  final List<Widget> children;
  final List<Widget>? actions;
  final Widget? floatingActionButton;
  const PremiumPage({super.key, required this.title, required this.children, this.actions, this.floatingActionButton});

  @override
  Widget build(BuildContext context) => Scaffold(floatingActionButton: floatingActionButton,
    appBar: AppBar(title: Text(title), actions: actions),
    body: ListView(padding: const EdgeInsets.fromLTRB(20, 12, 20, 104), children: children),
  );
}

class PremiumCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry padding;
  const PremiumCard({super.key, required this.child, this.padding = const EdgeInsets.all(16)});
  @override
  Widget build(BuildContext context) => Container(
    padding: padding,
    decoration: BoxDecoration(color: BlueprintTheme.surface, borderRadius: BorderRadius.circular(16), border: Border.all(color: BlueprintTheme.border)),
    child: child,
  );
}

class PremiumTitle extends StatelessWidget {
  final String title;
  final String? subtitle;
  const PremiumTitle({super.key, required this.title, this.subtitle});
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 16),
    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text(title, style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w700)),
      if (subtitle != null) ...[const SizedBox(height: 4), Text(subtitle!, style: terminalLabel(fontSize: 12))],
    ]),
  );
}
