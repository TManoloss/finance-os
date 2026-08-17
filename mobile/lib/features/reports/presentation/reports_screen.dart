import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/shared/widgets/premium_page.dart';

// --- RIVERPOD PROVIDERS ---

final reportsListProvider = FutureProvider<List<dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports');
    return (resp.data['data'] ?? []) as List<dynamic>;
  } catch (_) {
    return [];
  }
});

final gamificationProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/gamification');
    return (resp.data['data'] ?? {}) as Map<String, dynamic>;
  } catch (_) {
    return {};
  }
});

final salaryPlanProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/salary-plan');
    if (resp.data['data'] == null) return null;
    return (resp.data['data']) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final personalInflationProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/personal-inflation');
    return (resp.data['data'] ?? resp.data) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final silentGrowthProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/silent-growth');
    return (resp.data['data'] ?? resp.data) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final impulseProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/impulse');
    return (resp.data['data'] ?? resp.data) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final compensationProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/compensation-pattern');
    return (resp.data['data'] ?? resp.data) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final mealCostProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/meal-cost');
    return (resp.data['data'] ?? resp.data) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final convenienceProvider = FutureProvider<Map<String, dynamic>?>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/convenience-index');
    return (resp.data['data'] ?? resp.data) as Map<String, dynamic>;
  } catch (_) {
    return null;
  }
});

final installmentTimelineProvider = FutureProvider<dynamic>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/installment-timeline');
    return resp.data['data'] ?? resp.data;
  } catch (_) {
    return null;
  }
});

class ReportsScreen extends ConsumerStatefulWidget {
  const ReportsScreen({super.key});

  @override
  ConsumerState<ReportsScreen> createState() => _ReportsScreenState();
}

class _ReportsScreenState extends ConsumerState<ReportsScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _generating = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _triggerReport(String type) async {
    setState(() => _generating = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/reports/trigger/$type');
      ref.invalidate(reportsListProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('RELATÓRIO_${type.toUpperCase()}_DISPARADO_NO_PIERRE', style: terminalLabel(color: Colors.white)),
            backgroundColor: BlueprintTheme.success,
          ),
        );
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: BlueprintTheme.danger,
            content: Text('FALHA_AO_DISPARAR_AGENTE', style: terminalLabel(color: Colors.white)),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _generating = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      appBar: AppBar(
        title: const Text('Relatórios'),
        bottom: TabBar(
          controller: _tabController,
          dividerColor: BlueprintTheme.border,
          indicatorColor: BlueprintTheme.accentPurple,
          indicatorWeight: 3,
          labelColor: BlueprintTheme.accentPurple,
          unselectedLabelColor: BlueprintTheme.textSecondary,
          isScrollable: true,
          tabs: const [
            Tab(text: 'Pierre'), Tab(text: 'Insights'), Tab(text: 'Metas'), Tab(text: 'Cronograma'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildCognitiveReportsTab(),
          _buildAdvancedMetricsTab(),
          _buildGamificationTab(),
          _buildInstallmentsTab(),
        ],
      ),
    );
  }

