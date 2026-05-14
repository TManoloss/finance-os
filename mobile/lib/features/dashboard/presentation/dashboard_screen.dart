import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../shared/widgets/blueprint_card.dart';
import 'dashboard_provider.dart';
import 'widgets/category_chart.dart';
import 'widgets/daily_spending_chart.dart';
import 'widgets/merchant_ranking_widget.dart';
import 'widgets/commitment_bar.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summaryAsync = ref.watch(summaryProvider);
    final currentPeriod = ref.watch(periodProvider);
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    final periods = [
      {'id': 'month', 'label': 'MÊS'},
      {'id': 'quarter', 'label': 'TRIM'},
      {'id': 'semester', 'label': 'SEM'},
      {'id': 'year', 'label': 'ANO'},
      {'id': 'all', 'label': 'TUDO'},
    ];

    return Scaffold(
      backgroundColor: const Color(0xFFD4D1CA),
      appBar: AppBar(
        title: const Text(
          'DASHBOARD_V2.2',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            letterSpacing: 2,
          ),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: false,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.refresh(summaryProvider),
          ),
        ],
      ),
      body: Column(
        children: [
          // Period Selector
          Container(
            height: 50,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: periods.length,
              separatorBuilder: (context, index) => const SizedBox(width: 8),
              itemBuilder: (context, index) {
                final p = periods[index];
                final isSelected = currentPeriod == p['id'];
                return ChoiceChip(
                  label: Text(
                    p['label']!,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 10,
                      color: isSelected ? Colors.white : Colors.black,
                    ),
                  ),
                  selected: isSelected,
                  onSelected: (selected) {
                    if (selected) {
                      ref.read(periodProvider.notifier).state = p['id']!;
                    }
                  },
                  selectedColor: Colors.black,
                  backgroundColor: Colors.white,
                  shape: const RoundedRectangleBorder(
                    side: BorderSide(color: Colors.black, width: 2),
                  ),
                );
              },
            ),
          ),
          Expanded(
            child: summaryAsync.when(
              data: (summary) {
                final consolidatedBalance = summary.checkingBalance - summary.creditBalance;

                return RefreshIndicator(
                  onRefresh: () async => ref.refresh(summaryProvider),
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(16.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        // 1. Saldo Consolidado (Líquido)
                        BlueprintCard(
                          label: 'SALDO_LÍQUIDO_CONSOLIDADO',
                          child: Text(
                            currencyFormat.format(consolidatedBalance),
                            style: TextStyle(
                              fontSize: 28, 
                              fontWeight: FontWeight.w900,
                              color: consolidatedBalance >= 0 ? Colors.black : Colors.red,
                              fontFamily: 'monospace',
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        // 2. Detalhamento Ativos vs Passivos
                        Row(
                          children: [
                            Expanded(
                              child: BlueprintCard(
                                label: 'CONTA_CORRENTE',
                                child: Text(
                                  currencyFormat.format(summary.checkingBalance),
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold, 
                                    color: Color(0xFF4ECDC4),
                                    fontSize: 16,
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: BlueprintCard(
                                label: 'CARTÃO_CRÉDITO',
                                child: Text(
                                  currencyFormat.format(summary.creditBalance),
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold, 
                                    color: Color(0xFFFF6B6B),
                                    fontSize: 16,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        // Totais do Período
                        Row(
                          children: [
                            Expanded(
                              child: BlueprintCard(
                                label: 'RECEBIDO_PERIODO',
                                child: Text(
                                  currencyFormat.format(summary.totalReceived),
                                  style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.blueGrey),
                                ),
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: BlueprintCard(
                                label: 'GASTO_PERIODO',
                                child: Text(
                                  currencyFormat.format(summary.totalSpent),
                                  style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.redAccent),
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 24),
                        // 3. Telemetria Diária
                        BlueprintCard(
                          label: 'FLUXO_DE_CAIXA_DETALHADO',
                          height: 220,
                          child: Padding(
                            padding: const EdgeInsets.only(top: 20, right: 10, bottom: 10),
                            child: DailySpendingChart(byDay: summary.byDay),
                          ),
                        ),
                        const SizedBox(height: 24),
                        // 4. Ranking
                        BlueprintCard(
                          label: 'PRINCIPAIS_DESTINOS_GASTOS',
                          child: MerchantRankingWidget(merchants: summary.topMerchants),
                        ),
                        const SizedBox(height: 24),
                        // 5. Categorias
                        BlueprintCard(
                          label: 'DISTRIBUICAO_CATEGORIAS',
                          height: 250,
                          child: CategoryChart(categories: summary.byCategory),
                        ),
                        const SizedBox(height: 24),
                        // 6. Investimentos Preview
                        BlueprintCard(
                          label: 'MÓDULO_INVESTIMENTOS',
                          child: Container(
                            padding: const EdgeInsets.symmetric(vertical: 32),
                            child: Column(
                              children: [
                                const Icon(Icons.trending_up, size: 48, color: Color(0xFF7C6FFF)),
                                const SizedBox(height: 12),
                                const Text(
                                  'EM_DESENVOLVIMENTO',
                                  style: TextStyle(fontWeight: FontWeight.w900, fontSize: 14, letterSpacing: 1),
                                ),
                                const SizedBox(height: 8),
                                const Text(
                                  'Acompanhe o crescimento do seu patrimonio\ne performance de ativos em tempo real.',
                                  textAlign: TextAlign.center,
                                  style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.blueGrey),
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
                child: CircularProgressIndicator(color: Colors.black),
              ),
              error: (err, stack) => Center(
                child: Padding(
                  padding: const EdgeInsets.all(32.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.error_outline, size: 48, color: Colors.red),
                      const SizedBox(height: 16),
                      Text(
                        'ERRO_SISTEMA: $err',
                        textAlign: TextAlign.center,
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.black,
                          foregroundColor: Colors.white,
                          shape: const RoundedRectangleBorder(),
                        ),
                        onPressed: () => ref.refresh(summaryProvider),
                        child: const Text('RETRY_CONNECTION'),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
