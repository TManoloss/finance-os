import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';

class HealthScreen extends ConsumerWidget {
  const HealthScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final healthAsync = ref.watch(healthScoreProvider);

    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      appBar: AppBar(title: const Text('HEALTH_DIAGNOSTIC_V1.0')),
      body: healthAsync.when(
        data: (data) {
          if (data.isEmpty) {
            return Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
              const Icon(LucideIcons.shieldAlert, size: 48, color: BlueprintTheme.textSecondary),
              const SizedBox(height: 16),
              Text('DADOS_INSUFICIENTES', style: terminalLabel()),
              const Text('Sincronize suas contas primeiro', style: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary)),
            ]));
          }

          final score = (data['overall_score'] as num?)?.toDouble() ?? 0;
          final dims = data['dimensions'] as Map<String, dynamic>? ?? {};
          final recommendations = data['recommendations'] as String? ?? data['analysis'] as String? ?? '';

          final dimensions = [
            _Dim('FLUXO_DE_CAIXA', (dims['cashflow'] as num?)?.toDouble() ?? 0, '25%'),
            _Dim('PARCELAMENTOS', (dims['installments'] as num?)?.toDouble() ?? 0, '20%'),
            _Dim('CONSISTÊNCIA', (dims['consistency'] as num?)?.toDouble() ?? 0, '20%'),
            _Dim('ASSINATURAS', (dims['subscriptions'] as num?)?.toDouble() ?? 0, '15%'),
            _Dim('DIVERSIFICAÇÃO', (dims['diversification'] as num?)?.toDouble() ?? 0, '10%'),
            _Dim('TENDÊNCIA', (dims['trend'] as num?)?.toDouble() ?? 0, '10%'),
          ];

          return ListView(
            children: [
              // Header neo-brutal
              Container(
                color: BlueprintTheme.elevated,
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
                child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                  Row(children: [
                    const Icon(LucideIcons.terminal, size: 12),
                    const SizedBox(width: 6),
                    Text('HEALTH_DIAGNOSTIC_V1.0', style: terminalLabel()),
                  ]),
                  const SizedBox(height: 4),
                  const Text('SCORE_DE_SAÚDE_FINANCEIRA', style: TextStyle(
                    fontFamily: 'monospace', fontSize: 18, fontWeight: FontWeight.w900, color: BlueprintTheme.textPrimary,
                  )),
                ]),
              ),
              Container(height: 2, color: BlueprintTheme.border),

              // Score central
              Container(
                padding: const EdgeInsets.all(32),
                color: BlueprintTheme.background,
                child: Row(children: [
                  // Gauge neo-brutal (quadrado com score)
                  Container(
                    width: 120, height: 120,
                    decoration: BoxDecoration(
                      border: Border.all(color: _scoreColor(score), width: 4),
                      boxShadow: [BoxShadow(color: _scoreColor(score), offset: const Offset(4, 4))],
                    ),
                    child: Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                      Text(score.toStringAsFixed(0), style: TextStyle(
                        fontFamily: 'monospace', fontSize: 40, fontWeight: FontWeight.w900, color: _scoreColor(score),
                      )),
                      Text('/100', style: terminalLabel(fontSize: 9)),
                    ])),
                  ),
                  const SizedBox(width: 20),
                  Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(_scoreLabel(score), style: TextStyle(
                      fontFamily: 'monospace', fontSize: 20, fontWeight: FontWeight.w900, color: _scoreColor(score),
                    )),
                    const SizedBox(height: 8),
                    Text('Seu score é baseado em padrões reais dos seus últimos 90 dias.',
                      style: const TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary, height: 1.4)),
                  ])),
                ]),
              ),
              Container(height: 2, color: BlueprintTheme.border),

              // Dimensões — grid 2 col
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
                child: Row(children: [
                  const Icon(LucideIcons.zap, size: 14, color: BlueprintTheme.accentPurple),
                  const SizedBox(width: 8),
                  Text('DIMENSÕES_DO_SCORE', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10)),
                ]),
              ),
              GridView.count(
                crossAxisCount: 2,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                childAspectRatio: 1.4,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                mainAxisSpacing: 12,
                crossAxisSpacing: 12,
                children: dimensions.map((d) => _dimensionCard(d)).toList(),
              ),
              const SizedBox(height: 20),
              Container(height: 2, color: BlueprintTheme.border),

              // Recomendações
              if (recommendations.isNotEmpty) ...[
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 20, 16, 12),
                  child: Row(children: [
                    const Icon(LucideIcons.zap, size: 14, color: BlueprintTheme.accentPurple),
                    const SizedBox(width: 8),
                    Text('RECOMENDAÇÕES_ESTRATÉGICAS', style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 10)),
                  ]),
                ),
                Container(
                  margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.surface,
                    border: Border.all(color: BlueprintTheme.border, width: 2),
                    boxShadow: const [BoxShadow(color: BlueprintTheme.border, offset: Offset(6, 6))],
                  ),
                  child: MarkdownBody(
                    data: recommendations,
                    styleSheet: MarkdownStyleSheet(
                      p: const TextStyle(color: BlueprintTheme.textPrimary, fontSize: 14, height: 1.6),
                      strong: const TextStyle(fontWeight: FontWeight.w900),
                      h2: const TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 15, color: BlueprintTheme.accentPurple),
                    ),
                  ),
                ),
              ],

              // Footer
              Container(
                padding: const EdgeInsets.all(16),
                color: BlueprintTheme.elevated,
                child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                  Text('HEALTH_SCORE_ENGINE // ONLINE', style: terminalLabel(fontSize: 8)),
                  Text('PRÓXIMO_CÁLCULO: 24H_CYCLE', style: terminalLabel(fontSize: 8)),
                ]),
              ),
            ],
          );
        },
        loading: () => const Center(child: CircularProgressIndicator(color: BlueprintTheme.accentPurple)),
        error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontFamily: 'monospace'))),
      ),
    );
  }

  Widget _dimensionCard(_Dim d) {
    final color = _scoreColor(d.score);
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        border: Border.all(color: BlueprintTheme.border, width: 2),
        boxShadow: const [BoxShadow(color: BlueprintTheme.border, offset: Offset(3, 3))],
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
          Expanded(child: Text(d.label, style: terminalLabel(fontSize: 8), overflow: TextOverflow.ellipsis)),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            color: BlueprintTheme.textPrimary,
            child: Text('PESO: ${d.weight}', style: const TextStyle(fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 7, color: BlueprintTheme.surface)),
          ),
        ]),
        const Spacer(),
        Text('${d.score.toStringAsFixed(0)}/100', style: moneyStyle(color: color, fontSize: 22)),
        const SizedBox(height: 6),
        // Progress bar com bordas retas
        Container(
          height: 6,
          decoration: BoxDecoration(border: Border.all(color: BlueprintTheme.border, width: 1), color: BlueprintTheme.background),
          child: FractionallySizedBox(
            widthFactor: (d.score / 100).clamp(0.0, 1.0),
            alignment: Alignment.centerLeft,
            child: Container(color: color),
          ),
        ),
      ]),
    );
  }

  Color _scoreColor(double score) {
    if (score >= 80) return BlueprintTheme.accentTeal;
    if (score >= 60) return BlueprintTheme.warning;
    return BlueprintTheme.danger;
  }

  String _scoreLabel(double score) {
    if (score >= 80) return 'EXCELENTE';
    if (score >= 60) return 'BOM';
    if (score >= 40) return 'ATENÇÃO';
    return 'CRÍTICO';
  }
}

class _Dim {
  final String label;
  final double score;
  final String weight;
  _Dim(this.label, this.score, this.weight);
}
