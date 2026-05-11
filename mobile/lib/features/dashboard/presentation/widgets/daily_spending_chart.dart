import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';

class DailySpendingChart extends StatelessWidget {
  final List<dynamic> byDay;

  const DailySpendingChart({super.key, required this.byDay});

  @override
  Widget build(BuildContext context) {
    return LineChart(
      LineChartData(
        gridData: const FlGridData(show: false),
        titlesData: const FlTitlesData(show: false),
        borderData: FlBorderData(show: false),
        lineBarsData: [
          LineChartBarData(
            spots: byDay.asMap().entries.map((e) {
              return FlSpot(e.key.toDouble(), (e.value['spent'] as num).toDouble());
            }).toList(),
            isCurved: false,
            color: Colors.black,
            barWidth: 2,
            dotData: const FlDotData(show: false),
          ),
        ],
      ),
    );
  }
}
