import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';

class HealthScreen extends ConsumerWidget {
  const HealthScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final healthAsync = ref.watch(healthScoreProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('SAÚDE_FINANCEIRA')),
      body: healthAsync.when(
        data: (data) {
          if (data.isEmpty) {
            return const Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.health_and_safety_outlined, size: 64, color: BlueprintTheme.textSecondary),
                  SizedBox(height: 16),
                  Text('DADOS_INSUFICIENTES', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: BlueprintTheme.textSecondary)),
                  Text('Sincronize suas contas primeiro', style: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary)),
                ],
              ),
            );
          }

          final score = (data['overall_score'] as num?)?.toDouble() ?? 0;
          final dims = data['dimensions'] as Map<String, dynamic>? ?? {};
          final recommendations = data['recommendations'] as String? ?? data['analysis'] as String? ?? '';

          final dimensions = [
            _Dimension('FLUXO_DE_CAIXA', (dims['cashflow'] as num?)?.toDouble() ?? 0, Icons.trending_up_rounded, '25%'),
            _Dimension('PARCELAMENTOS', (dims['installments'] as num?)?.toDouble() ?? 0, Icons.calendar_today_rounded, '20%'),
            _Dimension('CONSISTÊNCIA', (dims['consistency'] as num?)?.toDouble() ?? 0, Icons.show_chart_rounded, '15%'),
            _Dimension('ASSINATURAS', (dims['subscriptions'] as num?)?.toDouble() ?? 0, Icons.autorenew_rounded, '15%'),
            _Dimension('DIVERSIFICAÇÃO', (dims['diversification'] as num?)?.toDouble() ?? 0, Icons.pie_chart_rounded, '15%'),
            _Dimension('TENDÊNCIA', (dims['trend'] as num?)?.toDouble() ?? 0, Icons.insights_rounded, '10%'),
          ];

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                // Score Principal
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(32),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        _getScoreColor(score).withValues(alpha: 0.15),
                        BlueprintTheme.surface,
                      ],
                    ),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: _getScoreColor(score).withValues(alpha: 0.3)),
                  ),
                  child: Column(
                    children: [
                      SizedBox(
                        width: 160, height: 160,
                        child: CustomPaint(
                          painter: _ScoreGaugePainter(score: score, color: _getScoreColor(score)),
                          child: Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  score.toStringAsFixed(0),
                                  style: TextStyle(fontSize: 48, fontWeight: FontWeight.w900, fontFamily: 'monospace', color: _getScoreColor(score)),
                                ),
                                Text(_getScoreLabel(score), style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w900, color: BlueprintTheme.textSecondary)),
                              ],
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      const Text('SCORE_DE_SAÚDE_FINANCEIRA', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 1)),
                    ],
                  ),
                ),

                const SizedBox(height: 24),

                // Dimensões
                const Align(
                  alignment: Alignment.centerLeft,
                  child: Text('DIMENSÕES', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
                ),
                const SizedBox(height: 12),
                ...dimensions.map((dim) => _buildDimensionCard(dim)),

                // Recomendações
                if (recommendations.isNotEmpty) ...[
                  const SizedBox(height: 24),
                  const Align(
                    alignment: Alignment.centerLeft,
                    child: Text('ANÁLISE_DO_AGENTE', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
                  ),
                  const SizedBox(height: 12),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: BlueprintTheme.surface,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: BlueprintTheme.border),
                    ),
                    child: MarkdownBody(
                      data: recommendations,
                      styleSheet: MarkdownStyleSheet(
                        p: const TextStyle(color: BlueprintTheme.textPrimary, fontSize: 13, height: 1.5),
                        strong: const TextStyle(fontWeight: FontWeight.bold, color: BlueprintTheme.accentTeal),
                      ),
                    ),
                  ),
                ],
                const SizedBox(height: 32),
              ],
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger))),
      ),
    );
  }

  Widget _buildDimensionCard(_Dimension dim) {
    final color = _getScoreColor(dim.score);
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Row(
        children: [
          Icon(dim.icon, color: color, size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(dim.label, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 0.5)),
                    Text('${dim.score.toStringAsFixed(0)} / 100', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w900, fontFamily: 'monospace', color: color)),
                  ],
                ),
                const SizedBox(height: 6),
                ClipRRect(
                  borderRadius: BorderRadius.circular(3),
                  child: LinearProgressIndicator(
                    value: (dim.score / 100).clamp(0.0, 1.0),
                    minHeight: 5,
                    backgroundColor: BlueprintTheme.elevated,
                    valueColor: AlwaysStoppedAnimation(color),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Text(dim.weight, style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary)),
        ],
      ),
    );
  }

  Color _getScoreColor(double score) {
    if (score >= 80) return BlueprintTheme.accentTeal;
    if (score >= 60) return BlueprintTheme.success;
    if (score >= 40) return BlueprintTheme.warning;
    return BlueprintTheme.danger;
  }

  String _getScoreLabel(double score) {
    if (score >= 80) return 'EXCELENTE';
    if (score >= 60) return 'BOM';
    if (score >= 40) return 'ATENÇÃO';
    return 'CRÍTICO';
  }
}

class _Dimension {
  final String label;
  final double score;
  final IconData icon;
  final String weight;
  _Dimension(this.label, this.score, this.icon, this.weight);
}

class _ScoreGaugePainter extends CustomPainter {
  final double score;
  final Color color;
  _ScoreGaugePainter({required this.score, required this.color});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2 - 8;

    // Background arc
    final bgPaint = Paint()
      ..color = BlueprintTheme.elevated
      ..strokeWidth = 10
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      2.3, // start angle
      4.6, // sweep (270 degrees in radians)
      false,
      bgPaint,
    );

    // Score arc
    final scorePaint = Paint()
      ..color = color
      ..strokeWidth = 10
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      2.3,
      4.6 * (score / 100).clamp(0.0, 1.0),
      false,
      scorePaint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}
