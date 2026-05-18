import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../../../core/theme/blueprint_theme.dart';
import '../../data/dashboard_models.dart';

class MerchantRankingWidget extends StatelessWidget {
  final List<MerchantSummary> merchants;

  const MerchantRankingWidget({super.key, required this.merchants});

  @override
  Widget build(BuildContext context) {
    if (merchants.isEmpty) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24.0),
          child: Text(
            'SEM_DADOS_MERCANTES',
            style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
          ),
        ),
      );
    }

    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return ListView.separated(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: merchants.length,
      separatorBuilder: (context, index) => const Divider(color: BlueprintTheme.border, height: 1),
      itemBuilder: (context, index) {
        final m = merchants[index];
        return Padding(
          padding: const EdgeInsets.symmetric(vertical: 12.0),
          child: Row(
            children: [
              Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                  color: BlueprintTheme.elevated,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Center(
                  child: Text(
                    '${index + 1}',
                    style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      m.merchantName.toUpperCase(),
                      style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w900),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      '${m.count} OPERAÇÕES',
                      style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
                    ),
                  ],
                ),
              ),
              Text(
                currencyFormat.format(m.total),
                style: const TextStyle(
                  fontSize: 12, 
                  fontWeight: FontWeight.w900, 
                  fontFamily: 'monospace',
                  color: BlueprintTheme.danger
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
