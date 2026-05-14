import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';

class DailySpendingChart extends StatelessWidget {
  final List<dynamic> byDay;

  const DailySpendingChart({super.key, required this.byDay});

  @override
  Widget build(BuildContext context) {
    if (byDay.isEmpty) {
      return const Center(
        child: Text(
          'SEM_DADOS_DIARIOS',
          style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.grey),
        ),
      );
    }

    return LineChart(
      LineChartData(
        gridData: const FlGridData(
          show: true,
          drawVerticalLine: true,
          horizontalInterval: 100,
          verticalInterval: 1,
        ),
        titlesData: FlTitlesData(
          show: true,
          rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          leftTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 30,
              getTitlesWidget: (value, meta) => Text(
                value.toInt().toString(),
                style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold),
              ),
            ),
          ),
          bottomTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 20,
              getTitlesWidget: (value, meta) {
                final int index = value.toInt();
                if (index >= 0 && index < byDay.length) {
                  if (index % 5 == 0 || index == byDay.length - 1) {
                    return Text(
                      (index + 1).toString(),
                      style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold),
                    );
                  }
                }
                return const SizedBox.shrink();
              },
            ),
          ),
        ),
        lineTouchData: LineTouchData(
          touchTooltipData: LineTouchTooltipData(
            tooltipBgColor: Colors.black, // fl_chart v0.66 usa tooltipBgColor
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
        borderData: FlBorderData(
          show: true,
          border: const Border(
            bottom: BorderSide(color: Colors.black, width: 2),
            left: BorderSide(color: Colors.black, width: 2),
          ),
        ),
        lineBarsData: [
          // Recebidos (Crédito)
          LineChartBarData(
            spots: byDay.asMap().entries.map((e) {
              final received = (e.value['total_received'] ?? e.value['received'] ?? 0) as num;
              return FlSpot(e.key.toDouble(), received.toDouble());
            }).toList(),
            isCurved: false,
            color: const Color(0xFF4ECDC4),
            barWidth: 2,
            dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(
              show: true,
              color: const Color(0xFF4ECDC4).withOpacity(0.2),
            ),
          ),
          // Gastos (Débito)
          LineChartBarData(
            spots: byDay.asMap().entries.map((e) {
              final spent = (e.value['total_spent'] ?? e.value['spent'] ?? 0) as num;
              return FlSpot(e.key.toDouble(), spent.toDouble());
            }).toList(),
            isCurved: false,
            color: const Color(0xFFFF6B6B),
            barWidth: 2,
            dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(
              show: true,
              color: const Color(0xFFFF6B6B).withOpacity(0.2),
            ),
          ),
        ],
      ),
    );
  }
}
