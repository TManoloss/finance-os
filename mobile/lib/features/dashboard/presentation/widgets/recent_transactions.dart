import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class RecentTransactions extends StatelessWidget {
  final List<dynamic> transactions;
  const RecentTransactions({super.key, required this.transactions});

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final dateFmt = DateFormat('dd/MM');
    final items = transactions.take(8).toList();

    if (items.isEmpty) {
      return Container(
        padding: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: BlueprintTheme.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: BlueprintTheme.border),
        ),
        child: const Center(
          child: Text('NENHUMA_TRANSAÇÃO_RECENTE', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: BlueprintTheme.textSecondary)),
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Text('ÚLTIMAS_OPERAÇÕES', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
          ),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            padding: const EdgeInsets.all(12),
            itemCount: items.length,
            separatorBuilder: (_, __) => Divider(color: BlueprintTheme.border.withValues(alpha: 0.5), height: 1),
            itemBuilder: (context, index) {
              final tx = items[index];
              final isCredit = (tx['direction'] ?? '') == 'credit';
              final amount = (tx['amount'] as num?)?.toDouble() ?? 0;
              final description = (tx['description'] ?? '').toString();
              final date = tx['date'] != null ? DateTime.tryParse(tx['date'].toString()) : null;
              final categoryName = tx['category']?['name'] ?? tx['category_name'] ?? '';
              final categoryColor = tx['category']?['color'] ?? '#7C6FFF';

              return Padding(
                padding: const EdgeInsets.symmetric(vertical: 10),
                child: Row(
                  children: [
                    Container(
                      width: 36, height: 36,
                      decoration: BoxDecoration(
                        color: (isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.danger).withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Icon(
                        isCredit ? Icons.add_rounded : Icons.remove_rounded,
                        color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.danger,
                        size: 18,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            description.toUpperCase(),
                            style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 11, overflow: TextOverflow.ellipsis),
                            maxLines: 1,
                          ),
                          const SizedBox(height: 2),
                          Text(
                            '${date != null ? dateFmt.format(date) : '—'} • ${categoryName.toUpperCase()}',
                            style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      fmt.format(amount),
                      style: TextStyle(
                        fontWeight: FontWeight.w900,
                        fontSize: 13,
                        fontFamily: 'monospace',
                        color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.textPrimary,
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}
