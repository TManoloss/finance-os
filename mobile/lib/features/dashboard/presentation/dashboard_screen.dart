import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/data/dashboard_models.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/summary_cards.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/category_chart.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/daily_spending_chart.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/recent_transactions.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/merchant_ranking_widget.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/commitment_bar.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/activity_feed.dart';

/// Provider para transações recentes (últimas 8) no dashboard
final recentTxProvider = FutureProvider<List<dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/transactions?page=1&page_size=8');
    return resp.data['data']['transactions'] ?? [];
  } catch (e) {
    return [];
  }
});

/// Provider para o feed de atividades
final activityFeedProvider = FutureProvider<List<dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/feed');
    final data = resp.data['data'];
    if (data is List) return data;
    return [];
  } catch (e) {
    return [];
  }
});

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summaryAsync = ref.watch(summaryProvider);
    final recentTxAsync = ref.watch(recentTxProvider);
    final feedAsync = ref.watch(activityFeedProvider);
    final period = ref.watch(periodProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('FINANCE_OS'),
        actions: [
          IconButton(
            icon: const Icon(Icons.health_and_safety_rounded, size: 20),
            onPressed: () => context.push('/health'),
          ),
          IconButton(
            icon: const Icon(Icons.settings_rounded, size: 22),
            onPressed: () => context.push('/settings'),
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          // ignore: unused_result
          ref.refresh(summaryProvider);
          // ignore: unused_result
          ref.refresh(recentTxProvider);
          // ignore: unused_result
          ref.refresh(activityFeedProvider);
        },
        child: summaryAsync.when(
          data: (summary) {
            // Converter dados dinâmicos para tipos fortes
            final categories = summary.byCategory
                .whereType<Map<String, dynamic>>()
                .map((e) => CategorySummary.fromJson(e))
                .toList();
            final dailyData = summary.byDay
                .whereType<Map<String, dynamic>>()
                .map((e) => DailyBalance.fromJson(e))
                .toList();
            final merchants = summary.topMerchants
                .whereType<Map<String, dynamic>>()
                .map((e) => MerchantSummary.fromJson(e))
                .toList();

            return SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Filtro de período
                  SizedBox(
                    height: 32,
                    child: ListView(
                      scrollDirection: Axis.horizontal,
                      children: [
                        _buildPeriodChip(ref, 'month', 'MÊS', period),
                        const SizedBox(width: 8),
                        _buildPeriodChip(ref, 'quarter', 'TRIMESTRE', period),
                        const SizedBox(width: 8),
                        _buildPeriodChip(ref, 'semester', 'SEMESTRE', period),
                        const SizedBox(width: 8),
                        _buildPeriodChip(ref, 'year', 'ANO', period),
                      ],
                    ),
                  ),
                  const SizedBox(height: 20),

                  // Summary Cards
                  SummaryCards(summary: summary),
                  const SizedBox(height: 24),

                  // Commitment Bar — barra de comprometimento
                  CommitmentBar(
                    spent: summary.totalSpent,
                    limit: summary.totalReceived > 0 ? summary.totalReceived : 1,
                    label: 'COMPROMETIMENTO_RENDA',
                  ),
                  const SizedBox(height: 24),

                  // Gráfico de categorias
                  if (categories.isNotEmpty) ...[
                    Text('DISTRIBUIÇÃO_POR_CATEGORIA',
                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
                    const SizedBox(height: 12),
                    SizedBox(height: 200, child: CategoryChart(categories: categories)),
                    const SizedBox(height: 24),
                  ],

                  // Gráfico de gastos diários
                  if (dailyData.isNotEmpty) ...[
                    Text('GASTOS_DIÁRIOS',
                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
                    const SizedBox(height: 12),
                    SizedBox(height: 200, child: DailySpendingChart(byDay: dailyData)),
                    const SizedBox(height: 24),
                  ],

                  // Top Merchants
                  if (merchants.isNotEmpty) ...[
                    Text('TOP_MERCHANTS',
                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
                    const SizedBox(height: 12),
                    MerchantRankingWidget(merchants: merchants),
                    const SizedBox(height: 24),
                  ],

                  // Atalhos rápidos
                  Row(
                    children: [
                      Expanded(
                        child: _buildQuickAction(
                          icon: Icons.store_rounded,
                          label: 'MERCHANTS',
                          color: BlueprintTheme.accentTeal,
                          onTap: () => context.push('/merchants'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _buildQuickAction(
                          icon: Icons.calculate_rounded,
                          label: 'SIMULADOR',
                          color: BlueprintTheme.warning,
                          onTap: () => context.push('/simulator'),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),

                  // Replay Financeiro
                  GestureDetector(
                    onTap: () {
                      final now = DateTime.now();
                      final month = '${now.year}-${now.month.toString().padLeft(2, '0')}';
                      context.push('/replay/$month');
                    },
                    child: Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(20),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                          colors: [Color(0xFF3A2D7D), Color(0xFF1A1A24)],
                        ),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: BlueprintTheme.accentPurple.withValues(alpha: 0.5)),
                      ),
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: const BoxDecoration(color: BlueprintTheme.accentPurple, shape: BoxShape.circle),
                            child: const Icon(Icons.play_arrow_rounded, color: Colors.white),
                          ),
                          const SizedBox(width: 16),
                          const Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text('REPLAY_FINANCEIRO', style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.w900)),
                                SizedBox(height: 4),
                                Text('ASSISTIR_RETROSPECTIVA_DO_MÊS', style: TextStyle(color: BlueprintTheme.textSecondary, fontSize: 10, fontWeight: FontWeight.bold)),
                              ],
                            ),
                          ),
                          const Icon(Icons.chevron_right_rounded, color: BlueprintTheme.textSecondary),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Últimas transações
                  recentTxAsync.when(
                    data: (txList) => RecentTransactions(transactions: txList),
                    loading: () => const Center(child: CircularProgressIndicator()),
                    error: (e, _) => const SizedBox.shrink(),
                  ),
                  const SizedBox(height: 24),

                  // Activity Feed
                  feedAsync.when(
                    data: (feed) => ActivityFeedWidget(feedItems: feed),
                    loading: () => const SizedBox.shrink(),
                    error: (e, _) => const SizedBox.shrink(),
                  ),
                  const SizedBox(height: 32),
                ],
              ),
            );
          },
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (err, stack) => Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline_rounded, size: 48, color: BlueprintTheme.danger),
                const SizedBox(height: 16),
                Text('ERRO_DE_CARGA: $err', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary), textAlign: TextAlign.center),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () => ref.refresh(summaryProvider),
                  child: const Text('RECARREGAR'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildPeriodChip(WidgetRef ref, String value, String label, String current) {
    final isActive = current == value;
    return GestureDetector(
      onTap: () => ref.read(periodProvider.notifier).state = value,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
        decoration: BoxDecoration(
          color: isActive ? BlueprintTheme.accentPurple : BlueprintTheme.surface,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: isActive ? BlueprintTheme.accentPurple : BlueprintTheme.border),
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 10,
            fontWeight: FontWeight.w900,
            color: isActive ? Colors.white : BlueprintTheme.textSecondary,
          ),
        ),
      ),
    );
  }

  Widget _buildQuickAction({required IconData icon, required String label, required Color color, required VoidCallback onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: BlueprintTheme.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: BlueprintTheme.border),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(icon, color: color, size: 18),
            ),
            const SizedBox(width: 12),
            Text(label, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, color: color)),
          ],
        ),
      ),
    );
  }
}
