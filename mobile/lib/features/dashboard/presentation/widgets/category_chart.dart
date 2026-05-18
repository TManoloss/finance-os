import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../../../../core/theme/blueprint_theme.dart';
import '../../data/dashboard_models.dart';

class CategoryChart extends StatelessWidget {
  final List<CategorySummary> categories;

  const CategoryChart({super.key, required this.categories});

  @override
  Widget build(BuildContext context) {
    if (categories.isEmpty) {
      return const Center(
        child: Text(
          'SEM_DADOS_CATEGORIAS',
          style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
        ),
      );
    }

    final sections = categories.map((cat) {
      final color = _parseColor(cat.color);
      return PieChartSectionData(
        color: color,
        value: cat.total,
        title: '${cat.percentage.toInt()}%',
        radius: 40,
        titleStyle: const TextStyle(
          fontSize: 8,
          fontWeight: FontWeight.bold,
          color: Colors.white,
        ),
      );
    }).toList();

    return Row(
      children: [
        Expanded(
          flex: 2,
          child: PieChart(
            PieChartData(
              sections: sections,
              centerSpaceRadius: 40,
              sectionsSpace: 2,
            ),
          ),
        ),
        const SizedBox(width: 24),
        Expanded(
          flex: 3,
          child: ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: categories.length.clamp(0, 6),
            separatorBuilder: (context, index) => const SizedBox(height: 8),
            itemBuilder: (context, index) {
              final cat = categories[index];
              return Row(
                children: [
                  Container(
                    width: 8,
                    height: 8,
                    decoration: BoxDecoration(
                      color: _parseColor(cat.color),
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      cat.categoryName.toUpperCase(),
                      style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textPrimary),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Text(
                    '${cat.percentage.toInt()}%',
                    style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
                  ),
                ],
              );
            },
          ),
        ),
      ],
    );
  }

  Color _parseColor(String? hexColor) {
    if (hexColor == null || hexColor.isEmpty) return BlueprintTheme.accentPurple;
    try {
      return Color(int.parse(hexColor.replaceFirst('#', '0xFF')));
    } catch (_) {
      return BlueprintTheme.accentPurple;
    }
  }
}
