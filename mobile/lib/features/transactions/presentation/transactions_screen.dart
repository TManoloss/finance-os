import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../dashboard/presentation/dashboard_provider.dart';
import '../../../core/theme/blueprint_theme.dart';

class TransactionsScreen extends ConsumerWidget {
  const TransactionsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final transactionsAsync = ref.watch(transactionsProvider);
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final dateFormat = DateFormat('dd/MM/yyyy');

    return Scaffold(
      appBar: AppBar(
        title: const Text('TRANSAÇÕES'),
        actions: [
          IconButton(
            icon: const Icon(Icons.filter_list_rounded, size: 20),
            onPressed: () {
              // TODO: Implementar filtros
            },
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: transactionsAsync.when(
        data: (transactions) {
          if (transactions.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.inventory_2_outlined, size: 48, color: BlueprintTheme.textSecondary.withOpacity(0.5)),
                  const SizedBox(height: 16),
                  const Text(
                    'BUFFER_EMPTY: SEM_OPERAÇÕES',
                    style: TextStyle(fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, fontSize: 10),
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () async => ref.refresh(transactionsProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: transactions.length,
              separatorBuilder: (context, index) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final tx = transactions[index];
                final isCredit = tx.direction == 'credit';

                return Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.surface,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: BlueprintTheme.border, width: 1),
                  ),
                  child: Row(
                    children: [
                      Container(
                        width: 40,
                        height: 40,
                        decoration: BoxDecoration(
                          color: (isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.danger).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Center(
                          child: Icon(
                            isCredit ? Icons.add_rounded : Icons.remove_rounded,
                            color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.danger,
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              tx.description.toUpperCase(),
                              style: const TextStyle(
                                fontWeight: FontWeight.w900,
                                fontSize: 12,
                                letterSpacing: -0.2,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '${dateFormat.format(tx.date)} | ${tx.accountName.toUpperCase()} | ${tx.categoryName.toUpperCase()}',
                              style: const TextStyle(
                                color: BlueprintTheme.textSecondary,
                                fontSize: 8,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        currencyFormat.format(tx.amount),
                        style: TextStyle(
                          fontWeight: FontWeight.w900,
                          fontSize: 14,
                          fontFamily: 'monospace',
                          color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.textPrimary,
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, stack) => Center(child: Text('ERRO: $err')),
      ),
    );
  }
}