  // --- ABA 1: IA COGNITIVA (RELATÓRIOS DO PIERRE) ---
  Widget _buildCognitiveReportsTab() {
    final reportsAsync = ref.watch(reportsListProvider);

    return reportsAsync.when(
      data: (reports) {
        return RefreshIndicator(
          onRefresh: () async => ref.invalidate(reportsListProvider),
          color: BlueprintTheme.accentPurple,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // Botões de Ação Brutalistas para Disparar Agentes
              PremiumCard(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Atualizar análises', style: const TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: _buildTriggerButton('daily', 'DIÁRIO', LucideIcons.sparkles),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: _buildTriggerButton('weekly', 'SEMANAL', LucideIcons.calendar),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: _buildTriggerButton('monthly', 'MENSAL', LucideIcons.barChart3),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),

              if (reports.isEmpty)
                _buildEmptyState(
                  'NÚCLEO_DE_DADOS_VAZIO',
                  'O Pierre ainda está estruturando suas informações financeiras. Dispare um relatório acima para acelerar o processo!',
                )
              else
                ...reports.map((report) {
                  final agentType = (report['agent_type'] ?? 'daily').toString().toUpperCase();
                  final summary = (report['summary_markdown'] ?? '').toString();
                  final insights = report['insights'] is List ? report['insights'] as List<dynamic> : const [];
                  final rawDate = report['created_at'] != null ? DateTime.tryParse(report['created_at'].toString()) : null;
                  final formattedDate = rawDate != null ? DateFormat("dd/MM/yyyy HH:mm").format(rawDate) : 'DATA_INDISPONÍVEL';

                  return Padding(padding: const EdgeInsets.only(bottom: 16), child: PremiumCard(
                    child: ExpansionTile(
                      collapsedBackgroundColor: BlueprintTheme.surface,
                      backgroundColor: BlueprintTheme.surface,
                      iconColor: BlueprintTheme.accentPurple,
                      title: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(6),
                            color: agentType == 'DAILY' ? BlueprintTheme.accentPurple : BlueprintTheme.accentTeal,
                            child: Icon(
                              agentType == 'DAILY' ? LucideIcons.sparkles : LucideIcons.trendingUp,
                              size: 14,
                              color: Colors.white,
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Relatório $agentType', style: const TextStyle(fontWeight: FontWeight.w700),
                                ),
                                Text(
                                  formattedDate,
                                  style: const TextStyle(fontSize: 9, color: BlueprintTheme.textSecondary, fontFamily: 'monospace'),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      children: [
                        Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              // Resumo
                              Container(
                                padding: const EdgeInsets.all(12),
                                decoration: neoBrutalCard(backgroundColor: BlueprintTheme.elevated),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Resumo executivo', style: terminalLabel(fontSize: 10),
                                    ),
                                    const SizedBox(height: 8),
                                    MarkdownBody(
                                      data: summary,
                                      styleSheet: MarkdownStyleSheet(
                                        p: const TextStyle(fontSize: 12, height: 1.5, color: BlueprintTheme.textPrimary),
                                        strong: const TextStyle(fontWeight: FontWeight.w900),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(height: 16),

                              // Insights
                              if (insights.isNotEmpty) ...[
                                Text(
                                  'DETALHAMENTO_INSIGHTS',
                                  style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                                ),
                                const SizedBox(height: 8),
                                ...insights.map((insight) {
                                  return Container(
                                    margin: const EdgeInsets.only(bottom: 6),
                                    padding: const EdgeInsets.all(10),
                                    decoration: BoxDecoration(
                                      color: BlueprintTheme.surface,
                                      border: Border.all(color: BlueprintTheme.border, width: 2),
                                    ),
                                    child: Row(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        const Icon(LucideIcons.zap, size: 12, color: BlueprintTheme.warning),
                                        const SizedBox(width: 8),
                                        Expanded(
                                          child: Text(
                                            insight.toString(),
                                            style: const TextStyle(fontSize: 11, color: BlueprintTheme.textPrimary),
                                          ),
                                        ),
                                      ],
                                    ),
                                  );
                                }),
                              ],
                            ],
                          ),
                        ),
                      ],
                    ),
                  ));
                }),
            ],
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator(color: BlueprintTheme.accentPurple)),
      error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace'))),
    );
  }

