import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';

class CardsScreen extends ConsumerWidget {
  const CardsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final installmentsAsync = ref.watch(installmentsProvider);
    final subscriptionsAsync = ref.watch(subscriptionsProvider);
    final accountsAsync = ref.watch(connectedAccountsProvider);
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      appBar: AppBar(title: const Text('GESTAO_DE_CREDITO')),
      body: RefreshIndicator(
        color: BlueprintTheme.accentPurple,
        onRefresh: () async {
          ref.invalidate(installmentsProvider);
          ref.invalidate(subscriptionsProvider);
          ref.invalidate(connectedAccountsProvider);
        },
        child: ListView(
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            // ── Header ─────────────────────────────────────────────────────
            Container(
              width: double.infinity,
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
              color: BlueprintTheme.elevated,
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Row(children: [
                  const Icon(LucideIcons.terminal, size: 12),
                  const SizedBox(width: 6),
                  Text('CREDIT_MONITOR_V1.0', style: terminalLabel()),
                ]),
                const SizedBox(height: 4),
                accountsAsync.when(
                  data: (accs) {
                    final cnt = accs.where((a) => (a['account_type'] ?? '').toString().toUpperCase().contains('CREDIT')).length;
                    return Text('CONTAS_MONITORADAS: $cnt', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10));
                  },
                  loading: () => Text('CONTAS_MONITORADAS: ...', style: terminalLabel()),
                  error: (_, __) => const SizedBox.shrink(),
                ),
              ]),
            ),
            Container(height: 2, color: BlueprintTheme.border),

            // ── Cartões ────────────────────────────────────────────────────
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
              child: Row(children: [
                Container(width: 8, height: 8, color: BlueprintTheme.textPrimary),
                const SizedBox(width: 8),
                Text('CARTOES_ATIVOS', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10)),
              ]),
            ),
            accountsAsync.when(
              data: (accounts) {
                final cards = accounts.where((a) => (a['account_type'] ?? '').toString().toUpperCase().contains('CREDIT')).toList();
                if (cards.isEmpty) return _emptyState('NO_CREDIT_LINES_DETECTED', LucideIcons.creditCard);
                return Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Column(children: cards.map((card) => _buildCreditCard(card, fmt)).toList()),
                );
              },
              loading: () => const Padding(padding: EdgeInsets.all(32), child: Center(child: CircularProgressIndicator())),
              error: (e, _) => Padding(padding: const EdgeInsets.all(16), child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace', fontSize: 10))),
            ),
            Container(height: 2, color: BlueprintTheme.border, margin: const EdgeInsets.only(top: 16)),

            // ── Parcelamentos ──────────────────────────────────────────────
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
              child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                Row(children: [
                  const Icon(LucideIcons.clock, size: 14),
                  const SizedBox(width: 8),
                  Text('COMPRAS_PARCELADAS_EM_ABERTO', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10)),
                ]),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  color: BlueprintTheme.danger,
                  child: const Text('ALERTA_DE_FLUXO', style: TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 8, color: Colors.white)),
                ),
              ]),
            ),
            installmentsAsync.when(
              data: (insts) {
                if (insts.isEmpty) return _emptyState('NO_INSTALLMENTS_DETECTED', LucideIcons.shieldCheck);
                return Column(children: insts.map<Widget>((ins) => _buildInstallmentRow(ins, fmt)).toList());
              },
              loading: () => const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
              error: (e, _) => Padding(padding: const EdgeInsets.all(16), child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace', fontSize: 10))),
            ),
            Container(height: 2, color: BlueprintTheme.border, margin: const EdgeInsets.only(top: 8)),

            // ── Assinaturas ────────────────────────────────────────────────
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
              child: Text('ASSINATURAS_DETECTADAS', style: terminalLabel(color: BlueprintTheme.warning, fontSize: 10)),
            ),
            subscriptionsAsync.when(
              data: (subs) {
                if (subs.isEmpty) return _emptyState('NENHUMA_ASSINATURA_DETECTADA', LucideIcons.refreshCcw);
                return Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Column(children: subs.map<Widget>((sub) => _buildSubscriptionRow(sub, fmt)).toList()),
                );
              },
              loading: () => const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
              error: (e, _) => const SizedBox.shrink(),
            ),
            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  Widget _buildCreditCard(Map<String, dynamic> card, NumberFormat fmt) {
    final balance = (card['balance'] as num?)?.toDouble() ?? 0;
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: BlueprintTheme.elevated,
        border: Border.all(color: BlueprintTheme.border, width: 2),
        boxShadow: const [BoxShadow(color: BlueprintTheme.border, offset: Offset(4, 4))],
      ),
      padding: const EdgeInsets.all(20),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Expanded(child: Text(
            (card['institution_name'] ?? 'CARTÃO').toString().toUpperCase(),
            style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 12),
          )),
          const Icon(LucideIcons.creditCard, size: 18),
        ]),
        const SizedBox(height: 12),
        Text('**** **** **** 4521', style: moneyStyle(fontSize: 16)),
        const SizedBox(height: 12),
        Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
          Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('SALDO_DEVEDOR', style: terminalLabel(fontSize: 8)),
            Text(fmt.format(balance.abs()), style: moneyStyle(color: balance < 0 ? BlueprintTheme.danger : BlueprintTheme.textPrimary, fontSize: 20)),
          ]),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            color: BlueprintTheme.textPrimary,
            child: const Text('ACTIVE', style: TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 9, color: BlueprintTheme.surface)),
          ),
        ]),
      ]),
    );
  }

  Widget _buildInstallmentRow(Map<String, dynamic> ins, NumberFormat fmt) {
    final current = (ins['installment_current'] as num?)?.toInt() ?? 0;
    final total = (ins['installments_total'] as num?)?.toInt() ?? 1;
    final merchant = (ins['merchant_name'] ?? '').toString().toUpperCase();
    final amount = (ins['total_amount'] as num?)?.toDouble() ?? 0;
    final instValue = total > 0 ? amount / total : 0.0;
    final progress = total > 0 ? current / total : 0.0;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: BlueprintTheme.border, width: 1)),
        color: BlueprintTheme.surface,
      ),
      child: Row(children: [
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(merchant, style: const TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12)),
          const SizedBox(height: 4),
          Row(children: [
            Text('$current/$total', style: terminalLabel(fontSize: 9)),
            const SizedBox(width: 8),
            Expanded(child: Container(
              height: 6,
              decoration: BoxDecoration(
                border: Border.all(color: BlueprintTheme.border, width: 1),
                color: BlueprintTheme.background,
              ),
              child: FractionallySizedBox(
                widthFactor: progress.clamp(0.0, 1.0),
                alignment: Alignment.centerLeft,
                child: Container(color: BlueprintTheme.accentPurple),
              ),
            )),
          ]),
        ])),
        const SizedBox(width: 12),
        Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
          Text(fmt.format(instValue), style: moneyStyle(fontSize: 14)),
          Text('TOTAL: ${fmt.format(amount)}', style: terminalLabel(color: BlueprintTheme.danger, fontSize: 8)),
        ]),
      ]),
    );
  }

  Widget _buildSubscriptionRow(Map<String, dynamic> sub, NumberFormat fmt) {
    final merchant = (sub['merchant'] ?? sub['merchant_name'] ?? '?').toString().toUpperCase();
    final monthlyValue = (sub['monthly_value'] as num?)?.toDouble() ?? 0;
    final isIrregular = (sub['status'] ?? '') == 'irregular';
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        border: Border.all(color: isIrregular ? BlueprintTheme.warning : BlueprintTheme.border, width: 2),
      ),
      child: Row(children: [
        Container(
          width: 36, height: 36,
          color: (isIrregular ? BlueprintTheme.warning : BlueprintTheme.accentTeal).withValues(alpha: 0.15),
          child: Icon(LucideIcons.refreshCcw, color: isIrregular ? BlueprintTheme.warning : BlueprintTheme.accentTeal, size: 16),
        ),
        const SizedBox(width: 12),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(merchant, style: const TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12)),
          Text(isIrregular ? 'STATUS: IRREGULAR' : 'STATUS: ATIVA',
            style: terminalLabel(color: isIrregular ? BlueprintTheme.warning : BlueprintTheme.accentTeal, fontSize: 8)),
        ])),
        Text('${fmt.format(monthlyValue)}/mês', style: moneyStyle(fontSize: 12)),
      ]),
    );
  }

  Widget _emptyState(String label, IconData icon) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        color: BlueprintTheme.elevated,
        border: Border.all(color: BlueprintTheme.border, width: 1),
      ),
      child: Column(children: [
        Icon(icon, size: 32, color: BlueprintTheme.textSecondary),
        const SizedBox(height: 8),
        Text(label, style: terminalLabel(), textAlign: TextAlign.center),
      ]),
    );
  }
}
