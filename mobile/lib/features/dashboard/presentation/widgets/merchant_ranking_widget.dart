import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class MerchantRankingWidget extends StatelessWidget {
  final List<dynamic> merchants;

  const MerchantRankingWidget({super.key, required this.merchants});

  @override
  Widget build(BuildContext context) {
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    
    if (merchants.isEmpty) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(16.0),
          child: Text(
            'SEM_DADOS_DE_MERCANTES',
            style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.grey),
          ),
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (var i = 0; i < merchants.length; i++)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 6.0),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    '${i + 1}. ${merchants[i]['merchant_name']}', 
                    style: const TextStyle(
                      fontSize: 12, 
                      fontWeight: FontWeight.bold,
                      overflow: TextOverflow.ellipsis,
                    ),
                    maxLines: 1,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  currencyFormat.format(merchants[i]['total']), 
                  style: const TextStyle(
                    fontSize: 12, 
                    fontWeight: FontWeight.w900, 
                    color: Colors.redAccent,
                    fontFamily: 'monospace',
                  ),
                ),
              ],
            ),
          ),
      ],
    );
  }
}
