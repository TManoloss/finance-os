import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../../../../core/theme/blueprint_theme.dart';
import '../../data/dashboard_models.dart';

class DailySpendingChart extends StatelessWidget {
  final List<DailyBalance> byDay;

  const DailySpendingChart({super.key, required this.byDay});

  @override
  Widget build(BuildContext context) {
    if (byDay.isEmpty) {
      return const Center(
        child: Text(
          'SEM_DADOS_TELEMETRIA',
          style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
        ),
      );
    }

    final spots = byDay.asMap().entries.map((e) {
      return FlSpot(e.key.toDouble(), e.value.totalSpent);
    }).toList();

    return LineChart(
      LineChartData(
        gridData: FlGridData(
          show: true,
          drawVerticalLine: false,
          getDrawingHorizontalLine: (value) => FlLine(
            color: BlueprintTheme.border.withValues(alpha: 0.1),
            strokeWidth: 1,
          ),
        ),
        titlesData: FlTitlesData(
          show: true,
          rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          bottomTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 22,
              interval: (byDay.length / 5).clamp(1, 31).toDouble(),
              getTitlesWidget: (value, meta) {
                if (value.toInt() >= 0 && value.toInt() < byDay.length) {
                  final date = byDay[value.toInt()].date;
                  return Padding(
                    padding: const EdgeInsets.only(top: 8.0),
                    child: Text(
                      '${date.day}/${date.month}',
                      style: const TextStyle(color: BlueprintTheme.textSecondary, fontSize: 8, fontWeight: FontWeight.bold),
                    ),
                  );
                }
                return const Text('');
              },
            ),
          ),
          leftTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 40,
              getTitlesWidget: (value, meta) {
                return Text(
                  value.toInt().toString(),
                  style: const TextStyle(color: BlueprintTheme.textSecondary, fontSize: 8, fontWeight: FontWeight.bold),
                );
              },
            ),
          ),
        ),
        borderData: FlBorderData(show: false),
        lineBarsData: [
          LineChartBarData(
            spots: spots,
            isCurved: true,
            color: BlueprintTheme.accentPurple,
            barWidth: 3,
            isStrokeCapRound: true,
            dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(
              show: true,
              gradient: LinearGradient(
                colors: [
                  BlueprintTheme.accentPurple.withValues(alpha: 0.3),
                  BlueprintTheme.accentPurple.withValues(alpha: 0),
                ],
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
              ),
            ),
          ),
        ],
        lineTouchData: LineTouchData(
          touchTooltipData: LineTouchTooltipData(
            tooltipBgColor: BlueprintTheme.elevated,
            tooltipBorder: const BorderSide(color: BlueprintTheme.border, width: 1),
            getTooltipItems: (touchedSpots) {
              return touchedSpots.map((spot) {
                return LineTooltipItem(
                  'R\$ ${spot.y.toStringAsFixed(2)}',
                  const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 10),
                );
              }).toList();
            },
          ),
        ),
      ),
    );
  }
}
