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
    final netBalance = summary.checkingBalance + summary.creditBalance;

    return Column(
      children: [
        // Card principal: Saldo líquido
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [Color(0xFF1A1A2E), Color(0xFF0A0A0F)],
            ),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: BlueprintTheme.accentPurple.withValues(alpha: 0.3)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'SALDO_LÍQUIDO',
                style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, color: BlueprintTheme.textSecondary, letterSpacing: 1),
              ),
              const SizedBox(height: 8),
              Text(
                fmt.format(netBalance),
                style: TextStyle(
                  fontSize: 32,
                  fontWeight: FontWeight.w900,
                  fontFamily: 'monospace',
                  letterSpacing: -1,
                  color: netBalance >= 0 ? BlueprintTheme.accentTeal : BlueprintTheme.danger,
                ),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  _buildMiniStat('CONTA', fmt.format(summary.checkingBalance), BlueprintTheme.textPrimary),
                  const SizedBox(width: 24),
                  _buildMiniStat('CRÉDITO', fmt.format(summary.creditBalance), BlueprintTheme.warning),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        // Row: Gastos / Receitas / Parcelas
        Row(
          children: [
            Expanded(
              child: _buildMetricCard(
                label: 'GASTO_MÊS',
                value: fmt.format(summary.totalSpent),
                color: BlueprintTheme.danger,
                icon: Icons.arrow_downward_rounded,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildMetricCard(
                label: 'RECEITA_MÊS',
                value: fmt.format(summary.totalReceived),
                color: BlueprintTheme.accentTeal,
                icon: Icons.arrow_upward_rounded,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _buildMetricCard(
                label: 'FATURA_FECHADA',
                value: fmt.format(summary.closedInvoice),
                color: BlueprintTheme.warning,
                icon: Icons.receipt_long_rounded,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _buildMetricCard(
                label: 'PARCELAS_MÊS',
                value: fmt.format(summary.monthInstallments),
                color: BlueprintTheme.accentPurple,
                icon: Icons.calendar_today_rounded,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildMiniStat(String label, String value, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 1)),
        Text(value, style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, fontFamily: 'monospace', color: color)),
      ],
    );
  }

  Widget _buildMetricCard({required String label, required String value, required Color color, required IconData icon}) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, color: color, size: 14),
              ),
              const Spacer(),
            ],
          ),
          const SizedBox(height: 12),
          Text(label, style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 0.5)),
          const SizedBox(height: 4),
          Text(value, style: TextStyle(fontSize: 15, fontWeight: FontWeight.w900, fontFamily: 'monospace', color: color)),
        ],
      ),
    );
  }
}
