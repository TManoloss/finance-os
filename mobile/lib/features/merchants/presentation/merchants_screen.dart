import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';

class MerchantsScreen extends ConsumerWidget {
  const MerchantsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final merchantsAsync = ref.watch(merchantsProvider);
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      appBar: AppBar(title: const Text('PRINCIPAIS_DESTINOS_DE_GASTOS')),
      body: Column(
        children: [
          Container(
            color: BlueprintTheme.elevated,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(children: [
              const Icon(LucideIcons.trendingUp, size: 14),
              const SizedBox(width: 8),
              Text('ANÁLISE_DE_MERCHANTS', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10)),
            ]),
          ),
          Container(height: 2, color: BlueprintTheme.border),
          Expanded(
            child: merchantsAsync.when(
              data: (merchants) {
                if (merchants.isEmpty) {
                  return Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                    const Icon(LucideIcons.store, size: 48, color: BlueprintTheme.textSecondary),
                    const SizedBox(height: 12),
                    Text('SEM_DADOS_DE_MERCHANTS', style: terminalLabel()),
                  ]));
                }
                return RefreshIndicator(
                  color: BlueprintTheme.accentPurple,
                  onRefresh: () async => ref.refresh(merchantsProvider),
                  child: ListView.separated(
                    itemCount: merchants.length,
                    separatorBuilder: (_, __) => Container(height: 1, color: BlueprintTheme.border),
                    itemBuilder: (_, index) {
                      final m = merchants[index];
                      final name = (m['merchant'] ?? m['merchant_name'] ?? '?').toString();
                      final total = (m['total'] as num?)?.toDouble() ?? 0;
                      final count = (m['count'] as num?)?.toInt() ?? 0;
                      final avgTicket = count > 0 ? total / count : 0.0;

                      return Container(
                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                        color: BlueprintTheme.surface,
                        child: Row(children: [
                          // Ranking badge quadrado
                          Container(
                            width: 32, height: 32,
                            color: BlueprintTheme.elevated,
                            child: Center(child: Text('${index + 1}', style: const TextStyle(
                              fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12,
                            ))),
                          ),
                          const SizedBox(width: 14),
                          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            Text(name.toUpperCase(), style: const TextStyle(
                              fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12,
                            ), overflow: TextOverflow.ellipsis),
                            const SizedBox(height: 4),
                            Row(children: [
                              _tag('$count OPERAÇÕES'),
                              const SizedBox(width: 6),
                              _tag('TICKET: ${fmt.format(avgTicket)}'),
                            ]),
                          ])),
                          const SizedBox(width: 10),
                          Text(fmt.format(total), style: moneyStyle(color: BlueprintTheme.danger, fontSize: 14)),
                        ]),
                      );
                    },
                  ),
                );
              },
              loading: () => const Center(child: CircularProgressIndicator(color: BlueprintTheme.accentPurple)),
              error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace'))),
            ),
          ),
        ],
      ),
    );
  }

  Widget _tag(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      color: BlueprintTheme.elevated,
      child: Text(label, style: terminalLabel(fontSize: 8)),
    );
  }
}
