import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

/// Provider para relatórios do agente, por tipo
final reportProvider = FutureProvider.family<Map<String, dynamic>, String>((ref, type) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports?type=$type');
    return (resp.data['data'] ?? {}) as Map<String, dynamic>;
  } catch (e) {
    return {};
  }
});

/// Provider para gerar relatório sob demanda
final triggerReportProvider = FutureProvider.family<Map<String, dynamic>, String>((ref, type) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.post('/reports/trigger', data: {'type': type});
    return (resp.data['data'] ?? {}) as Map<String, dynamic>;
  } catch (e) {
    return {};
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
    _tabController = TabController(length: 3, vsync: this);
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
      await api.dio.post('/reports/trigger', data: {'type': type});
      ref.invalidate(reportProvider(type));
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('RELATÓRIO GERADO COM SUCESSO')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('FALHA AO GERAR RELATÓRIO'), backgroundColor: BlueprintTheme.danger),
        );
      }
    } finally {
      if (mounted) setState(() => _generating = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('RELATÓRIOS'),
        bottom: TabBar(
          controller: _tabController,
          labelStyle: const TextStyle(fontWeight: FontWeight.w900, fontSize: 10, letterSpacing: 0.5),
          unselectedLabelStyle: const TextStyle(fontWeight: FontWeight.bold, fontSize: 10),
          indicatorColor: BlueprintTheme.accentPurple,
          labelColor: BlueprintTheme.accentPurple,
          unselectedLabelColor: BlueprintTheme.textSecondary,
          tabs: const [
            Tab(text: 'DIÁRIO'),
            Tab(text: 'SEMANAL'),
            Tab(text: 'MENSAL'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildReportTab('daily'),
          _buildReportTab('weekly'),
          _buildReportTab('monthly'),
        ],
      ),
    );
  }

  Widget _buildReportTab(String type) {
    final reportAsync = ref.watch(reportProvider(type));

    return reportAsync.when(
      data: (data) {
        final summary = (data['summary_markdown'] ?? data['summary'] ?? '').toString();
        final insights = data['insights'];

        return SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Botão de gerar
              GestureDetector(
                onTap: _generating ? null : () => _triggerReport(type),
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.accentPurple.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: BlueprintTheme.accentPurple.withValues(alpha: 0.3)),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      if (_generating)
                        const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: BlueprintTheme.accentPurple))
                      else
                        const Icon(Icons.auto_awesome_rounded, size: 18, color: BlueprintTheme.accentPurple),
                      const SizedBox(width: 8),
                      Text(
                        _generating ? 'GERANDO...' : 'GERAR_RELATÓRIO_${type.toUpperCase()}',
                        style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 11, color: BlueprintTheme.accentPurple, letterSpacing: 0.5),
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 24),

              if (summary.isNotEmpty) ...[
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.surface,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: BlueprintTheme.border),
                  ),
                  child: MarkdownBody(
                    data: summary,
                    styleSheet: MarkdownStyleSheet(
                      p: const TextStyle(color: BlueprintTheme.textPrimary, fontSize: 14, height: 1.6),
                      h1: const TextStyle(color: BlueprintTheme.textPrimary, fontSize: 22, fontWeight: FontWeight.w900),
                      h2: const TextStyle(color: BlueprintTheme.accentPurple, fontSize: 18, fontWeight: FontWeight.w900),
                      h3: const TextStyle(color: BlueprintTheme.textPrimary, fontSize: 16, fontWeight: FontWeight.bold),
                      strong: const TextStyle(fontWeight: FontWeight.bold, color: BlueprintTheme.accentTeal),
                      listBullet: const TextStyle(color: BlueprintTheme.textSecondary),
                      code: const TextStyle(backgroundColor: BlueprintTheme.elevated, fontFamily: 'monospace', fontSize: 12, color: BlueprintTheme.accentTeal),
                      blockquote: const TextStyle(color: BlueprintTheme.textSecondary, fontStyle: FontStyle.italic),
                    ),
                  ),
                ),
              ] else ...[
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(40),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.surface,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: BlueprintTheme.border),
                  ),
                  child: const Column(
                    children: [
                      Icon(Icons.description_outlined, size: 48, color: BlueprintTheme.textSecondary),
                      SizedBox(height: 16),
                      Text('NENHUM_RELATÓRIO_DISPONÍVEL', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: BlueprintTheme.textSecondary)),
                      SizedBox(height: 4),
                      Text('Clique no botão acima para gerar', style: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary)),
                    ],
                  ),
                ),
              ],

              // Insights
              if (insights is List && insights.isNotEmpty) ...[
                const SizedBox(height: 24),
                const Text('INSIGHTS', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
                const SizedBox(height: 12),
                ...insights.map((insight) {
                  final msg = insight is String ? insight : (insight['message'] ?? insight.toString());
                  return Container(
                    margin: const EdgeInsets.only(bottom: 8),
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: BlueprintTheme.elevated,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: BlueprintTheme.border),
                    ),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Icon(Icons.lightbulb_outline_rounded, size: 16, color: BlueprintTheme.warning),
                        const SizedBox(width: 8),
                        Expanded(child: Text(msg.toString(), style: const TextStyle(fontSize: 12, height: 1.4))),
                      ],
                    ),
                  );
                }),
              ],
              const SizedBox(height: 32),
            ],
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger))),
    );
  }
}
