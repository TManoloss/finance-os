import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class CategoryChart extends StatelessWidget {
  final List<dynamic> categories;
  const CategoryChart({super.key, required this.categories});

  @override
  Widget build(BuildContext context) {
    if (categories.isEmpty) {
      return const Center(
        child: Text(
          'NO_DATA_BUFFER',
          style: TextStyle(fontFamily: 'monospace', fontSize: 12),
        ),
      );
    }

    return PieChart(
      PieChartData(
        sectionsSpace: 2,
        centerSpaceRadius: 50,
        sections: categories.map((cat) {
          final percentage = (cat['percentage'] as num).toDouble();
          final index = categories.indexOf(cat);
          
          return PieChartSectionData(
            color: _getTechnicalColor(index),
            value: percentage,
            title: '${percentage.toStringAsFixed(0)}%',
            radius: 40,
            titleStyle: const TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          );
        }).toList(),
      ),
    );
  }

  Color _getTechnicalColor(int index) {
    const colors = [
      BlueprintTheme.accent, // 0xFF0000FF
      Color(0xFF666666),     // Tech Gray
      Color(0xFFAAAAAA),     // Soft Gray
      BlueprintTheme.border, // 0xFF000000
    ];
    return colors[index % colors.length];
  }
}
