import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/features/transactions/presentation/transactions_provider.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class TransactionsScreen extends ConsumerWidget {
  const TransactionsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final transactionsAsync = ref.watch(transactionsProvider);
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final dateFmt = DateFormat('dd/MM/yyyy');

    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      appBar: AppBar(
        title: const Text('EXTRATO_DE_OPERACOES'),
        actions: [
          IconButton(
            icon: const Icon(LucideIcons.filter, size: 18),
            onPressed: () {},
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: Column(
        children: [
          // Header neo-brutal
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            color: BlueprintTheme.elevated,
            child: Row(children: [
              const Icon(LucideIcons.arrowLeftRight, size: 14),
              const SizedBox(width: 8),
              Text('OPERACOES_REGISTRADAS', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10)),
            ]),
          ),
          Container(height: 2, color: BlueprintTheme.border),
          // Lista
          Expanded(
            child: transactionsAsync.when(
              data: (transactions) {
                if (transactions.isEmpty) {
                  return Center(
                    child: Padding(
                      padding: const EdgeInsets.all(40),
                      child: Text('BUFFER_EMPTY: SEM_OPERAÇÕES',
                          style: terminalLabel(), textAlign: TextAlign.center),
                    ),
                  );
                }
                return RefreshIndicator(
                  color: BlueprintTheme.accentPurple,
                  onRefresh: () async => ref.invalidate(transactionsProvider),
                  child: ListView.separated(
                    itemCount: transactions.length,
                    separatorBuilder: (_, __) => Container(height: 1, color: BlueprintTheme.border),
                    itemBuilder: (_, i) {
                      final tx = transactions[i];
                      final isCredit = tx.direction == 'credit';
                      return Container(
                        color: BlueprintTheme.surface,
                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                        child: Row(children: [
                          // ícone quadrado
                          Container(
                            width: 38, height: 38,
                            color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.danger,
                            child: Center(child: Text(isCredit ? '+' : '-',
                              style: const TextStyle(color: Colors.white, fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 18))),
                          ),
                          const SizedBox(width: 14),
                          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            Text(tx.description.toUpperCase(),
                              style: const TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12),
                              overflow: TextOverflow.ellipsis),
                            const SizedBox(height: 2),
                            Text(
                              '${DateTime.tryParse(tx.date) != null ? dateFmt.format(DateTime.parse(tx.date)) : tx.date} • ${tx.categoryName?.toUpperCase() ?? 'OUTROS'}',
                              style: terminalLabel(fontSize: 8),
                            ),
                          ])),
                          const SizedBox(width: 8),
                          Text(fmt.format(tx.amount),
                            style: moneyStyle(
                              color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.textPrimary,
                              fontSize: 13,
                            )),
                        ]),
                      );
                    },
                  ),
                );
              },
              loading: () => const Center(child: CircularProgressIndicator(color: BlueprintTheme.accentPurple)),
              error: (err, _) => Center(child: Text('ERRO: $err', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace', fontSize: 12))),
            ),
          ),
        ],
      ),
    );
  }
}
