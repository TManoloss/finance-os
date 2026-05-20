import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
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
      appBar: AppBar(title: const Text('CARTÕES')),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(installmentsProvider);
          ref.invalidate(subscriptionsProvider);
          ref.invalidate(connectedAccountsProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // --- Cartões de Crédito ---
              const Text('CARTÕES_CONECTADOS', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
              const SizedBox(height: 12),
              accountsAsync.when(
                data: (accounts) {
                  final cards = accounts.where((a) => (a['account_type'] ?? '').toString().toUpperCase().contains('CREDIT')).toList();
                  if (cards.isEmpty) {
                    return _buildEmptyState('NENHUM_CARTÃO_DETECTADO', Icons.credit_card_off_rounded);
                  }
                  return Column(
                    children: cards.map((card) => _buildCreditCard(card, fmt)).toList(),
                  );
                },
                loading: () => const Center(child: Padding(padding: EdgeInsets.all(32), child: CircularProgressIndicator())),
                error: (e, _) => Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontSize: 10)),
              ),

              const SizedBox(height: 32),

              // --- Parcelamentos Ativos ---
              const Text('PARCELAMENTOS_ATIVOS', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
              const SizedBox(height: 12),
              installmentsAsync.when(
                data: (installments) {
                  if (installments.isEmpty) {
                    return _buildEmptyState('NENHUM_PARCELAMENTO_ATIVO', Icons.check_circle_outline_rounded);
                  }
                  return Column(
                    children: installments.map((inst) => _buildInstallmentCard(inst, fmt)).toList(),
                  );
                },
                loading: () => const Center(child: Padding(padding: EdgeInsets.all(24), child: CircularProgressIndicator())),
                error: (e, _) => Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontSize: 10)),
              ),

              const SizedBox(height: 32),

              // --- Assinaturas Detectadas ---
              const Text('ASSINATURAS_DETECTADAS', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.warning)),
              const SizedBox(height: 12),
              subscriptionsAsync.when(
                data: (subs) {
                  if (subs.isEmpty) {
                    return _buildEmptyState('NENHUMA_ASSINATURA_DETECTADA', Icons.subscriptions_outlined);
                  }
                  return Column(
                    children: subs.map((sub) => _buildSubscriptionCard(sub, fmt)).toList(),
                  );
                },
                loading: () => const Center(child: Padding(padding: EdgeInsets.all(24), child: CircularProgressIndicator())),
                error: (e, _) => Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontSize: 10)),
              ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCreditCard(Map<String, dynamic> card, NumberFormat fmt) {
    final balance = (card['balance'] as num?)?.toDouble() ?? 0;
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF1A1A2E), Color(0xFF111118)],
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: BlueprintTheme.accentPurple.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              if (card['institution_logo'] != null)
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: Image.network(card['institution_logo'], width: 32, height: 32, fit: BoxFit.contain),
                ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  (card['institution_name'] ?? 'CARTÃO').toString().toUpperCase(),
                  style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 14),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const Icon(Icons.credit_card_rounded, size: 20, color: BlueprintTheme.textSecondary),
            ],
          ),
          const SizedBox(height: 16),
          const Text('SALDO_DEVEDOR', style: TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 1)),
          Text(
            fmt.format(balance.abs()),
            style: TextStyle(fontSize: 28, fontWeight: FontWeight.w900, fontFamily: 'monospace', color: balance < 0 ? BlueprintTheme.danger : BlueprintTheme.textPrimary),
          ),
        ],
      ),
    );
  }

  Widget _buildInstallmentCard(Map<String, dynamic> inst, NumberFormat fmt) {
    final currentInst = (inst['installment_current'] as num?)?.toInt() ?? 0;
    final totalInst = (inst['installments_total'] as num?)?.toInt() ?? 1;
    final merchant = (inst['merchant_name'] ?? 'DESCONHECIDO').toString();
    final amount = (inst['total_amount'] as num?)?.toDouble() ?? 0;
    final installmentValue = totalInst > 0 ? amount / totalInst : 0.0;
    final progress = totalInst > 0 ? currentInst / totalInst : 0.0;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
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
              Expanded(
                child: Text(merchant.toUpperCase(), style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12)),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: BlueprintTheme.accentPurple.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  '$currentInst/$totalInst',
                  style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w900, color: BlueprintTheme.accentPurple, fontFamily: 'monospace'),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Parcela: ${fmt.format(installmentValue)}', style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary)),
              Text('Total: ${fmt.format(amount)}', style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, fontFamily: 'monospace')),
            ],
          ),
          const SizedBox(height: 10),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: progress.clamp(0.0, 1.0),
              minHeight: 6,
              backgroundColor: BlueprintTheme.elevated,
              valueColor: const AlwaysStoppedAnimation(BlueprintTheme.accentPurple),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSubscriptionCard(Map<String, dynamic> sub, NumberFormat fmt) {
    final merchant = (sub['merchant'] ?? sub['merchant_name'] ?? '?').toString();
    final monthlyValue = (sub['monthly_value'] as num?)?.toDouble() ?? (sub['total'] as num?)?.toDouble() ?? 0;
    final status = (sub['status'] ?? 'active').toString();
    final isIrregular = status == 'irregular';

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: isIrregular ? BlueprintTheme.warning.withValues(alpha: 0.5) : BlueprintTheme.border),
      ),
      child: Row(
        children: [
          Container(
            width: 40, height: 40,
            decoration: BoxDecoration(
              color: (isIrregular ? BlueprintTheme.warning : BlueprintTheme.accentTeal).withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              Icons.autorenew_rounded,
              color: isIrregular ? BlueprintTheme.warning : BlueprintTheme.accentTeal,
              size: 20,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(merchant.toUpperCase(), style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12)),
                const SizedBox(height: 2),
                Text(
                  isIrregular ? 'STATUS: IRREGULAR' : 'STATUS: ATIVA',
                  style: TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: isIrregular ? BlueprintTheme.warning : BlueprintTheme.accentTeal),
                ),
              ],
            ),
          ),
          Text(
            '${fmt.format(monthlyValue)}/mês',
            style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 13, fontFamily: 'monospace'),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState(String label, IconData icon) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Column(
        children: [
          Icon(icon, size: 40, color: BlueprintTheme.textSecondary.withValues(alpha: 0.5)),
          const SizedBox(height: 12),
          Text(label, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: BlueprintTheme.textSecondary)),
        ],
      ),
    );
  }
}
