import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../../../shared/widgets/blueprint_card.dart';
import '../../../core/theme/blueprint_theme.dart';
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
    final stressAsync = ref.watch(stressScoreProvider);
    final survivalAsync = ref.watch(survivalModeProvider);
    final salaryPlanAsync = ref.watch(salaryPlanProvider);
    
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    final periods = [
      {'id': 'month', 'label': 'MÊS'},
      {'id': 'quarter', 'label': 'TRIMESTRE'},
      {'id': 'semester', 'label': 'SEMESTRE'},
      {'id': 'year', 'label': 'ANO'},
    ];

    // Handle Survival Mode Status Bar
    survivalAsync.whenData((data) {
      if (data['level'] == 'CRÍTICO') {
        SystemChrome.setSystemUIOverlayStyle(const SystemUiOverlayStyle(
          statusBarColor: Color(0xFF8B0000), // Dark Red
          statusBarIconBrightness: Brightness.light,
        ));
      } else {
        SystemChrome.setSystemUIOverlayStyle(SystemUiOverlayStyle.dark);
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text('DASHBOARD'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh, size: 20),
            onPressed: () {
              ref.refresh(summaryProvider);
              ref.refresh(stressScoreProvider);
              ref.refresh(survivalModeProvider);
              ref.refresh(salaryPlanProvider);
            },
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
            onRefresh: () async {
              ref.refresh(summaryProvider);
              ref.refresh(stressScoreProvider);
              ref.refresh(survivalModeProvider);
              ref.refresh(salaryPlanProvider);
            },
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
                        const SizedBox(height: 12),
                        
                        // Stress & Survival Indicators
                        Row(
                          children: [
                            stressAsync.maybeWhen(
                              data: (data) {
                                if (data.isEmpty) return const SizedBox.shrink();
                                final level = data['level'] ?? 'N/A';
                                final score = (data['score'] ?? 0.0).toDouble();
                                Color color = BlueprintTheme.accentTeal;
                                if (score < 40) color = BlueprintTheme.danger;
                                else if (score < 70) color = BlueprintTheme.warning;

                                return Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: color.withOpacity(0.1),
                                    border: Border.all(color: color, width: 1),
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Icon(Icons.psychology, size: 10, color: color),
                                      const SizedBox(width: 4),
                                      Text(
                                        'STRESS: ${level.toUpperCase()}',
                                        style: TextStyle(
                                          fontSize: 8,
                                          fontWeight: FontWeight.bold,
                                          color: color,
                                        ),
                                      ),
                                    ],
                                  ),
                                );
                              },
                              orElse: () => const SizedBox.shrink(),
                            ),
                            const SizedBox(width: 8),
                            survivalAsync.maybeWhen(
                              data: (data) {
                                if (data.isEmpty || data['level'] == 'TRANQUILO') return const SizedBox.shrink();
                                final level = data['level'] ?? 'N/A';
                                Color color = BlueprintTheme.warning;
                                if (level == 'CRÍTICO') color = BlueprintTheme.danger;

                                return Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: color,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Text(
                                    level.toUpperCase(),
                                    style: const TextStyle(
                                      fontSize: 8,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.white,
                                    ),
                                  ),
                                );
                              },
                              orElse: () => const SizedBox.shrink(),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),

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

                  // Daily Limit Widget
                  salaryPlanAsync.maybeWhen(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final dailyLimit = (data['safe_daily_limit'] ?? 0.0).toDouble();
                      final spentToday = summary.todaySpent;
                      final percent = (spentToday / dailyLimit).clamp(0.0, 1.0);
                      final isOver = spentToday > dailyLimit;

                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'LIMITE_DIÁRIO_SEGURO',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  Text(
                                    currencyFormat.format(dailyLimit),
                                    style: const TextStyle(
                                      fontSize: 20,
                                      fontWeight: FontWeight.w900,
                                      fontFamily: 'monospace',
                                    ),
                                  ),
                                  Text(
                                    '${(percent * 100).toInt()}%',
                                    style: TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                      color: isOver ? BlueprintTheme.danger : BlueprintTheme.textSecondary,
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Stack(
                                children: [
                                  Container(
                                    height: 8,
                                    width: double.infinity,
                                    decoration: BoxDecoration(
                                      color: BlueprintTheme.elevated,
                                      border: Border.all(color: BlueprintTheme.border),
                                    ),
                                  ),
                                  FractionallySizedBox(
                                    widthFactor: percent,
                                    child: Container(
                                      height: 8,
                                      decoration: BoxDecoration(
                                        color: isOver ? BlueprintTheme.danger : BlueprintTheme.accentPurple,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(
                                isOver 
                                  ? 'VOCÊ EXCEDEU O LIMITE RECOMENDADO PARA HOJE'
                                  : 'VOCÊ AINDA TEM ${currencyFormat.format(dailyLimit - spentToday)} DISPONÍVEIS HOJE',
                                style: TextStyle(
                                  fontSize: 8,
                                  fontWeight: FontWeight.bold,
                                  color: isOver ? BlueprintTheme.danger : BlueprintTheme.textSecondary,
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    orElse: () => const SizedBox.shrink(),
                  ),

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

                  // NUANCED INSIGHTS SECTION (MOVED UP)
                  Text(
                    'PIERRE_CORE_INTELLIGENCE',
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                      color: BlueprintTheme.textSecondary,
                      letterSpacing: 1.2,
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Meal Cost Insight
                  ref.watch(mealCostProvider).when(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final avgCost = data['avg_cost_per_meal'] ?? 0.0;
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'CUSTO_POR_REFEIÇÃO',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.restaurant, size: 16, color: BlueprintTheme.accentTeal),
                                  const SizedBox(width: 8),
                                  Text(
                                    'MÉDIA: ${currencyFormat.format(avgCost)}',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                data['insight'] ?? 'Analise seus gastos por canal (Delivery, Mercado, Restaurante).',
                                style: const TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

                  // Monthly Cycle Insight
                  ref.watch(salaryEffectProvider).when(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final spike = data['spending_increase_percent'] ?? 0;
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'CICLO_MENSAL',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.monetization_on, size: 16, color: BlueprintTheme.accentTeal),
                                  const SizedBox(width: 8),
                                  Text(
                                    'EFEITO SALÁRIO: +$spike%',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                data['insight'] ?? 'Seus gastos estabilizam cerca de 10 dias após o recebimento.',
                                style: const TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

                  const SizedBox(height: 8),

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

                  // NUANCED INSIGHTS SECTION
                  Text(
                    'NUANCED INSIGHTS',
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                      color: BlueprintTheme.textSecondary,
                      letterSpacing: 1.2,
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Meal Cost Insight
                  ref.watch(mealCostProvider).when(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final avgCost = data['avg_cost_per_meal'] ?? 0.0;
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'CUSTO_POR_REFEIÇÃO',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.restaurant, size: 16, color: BlueprintTheme.accentTeal),
                                  const SizedBox(width: 8),
                                  Text(
                                    'MÉDIA: ${currencyFormat.format(avgCost)}',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                data['insight'] ?? 'Analise seus gastos por canal (Delivery, Mercado, Restaurante).',
                                style: const TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

                  // Ticket Analysis Insight
                  ref.watch(ticketAnalysisProvider).when(
                    data: (data) {
                      if (data.isEmpty || data['top_decompositions'] == null) return const SizedBox.shrink();
                      final top = (data['top_decompositions'] as List).firstOrNull;
                      if (top == null) return const SizedBox.shrink();
                      final category = top['category'] ?? 'N/A';
                      final varTicket = (top['var_ticket'] ?? 0.0) * 100;
                      
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'ANÁLISE_DE_TICKET',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.label_important, size: 16, color: BlueprintTheme.accentPurple),
                                  const SizedBox(width: 8),
                                  Text(
                                    '$category: ${varTicket > 0 ? '+' : ''}${varTicket.toStringAsFixed(1)}% TICKET',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                data['insight_narrative'] ?? '',
                                maxLines: 3,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

                  // Loyalty Insight
                  ref.watch(loyaltyProvider).when(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final stats = data['loyalty_stats'] ?? {};
                      final leal = stats['leal'] ?? 0;
                      
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'LEALDADE_&_RECORRÊNCIA',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.favorite, size: 16, color: BlueprintTheme.danger),
                                  const SizedBox(width: 8),
                                  Text(
                                    '$leal ESTABELECIMENTOS LEAIS',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                data['insight_narrative'] ?? '',
                                maxLines: 3,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

                  // TEMPORAL ANALYTICS SECTION
                  Text(
                    'ANÁLISE TEMPORAL',
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                      color: BlueprintTheme.textSecondary,
                      letterSpacing: 1.2,
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Weekly Profile Insight
                  ref.watch(weeklyProfileProvider).when(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final peakDay = data['peak_spending_day'] ?? '';
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'PERFIL_SEMANAL',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.calendar_today, size: 16, color: BlueprintTheme.accentPurple),
                                  const SizedBox(width: 8),
                                  Text(
                                    'PICO: $peakDay',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              const Text(
                                'Seu maior volume de gastos acontece geralmente aos sábados, concentrado no período da tarde.',
                                style: TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

                  // Monthly Cycle Insight
                  ref.watch(salaryEffectProvider).when(
                    data: (data) {
                      if (data.isEmpty) return const SizedBox.shrink();
                      final spike = data['spending_increase_percent'] ?? 0;
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16.0),
                        child: BlueprintCard(
                          label: 'CICLO_MENSAL',
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  const Icon(Icons.monetization_on, size: 16, color: BlueprintTheme.accentTeal),
                                  const SizedBox(width: 8),
                                  Text(
                                    'EFEITO SALÁRIO: +$spike%',
                                    style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              Text(
                                data['insight'] ?? 'Seus gastos estabilizam cerca de 10 dias após o recebimento.',
                                style: const TextStyle(fontSize: 10, color: BlueprintTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                    loading: () => const SizedBox.shrink(),
                    error: (_, __) => const SizedBox.shrink(),
                  ),

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
