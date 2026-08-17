import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class MoreScreen extends StatelessWidget {
  const MoreScreen({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Mais')),
    body: ListView(
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 104),
      children: [
        Text('FinanceOS', style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w700)),
        const SizedBox(height: 4),
        Text('Inteligência, ferramentas e configurações.', style: terminalLabel(fontSize: 12)),
        const SizedBox(height: 22),
        _MenuSection('Inteligência', [
          _MenuItem('Pierre', 'Converse com seu assistente financeiro', LucideIcons.messageCircle, '/chat'),
          _MenuItem('Relatórios', 'Insights e análises do seu comportamento', LucideIcons.sparkles, '/reports'),
          _MenuItem('Saúde financeira', 'Score e recomendações', LucideIcons.heartPulse, '/health'),
          _MenuItem('Simulador', 'Veja o impacto de decisões', LucideIcons.calculator, '/simulator'),
        ]),
        const SizedBox(height: 18),
        _MenuSection('Dados', [
          _MenuItem('Estabelecimentos', 'Seus principais destinos de gastos', LucideIcons.store, '/merchants'),
          _MenuItem('Configurações', 'Contas, sincronização e integrações', LucideIcons.settings, '/settings'),
        ]),
      ],
    ),
  );
}

class _MenuSection extends StatelessWidget {
  final String title; final List<_MenuItem> items;
  const _MenuSection(this.title, this.items);
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(title, style: terminalLabel(fontSize: 11)),
      const SizedBox(height: 9),
      Container(
        decoration: neoBrutalCard(),
        child: Column(
          children: List.generate(
            items.length,
            (index) => Column(children: [
              ListTile(
                leading: Icon(items[index].icon, size: 19),
                title: Text(items[index].title, style: const TextStyle(fontWeight: FontWeight.w600)),
                subtitle: Text(items[index].subtitle, style: terminalLabel(fontSize: 10)),
                trailing: const Icon(LucideIcons.chevronRight, size: 17),
                onTap: () => context.push(items[index].route),
              ),
              if (index < items.length - 1) const Divider(indent: 56),
            ]),
          ),
        ),
      ),
    ],
  );
}

class _MenuItem { final String title, subtitle, route; final IconData icon; const _MenuItem(this.title, this.subtitle, this.icon, this.route); }
