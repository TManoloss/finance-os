import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/data/summary_model.dart';

class SummaryCards extends StatelessWidget {
  final FinancialSummary summary;
  const SummaryCards({super.key, required this.summary});

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final balance = summary.checkingBalance - (summary.closedInvoice + summary.monthInstallments);

    final stats = [
      {'label': 'SALDO_EM_CONTA', 'value': summary.checkingBalance, 'color': BlueprintTheme.accentTeal},
      {'label': 'FATURA_FECHADA', 'value': summary.closedInvoice + summary.monthInstallments, 'color': BlueprintTheme.danger},
      {'label': 'FATURA_EM_ABERTO', 'value': summary.currentInvoice, 'color': BlueprintTheme.warning},
      {'label': 'TOTAL_RECEBIDO', 'value': summary.totalReceived, 'color': BlueprintTheme.accentTeal},
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Header com saldo líquido — igual ao web
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          color: BlueprintTheme.elevated,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(Icons.terminal, size: 12, color: BlueprintTheme.textSecondary),
                  const SizedBox(width: 6),
                  Text('FINANCE_CORE_V1.1', style: terminalLabel()),
                ],
              ),
              const SizedBox(height: 4),
              const Text('SALDO_TOTAL_LIQUIDO', style: TextStyle(
                fontFamily: 'monospace', fontSize: 10, fontWeight: FontWeight.w900,
                color: BlueprintTheme.textSecondary, letterSpacing: 1,
              )),
              const SizedBox(height: 4),
              Text(
                fmt.format(balance),
                style: moneyStyle(
                  color: balance >= 0 ? BlueprintTheme.accentTeal : BlueprintTheme.danger,
                  fontSize: 32,
                ),
              ),
            ],
          ),
        ),
        // Grid de stats — 2 colunas com bordas separando (igual ao web)
        GridView.count(
          crossAxisCount: 2,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          childAspectRatio: 2.2,
          children: stats.map((s) => _buildStatCell(
            label: s['label'] as String,
            value: fmt.format(s['value'] as double),
            color: s['color'] as Color,
          )).toList(),
        ),
      ],
    );
  }

  Widget _buildStatCell({required String label, required String value, required Color color}) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: const BoxDecoration(
        color: BlueprintTheme.background,
        border: Border(
          right: BorderSide(color: BlueprintTheme.border, width: 2),
          bottom: BorderSide(color: BlueprintTheme.border, width: 2),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Row(children: [
            Container(width: 4, height: 4, color: BlueprintTheme.textSecondary),
            const SizedBox(width: 4),
            Expanded(child: Text(label, style: terminalLabel(fontSize: 8), overflow: TextOverflow.ellipsis)),
          ]),
          const SizedBox(height: 4),
          Text(value, style: moneyStyle(color: color, fontSize: 15), overflow: TextOverflow.ellipsis),
        ],
      ),
    );
  }
}