  Widget _buildTriggerButton(String type, String label, IconData icon) {
    return GestureDetector(
      onTap: _generating ? null : () => _triggerReport(type),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: BlueprintTheme.surface,
          border: Border.all(color: BlueprintTheme.border, width: 2),
          boxShadow: const [
            BoxShadow(
              color: BlueprintTheme.border,
              offset: Offset(2, 2),
            ),
          ],
        ),
        child: Column(
          children: [
            Icon(icon, size: 14, color: BlueprintTheme.accentPurple),
            const SizedBox(height: 4),
            Text(
              label,
              style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 8),
            ),
          ],
        ),
      ),
    );
  }

  // --- ABA 2: PADRÕES E MÉTRICAS IA AVANÇADAS ---
  Widget _buildAdvancedMetricsTab() {
    final inflationAsync = ref.watch(personalInflationProvider);
    final growthAsync = ref.watch(silentGrowthProvider);
    final impulseAsync = ref.watch(impulseProvider);
    final compensationAsync = ref.watch(compensationProvider);
    final mealCostAsync = ref.watch(mealCostProvider);
    final convenienceAsync = ref.watch(convenienceProvider);

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(personalInflationProvider);
        ref.invalidate(silentGrowthProvider);
        ref.invalidate(impulseProvider);
        ref.invalidate(compensationProvider);
        ref.invalidate(mealCostProvider);
        ref.invalidate(convenienceProvider);
      },
      color: BlueprintTheme.accentPurple,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Row(
            children: [
              const Icon(LucideIcons.trendingUp, size: 14, color: BlueprintTheme.accentPurple),
              const SizedBox(width: 8),
              Text(
                'INSIGHTS_DE_PADRÃO_V2.5',
                style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // 1. Inflação Pessoal
          inflationAsync.when(
            data: (data) => _buildMetricCard(
              title: 'INFLAÇÃO_PESSOAL',
              icon: LucideIcons.activity,
              color: BlueprintTheme.accentPurple,
              child: data != null
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(
                              '${data['inflation_rate_percent'] ?? '0'}%',
                              style: moneyStyle(color: BlueprintTheme.danger, fontSize: 24),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'TAXA_ANUAL_PROJETADA',
                              style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          data['narrative']?.toString() ?? 'Seus gastos em categorias essenciais registraram aumento de custo nos últimos meses.',
                          style: const TextStyle(fontSize: 11, height: 1.4),
                        ),
                      ],
                    )
                  : const Text('Sem dados calculados de inflação pessoal pelo Pierre ainda.', style: TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary)),
            ),
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 16),

          // 2. Crescimento Silencioso
          growthAsync.when(
            data: (data) => _buildMetricCard(
              title: 'CRESCIMENTO_SILENCIOSO',
              icon: LucideIcons.trendingUp,
              color: BlueprintTheme.accentTeal,
              child: data != null
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(
                              data['growth_detected'] == true ? 'VULNERÁVEL' : 'SAUDÁVEL',
                              style: moneyStyle(
                                color: data['growth_detected'] == true ? BlueprintTheme.danger : BlueprintTheme.success,
                                fontSize: 18,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          data['narrative']?.toString() ?? 'Detecção de pequenos gastos cumulativos recorrentes que se expandiram no último ciclo.',
                          style: const TextStyle(fontSize: 11, height: 1.4),
                        ),
                      ],
                    )
                  : const Text('Gastos silenciosos sob controle. Nenhuma anomalia crítica detectada.', style: TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary)),
            ),
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 16),

          // 3. Índice de Conveniência
          convenienceAsync.when(
            data: (data) => _buildMetricCard(
              title: 'ÍNDICE_DE_CONVENIÊNCIA',
              icon: LucideIcons.shoppingBag,
              color: BlueprintTheme.accentPurple,
              child: data != null
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(
                              '${data['convenience_percentage'] ?? '0'}%',
                              style: moneyStyle(color: BlueprintTheme.accentPurple, fontSize: 24),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'DOS_GASTOS_DE_VULNERABILIDADE',
                              style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Total gasto em aplicativos de conveniência/delivery: R\$ ${data['total_spent'] ?? '0'}. Pierre aconselha racionalização.',
                          style: const TextStyle(fontSize: 11, height: 1.4),
                        ),
                      ],
                    )
                  : const Text('Sem histórico de compras em canais de conveniência este mês.', style: TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary)),
            ),
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 16),

          // 4. Custo de Refeição
          mealCostAsync.when(
            data: (data) => _buildMetricCard(
              title: 'CUSTO_MÉDIO_POR_REFEIÇÃO',
              icon: LucideIcons.utensils,
              color: BlueprintTheme.accentTeal,
              child: data != null
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(
                              'R\$ ${(data['average_cost'] ?? 0.0).toStringAsFixed(2)}',
                              style: moneyStyle(color: BlueprintTheme.accentTeal, fontSize: 22),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'VALOR_MÉDIO_DETECTADO',
                              style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Alimentação fora de casa representa uma parcela significativa. Pierre sugere preparo doméstico nos finais de semana.',
                          style: const TextStyle(fontSize: 11, height: 1.4),
                        ),
                      ],
                    )
                  : const Text('Poucas transações em alimentação registradas para análise de ticket médio.', style: TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary)),
            ),
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 16),

          // 5. Análise de Impulso
          impulseAsync.when(
            data: (data) => _buildMetricCard(
              title: 'ANÁLISE_DE_IMPULSOS',
              icon: LucideIcons.zap,
              color: BlueprintTheme.warning,
              child: data != null
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(
                              '${data['impulsive_txs_count'] ?? '0'} GASTOS',
                              style: moneyStyle(color: BlueprintTheme.warning, fontSize: 18),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'POR_IMPULSO_REGISTRADOS',
                              style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          data['narrative']?.toString() ?? 'Foram detectadas compras suspeitas fora do horário habitual de consumo ou em fins de semana de pico.',
                          style: const TextStyle(fontSize: 11, height: 1.4),
                        ),
                      ],
                    )
                  : const Text('Nenhum padrão crítico de impulso detectado nas últimas 4 semanas.', style: TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary)),
            ),
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 16),

          // 6. Padrões de Compensação
          compensationAsync.when(
            data: (data) => _buildMetricCard(
              title: 'PADRÃO_DE_COMPENSAÇÃO',
              icon: LucideIcons.alertTriangle,
              color: BlueprintTheme.danger,
              child: data != null
                  ? Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          data['detected'] == true ? 'SINDROME_DE_COMPENSAÇÃO: ATIVA' : 'SINDROME: INATIVA',
                          style: terminalLabel(color: data['detected'] == true ? BlueprintTheme.danger : BlueprintTheme.textPrimary, fontSize: 9),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          data['narrative']?.toString() ?? 'Detecção de surtos de despesas imediatamente após picos de stress profissional ou depósito de salários.',
                          style: const TextStyle(fontSize: 11, height: 1.4),
                        ),
                      ],
                    )
                  : const Text('Nenhum padrão reativo ou de compensação detectado.', style: TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary)),
            ),
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildMetricCard({
    required String title,
    required IconData icon,
    required Color color,
    required Widget child,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        border: Border.all(color: BlueprintTheme.border, width: 2),
        boxShadow: const [
          BoxShadow(
            color: BlueprintTheme.border,
            offset: Offset(4, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Icon(icon, size: 14, color: color),
              const SizedBox(width: 8),
              Text(
                title,
                style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 9),
              ),
            ],
          ),
          const Divider(color: BlueprintTheme.border, height: 20, thickness: 1),
          child,
        ],
      ),
    );
  }

  Widget _buildCardSkeleton() {
    return Container(
      height: 120,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        border: Border.all(color: BlueprintTheme.border, width: 2),
      ),
      child: const Center(
        child: SizedBox(
          width: 20,
          height: 20,
          child: CircularProgressIndicator(color: BlueprintTheme.accentPurple, strokeWidth: 2),
        ),
      ),
    );
  }

  // --- ABA 3: GAMIFICAÇÃO & PLANOS ---
  Widget _buildGamificationTab() {
    final gamificationAsync = ref.watch(gamificationProvider);
    final salaryPlanAsync = ref.watch(salaryPlanProvider);

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(gamificationProvider);
        ref.invalidate(salaryPlanProvider);
      },
      color: BlueprintTheme.accentPurple,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Seção Plano de Salário
          salaryPlanAsync.when(
            data: (plan) {
              if (plan == null) return const SizedBox();

              final detected = plan['salary_detected'] ?? 0.0;
              final commitments = plan['fixed_commitments'] ?? 0.0;
              final dailyLimit = plan['safe_daily_limit'] ?? 0.0;
              final planData = plan['plan_data'] ?? {};
              final narrative = planData['narrative']?.toString() ?? 'Seu planejamento de alocação inteligente do Pierre para o mês corrente.';
              final allocation = planData['allocation'] as Map<String, dynamic>? ?? {};

              return Container(
                margin: const EdgeInsets.only(bottom: 24),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: BlueprintTheme.surface,
                  border: Border.all(color: BlueprintTheme.border, width: 2),
                  boxShadow: const [
                    BoxShadow(
                      color: BlueprintTheme.border,
                      offset: Offset(4, 4),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      children: [
                        const Icon(LucideIcons.target, size: 16, color: BlueprintTheme.accentPurple),
                        const SizedBox(width: 8),
                        Text(
                          'PLANO_DE_SALÁRIO_INTELIGENTE',
                          style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10),
                        ),
                      ],
                    ),
                    const Divider(color: BlueprintTheme.border, height: 24, thickness: 1),
                    
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        _buildPlanStat('SALÁRIO_DETECTADO', 'R\$ ${detected.toStringAsFixed(2)}'),
                        _buildPlanStat('LIMITES_FIXOS', 'R\$ ${commitments.toStringAsFixed(2)}'),
                        _buildPlanStat('LIMITE_DIÁRIO_SEGURO', 'R\$ ${dailyLimit.toStringAsFixed(2)}'),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text(
                      narrative,
                      style: const TextStyle(fontSize: 11, height: 1.4, fontStyle: FontStyle.italic),
                    ),

                    if (allocation.isNotEmpty) ...[
                      const SizedBox(height: 16),
                      Text(
                        'DISTRIBUIÇÃO_RECOMENDADA',
                        style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                      ),
                      const SizedBox(height: 8),
                      ...allocation.entries.map((e) {
                        return Padding(
                          padding: const EdgeInsets.symmetric(vertical: 4),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(e.key.toUpperCase(), style: const TextStyle(fontFamily: 'monospace', fontSize: 10, fontWeight: FontWeight.bold)),
                              Text('${e.value}%', style: terminalLabel(color: BlueprintTheme.accentPurple, fontSize: 10)),
                            ],
                          ),
                        );
                      }),
                    ],
                  ],
                ),
              );
            },
            loading: () => _buildCardSkeleton(),
            error: (_, __) => const SizedBox(),
          ),

          gamificationAsync.when(
            data: (data) {
              final activeMissions = data['active_missions'] as List<dynamic>? ?? [];
              final achievements = data['awarded_achievements'] as List<dynamic>? ?? [];

              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Missões Ativas
                  Row(
                    children: [
                      const Icon(LucideIcons.zap, size: 14, color: BlueprintTheme.accentPurple),
                      const SizedBox(width: 8),
                      Text(
                        'MISSÕES_ATIVAS_DE_ECONOMIA',
                        style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 9),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  if (activeMissions.isEmpty)
                    _buildSubEmptyState('Sem missões financeiras ativas no momento.')
                  else
                    ...activeMissions.map((m) {
                      final title = m['title']?.toString() ?? 'MISSÃO_DE_GASTOS';
                      final desc = m['description']?.toString() ?? '';
                      final target = double.tryParse(m['target_value']?.toString() ?? '0.0') ?? 0.0;
                      final current = double.tryParse(m['current_value']?.toString() ?? '0.0') ?? 0.0;
                      
                      final progress = target > 0 ? (current / target).clamp(0.0, 1.0) : 0.0;

                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        padding: const EdgeInsets.all(14),
                        decoration: BoxDecoration(
                          color: BlueprintTheme.surface,
                          border: Border.all(color: BlueprintTheme.border, width: 2),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            Text(
                              title,
                              style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 9),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              desc,
                              style: const TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary),
                            ),
                            const SizedBox(height: 10),
                            LinearProgressIndicator(
                              value: progress,
                              backgroundColor: BlueprintTheme.elevated,
                              color: BlueprintTheme.accentTeal,
                              minHeight: 8,
                            ),
                            const SizedBox(height: 6),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'PROGRESSO: ${(progress * 100).toStringAsFixed(0)}%',
                                  style: const TextStyle(fontFamily: 'monospace', fontSize: 9, fontWeight: FontWeight.bold),
                                ),
                                Text(
                                  'R\$ ${current.toStringAsFixed(2)} / R\$ ${target.toStringAsFixed(2)}',
                                  style: const TextStyle(fontFamily: 'monospace', fontSize: 9, fontWeight: FontWeight.bold, color: BlueprintTheme.accentTeal),
                                ),
                              ],
                            ),
                          ],
                        ),
                      );
                    }),

                  const SizedBox(height: 24),

                  // Conquistas Ganhas
                  Row(
                    children: [
                      const Icon(LucideIcons.award, size: 14, color: BlueprintTheme.accentTeal),
                      const SizedBox(width: 8),
                      Text(
                        'DISTINTIVOS_CONQUISTADOS',
                        style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 9),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),

                  if (achievements.isEmpty)
                    _buildSubEmptyState('Nenhum distintivo conquistado neste mês ainda. Mantenha as finanças alinhadas!')
                  else
                    GridView.builder(
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        crossAxisSpacing: 8,
                        mainAxisSpacing: 8,
                        childAspectRatio: 2.2,
                      ),
                      itemCount: achievements.length,
                      itemBuilder: (context, idx) {
                        final ach = achievements[idx];
                        final id = ach['achievement_id']?.toString() ?? 'FINANCE_MASTER';

                        return Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                          decoration: BoxDecoration(
                            color: BlueprintTheme.elevated,
                            border: Border.all(color: BlueprintTheme.border, width: 2),
                          ),
                          child: Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(6),
                                color: BlueprintTheme.accentTeal.withValues(alpha: 0.1),
                                child: const Icon(LucideIcons.award, size: 16, color: BlueprintTheme.accentTeal),
                              ),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  id.toUpperCase().replaceAll('_', ' '),
                                  style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 8),
                                  overflow: TextOverflow.ellipsis,
                                  maxLines: 2,
                                ),
                              ),
                            ],
                          ),
                        );
                      },
                    ),
                ],
              );
            },
            loading: () => const Center(child: CircularProgressIndicator(color: BlueprintTheme.accentPurple)),
            error: (_, __) => const SizedBox(),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _buildPlanStat(String label, String val) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 7)),
        const SizedBox(height: 4),
        Text(val, style: moneyStyle(color: BlueprintTheme.accentPurple, fontSize: 13)),
      ],
    );
  }

  Widget _buildSubEmptyState(String msg) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.elevated,
        border: Border.all(color: BlueprintTheme.border, width: 1),
      ),
      child: Text(
        msg,
        style: const TextStyle(fontSize: 11, fontStyle: FontStyle.italic, color: BlueprintTheme.textSecondary),
        textAlign: TextAlign.center,
      ),
    );
  }

  // --- ABA 4: CRONOGRAMA DE PARCELAMENTOS ---
  Widget _buildInstallmentsTab() {
    final timelineAsync = ref.watch(installmentTimelineProvider);

    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(installmentTimelineProvider),
      color: BlueprintTheme.accentPurple,
      child: timelineAsync.when(
        data: (data) {
          if (data == null || (data is List && data.isEmpty)) {
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                _buildEmptyState(
                  'CRONOGRAMA_LIMPO',
                  'Não foram encontradas compras parceladas ativas ou registradas em seus cartões no momento.',
                ),
              ],
            );
          }

          // Se for uma lista de meses
          final list = data as List<dynamic>;

          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: list.length,
            itemBuilder: (context, idx) {
              final monthData = list[idx] as Map<String, dynamic>;
              final monthName = monthData['month']?.toString() ?? 'MES_FUTURO';
              final totalAmount = double.tryParse(monthData['total_amount']?.toString() ?? '0.0') ?? 0.0;
              final items = monthData['items'] as List<dynamic>? ?? [];

              return Container(
                margin: const EdgeInsets.only(bottom: 16),
                decoration: BoxDecoration(
                  color: BlueprintTheme.surface,
                  border: Border.all(color: BlueprintTheme.border, width: 2),
                  boxShadow: const [
                    BoxShadow(
                      color: BlueprintTheme.border,
                      offset: Offset(4, 4),
                    ),
                  ],
                ),
                child: ExpansionTile(
                  collapsedBackgroundColor: BlueprintTheme.surface,
                  backgroundColor: BlueprintTheme.surface,
                  title: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        monthName.toUpperCase(),
                        style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 11),
                      ),
                      Text(
                        'R\$ ${totalAmount.toStringAsFixed(2)}',
                        style: moneyStyle(color: BlueprintTheme.danger, fontSize: 14),
                      ),
                    ],
                  ),
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      color: BlueprintTheme.elevated,
                      child: Column(
                        children: items.map((item) {
                          final desc = item['description']?.toString() ?? 'COMPRA_PARCELADA';
                          final inst = item['installment']?.toString() ?? '';
                          final val = double.tryParse(item['amount']?.toString() ?? '0.0') ?? 0.0;

                          return Padding(
                            padding: const EdgeInsets.symmetric(vertical: 6),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        desc,
                                        style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: BlueprintTheme.textPrimary),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                      if (inst.isNotEmpty)
                                        Text(
                                          inst,
                                          style: const TextStyle(fontSize: 9, color: BlueprintTheme.textSecondary, fontFamily: 'monospace'),
                                        ),
                                    ],
                                  ),
                                ),
                                Text(
                                  'R\$ ${val.toStringAsFixed(2)}',
                                  style: const TextStyle(fontSize: 11, fontFamily: 'monospace', fontWeight: FontWeight.bold),
                                ),
                              ],
                            ),
                          );
                        }).toList(),
                      ),
                    ),
                  ],
                ),
              );
            },
          );
        },
        loading: () => const Center(child: CircularProgressIndicator(color: BlueprintTheme.accentPurple)),
        error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace'))),
      ),
    );
  }

  Widget _buildEmptyState(String title, String desc) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        color: BlueprintTheme.elevated,
        border: Border.all(color: BlueprintTheme.border, width: 2),
      ),
      child: Column(
        children: [
          const Icon(LucideIcons.fileText, size: 40, color: BlueprintTheme.textSecondary),
          const SizedBox(height: 12),
          Text(title, style: terminalLabel()),
          const SizedBox(height: 8),
          Text(
            desc,
            style: const TextStyle(fontSize: 11, color: BlueprintTheme.textSecondary),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
