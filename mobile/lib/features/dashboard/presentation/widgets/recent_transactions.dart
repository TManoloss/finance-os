import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class RecentTransactions extends StatelessWidget {
  final List<dynamic> transactions;
  const RecentTransactions({super.key, required this.transactions});

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    if (transactions.isEmpty) {
      return Padding(
        padding: const EdgeInsets.all(32),
        child: Center(
          child: Text('BUFFER_EMPTY: NENHUMA_OPERACAO_RECENTE',
              style: terminalLabel(), textAlign: TextAlign.center),
        ),
      );
    }
    return Column(
      children: transactions.map((tx) {
        final isCredit = (tx['direction'] ?? '') == 'credit';
        final amount = (tx['amount'] as num?)?.toDouble() ?? 0;
        final desc = (tx['description'] ?? '').toString().toUpperCase();
        final date = tx['date']?.toString() ?? '';
        final category = tx['category']?['name']?.toString() ?? 'NULL';
        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: const BoxDecoration(
            border: Border(bottom: BorderSide(color: BlueprintTheme.border, width: 1)),
          ),
          child: Row(children: [
            // ícone + / - quadrado (igual ao web)
            Container(
              width: 40, height: 40,
              color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.danger,
              child: Center(child: Text(isCredit ? '+' : '-',
                style: const TextStyle(color: Colors.white, fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 18))),
            ),
            const SizedBox(width: 14),
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(desc, style: const TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12),
                overflow: TextOverflow.ellipsis),
              const SizedBox(height: 2),
              Text('$date | $category', style: terminalLabel(fontSize: 8)),
            ])),
            const SizedBox(width: 8),
            Text(fmt.format(amount),
              style: moneyStyle(
                color: isCredit ? BlueprintTheme.accentTeal : BlueprintTheme.textPrimary,
                fontSize: 14,
              )),
          ]),
        );
      }).toList(),
    );
  }
}
