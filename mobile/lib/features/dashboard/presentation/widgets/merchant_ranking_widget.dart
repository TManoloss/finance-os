import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class MerchantRankingWidget extends StatelessWidget {
  final List<dynamic> merchants;

  const MerchantRankingWidget({super.key, required this.merchants});

  @override
  Widget build(BuildContext context) {
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (var i = 0; i < merchants.length; i++)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 4.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('${i + 1}. ${merchants[i]['merchant_name']}', style: const TextStyle(fontSize: 12)),
                Text(currencyFormat.format(merchants[i]['total']), style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
      ],
    );
  }
}
