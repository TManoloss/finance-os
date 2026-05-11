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
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      backgroundColor: const Color(0xFFD4D1CA),
      appBar: AppBar(
        title: const Text(
          'DASHBOARD_V2.0',
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
      body: summaryAsync.when(
        data: (summary) {
          return SingleChildScrollView(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // 1. Estado Líquido
                BlueprintCard(
                  label: 'ESTADO_LIQUIDO_ATUAL',
                  child: Text(
                    currencyFormat.format(summary.checkingBalance - summary.creditBalance),
                    style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                ),
                const SizedBox(height: 16),
                // 2. Detalhamento Ativos vs Passivos
                Row(
                  children: [
                    Expanded(
                      child: BlueprintCard(
                        label: 'EM_CONTA',
                        child: Text(
                          currencyFormat.format(summary.checkingBalance),
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: BlueprintCard(
                        label: 'EM_CARTÕES',
                        child: Text(
                          currencyFormat.format(summary.creditBalance),
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                BlueprintCard(
                  child: CommitmentBar(
                    checking: summary.checkingBalance,
                    credit: summary.creditBalance,
                  ),
                ),
                const SizedBox(height: 24),
                // 3. Telemetria Diária
                BlueprintCard(
                  label: 'TELEMETRIA_DIARIA',
                  height: 150,
                  child: DailySpendingChart(byDay: summary.byDay),
                ),
                const SizedBox(height: 24),
                // 4. Ranking
                BlueprintCard(
                  label: 'RANKING_MERCANTES',
                  child: MerchantRankingWidget(merchants: summary.topMerchants),
                ),
                const SizedBox(height: 24),
                // 5. Categorias (Original)
                BlueprintCard(
                  label: 'TELEMETRIA_GASTOS',
                  height: 200,
                  child: CategoryChart(categories: summary.byCategory),
                ),
              ],
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
                ),
                const SizedBox(height: 16),
                ElevatedButton(
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
