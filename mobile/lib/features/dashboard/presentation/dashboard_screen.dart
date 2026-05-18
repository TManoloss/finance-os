import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../shared/widgets/blueprint_card.dart';
import '../../theme/blueprint_theme.dart';
import 'dashboard_provider.dart';
import 'widgets/category_chart.dart';
import 'widgets/daily_spending_chart.dart';
import 'widgets/merchant_ranking_widget.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summaryAsync = ref.watch(summaryProvider);
    final currentPeriod = ref.watch(periodProvider);
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    final periods = [
      {'id': 'month', 'label': 'MÊS'},
      {'id': 'quarter', 'label': 'TRIMESTRE'},
      {'id': 'semester', 'label': 'SEMESTRE'},
      {'id': 'year', 'label': 'ANO'},
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('DASHBOARD'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh, size: 20),
            onPressed: () => ref.refresh(summaryProvider),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: summaryAsync.when(
        data: (summary) {
          final checkingBalance = summary.checkingBalance;
          final closedInvoice = summary.closedInvoice;
          final monthInstallments = summary.monthInstallments;
          final balance = checkingBalance - (closedInvoice + monthInstallments);

          return RefreshIndicator(
            onRefresh: () async => ref.refresh(summaryProvider),
            child: SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.symmetric(horizontal: 16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Greeting & Date
                  const SizedBox(height: 8),
                  Text(
                    'BOM DIA, OPERADOR',
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                      color: BlueprintTheme.textSecondary,
                      letterSpacing: 1.2,
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Main Balance Card
                  BlueprintCard(
                    label: 'SALDO_TOTAL_LIQUIDO',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          currencyFormat.format(balance),
                          style: TextStyle(
                            fontSize: 32,
                            fontWeight: FontWeight.w900,
                            color: balance >= 0 
                                ? BlueprintTheme.accentTeal 
                                : BlueprintTheme.danger,
                            fontFamily: 'monospace',
                            letterSpacing: -1,
                          ),
                        ),
                        const SizedBox(height: 16),
                        // INFLAÇÃO PESSOAL PREVIEW
                        ref.watch(inflationProvider).when(
                          data: (data) {
                            if (data.isEmpty) return const SizedBox.shrink();
                            final rate = data['personal_inflation_rate'] ?? 0.0;
                            return Padding(
                              padding: const EdgeInsets.bottom(16.0),
                              child: BlueprintCard(
                                label: 'INFLAÇÃO_PESSOAL_MENSAL',
                                child: Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  children: [
                                    Text(
                                      '${rate.toStringAsFixed(2)}%',
                                      style: TextStyle(
                                        fontSize: 24,
                                        fontWeight: FontWeight.w900,
                                        color: rate > 5 ? Colors.red : Colors.black,
                                      ),
                                    ),
                                    const Icon(Icons.trending_up, color: Colors.black),
                                  ],
                                ),
                              ),
                            );
                          },
                          loading: () => const SizedBox.shrink(),
                          error: (_, __) => const SizedBox.shrink(),
                        ),
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                              decoration: BoxDecoration(
                                color: BlueprintTheme.accentTeal.withOpacity(0.1),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Text(
                                '+2.4% ESTE MÊS',
                                style: TextStyle(
                                  fontSize: 8,
                                  fontWeight: FontWeight.bold,
                                  color: BlueprintTheme.accentTeal,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Quick Stats Grid
                  Row(
                    children: [
                      Expanded(
                        child: BlueprintCard(
                          label: 'GASTO_MES',
                          padding: const EdgeInsets.all(12),
                          child: Text(
                            currencyFormat.format(summary.totalSpent),
                            style: const TextStyle(
                              fontWeight: FontWeight.w900,
                              color: BlueprintTheme.danger,
                              fontSize: 14,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: BlueprintCard(
                          label: 'RECEBIDO',
                          padding: const EdgeInsets.all(12),
                          child: Text(
                            currencyFormat.format(summary.totalReceived),
                            style: const TextStyle(
                              fontWeight: FontWeight.w900,
                              color: BlueprintTheme.accentTeal,
                              fontSize: 14,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),

                  // Period Selector (Tabs style)
                  SizedBox(
                    height: 32,
                    child: ListView.separated(
                      scrollDirection: Axis.horizontal,
                      itemCount: periods.length,
                      separatorBuilder: (context, index) => const SizedBox(width: 8),
                      itemBuilder: (context, index) {
                        final p = periods[index];
                        final isSelected = currentPeriod == p['id'];
                        return GestureDetector(
                          onTap: () => ref.read(periodProvider.notifier).state = p['id']!,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 16),
                            alignment: Alignment.center,
                            decoration: BoxDecoration(
                              color: isSelected ? BlueprintTheme.accentPurple : BlueprintTheme.elevated,
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(
                                color: isSelected ? BlueprintTheme.accentPurple : BlueprintTheme.border,
                                width: 1,
                              ),
                            ),
                            child: Text(
                              p['label']!,
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                fontSize: 10,
                                color: isSelected ? Colors.white : BlueprintTheme.textSecondary,
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Cashflow Chart
                  BlueprintCard(
                    label: 'FLUXO_DE_CAIXA_REAL_TIME',
                    height: 240,
                    child: DailySpendingChart(byDay: summary.byDay),
                  ),
                  const SizedBox(height: 24),

                  // Category Donut
                  BlueprintCard(
                    label: 'GASTOS_POR_CATEGORIA',
                    height: 240,
                    child: CategoryChart(categories: summary.byCategory),
                  ),
                  const SizedBox(height: 24),

                  // Top Merchants
                  BlueprintCard(
                    label: 'PRINCIPAIS_DESTINOS',
                    child: MerchantRankingWidget(merchants: summary.topMerchants),
                  ),
                  const SizedBox(height: 24),

                  // Future Projection Link / Preview
                  BlueprintCard(
                    label: 'CENTRO_DE_INVESTIMENTOS',
                    child: Container(
                      padding: const EdgeInsets.symmetric(vertical: 24),
                      child: Column(
                        children: [
                          Icon(
                            Icons.trending_up, 
                            size: 32, 
                            color: BlueprintTheme.accentPurple.withOpacity(0.5)
                          ),
                          const SizedBox(height: 12),
                          const Text(
                            'MODULO_EM_DESENVOLVIMENTO',
                            style: TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                          ),
                          const SizedBox(height: 4),
                          const Text(
                            'Acompanhe o crescimento do seu patrimonio.',
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 10, 
                              fontWeight: FontWeight.bold, 
                              color: BlueprintTheme.textSecondary
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 40),
                ],
              ),
            ),
          );
        },
        loading: () => const Center(
          child: CircularProgressIndicator(color: BlueprintTheme.accentPurple),
        ),
        error: (err, stack) => Center(
          child: Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48, color: BlueprintTheme.danger),
                const SizedBox(height: 16),
                Text(
                  'ERRO_SISTEMA: $err',
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 24),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: BlueprintTheme.accentPurple,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  onPressed: () => ref.refresh(summaryProvider),
                  child: const Text('RETRY_CONNECTION'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
