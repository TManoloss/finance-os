import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../shared/widgets/blueprint_card.dart';
import 'dashboard_provider.dart';
import 'widgets/category_chart.dart';

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
          'DASHBOARD_V1.0',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            letterSpacing: 2,
          ),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: false,
      ),
      body: summaryAsync.when(
        data: (summary) {
          final balance = summary.totalReceived - summary.totalSpent;

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                BlueprintCard(
                  label: 'TELEMETRIA_GASTOS',
                  height: 200,
                  child: CategoryChart(categories: summary.byCategory),
                ),
                const SizedBox(height: 24),
                GridView.count(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  crossAxisCount: 2,
                  crossAxisSpacing: 16,
                  mainAxisSpacing: 16,
                  childAspectRatio: 1.2,
                  children: [
                    BlueprintCard(
                      label: 'TOTAL_RECEBIDO',
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            currencyFormat.format(summary.totalReceived),
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: Colors.green,
                            ),
                          ),
                          const Text(
                            'REF: ATUAL',
                            style: TextStyle(fontSize: 10),
                          ),
                        ],
                      ),
                    ),
                    BlueprintCard(
                      label: 'TOTAL_GASTO',
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            currencyFormat.format(summary.totalSpent),
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: Colors.red,
                            ),
                          ),
                          const Text(
                            'REF: ATUAL',
                            style: TextStyle(fontSize: 10),
                          ),
                        ],
                      ),
                    ),
                    BlueprintCard(
                      label: 'SALDO_DISPONIVEL',
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            currencyFormat.format(balance),
                            style: TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: balance >= 0 ? Colors.blue : Colors.orange,
                            ),
                          ),
                          const Text(
                            'CALC_SYSTEM',
                            style: TextStyle(fontSize: 10),
                          ),
                        ],
                      ),
                    ),
                    BlueprintCard(
                      label: 'CATEGORIAS',
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            '${summary.byCategory.length}',
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const Text(
                            'ATIVAS',
                            style: TextStyle(fontSize: 10),
                          ),
                        ],
                      ),
                    ),
                  ],
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
