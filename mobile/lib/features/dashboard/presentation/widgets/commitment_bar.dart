import 'package:flutter/material.dart';

class CommitmentBar extends StatelessWidget {
  final double checking;
  final double credit;

  const CommitmentBar({super.key, required this.checking, required this.credit});

  @override
  Widget build(BuildContext context) {
    final double percentage = (checking > 0) ? (credit.abs() / checking).clamp(0.0, 1.0) : 1.0;
    final bool isWarning = percentage > 0.7;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text('INDICE_COMPROMETIMENTO', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
            Text('${(percentage * 100).toStringAsFixed(0)}%', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: isWarning ? Colors.red : Colors.black)),
          ],
        ),
        const SizedBox(height: 4),
        Container(
          height: 12,
          decoration: BoxDecoration(border: Border.all(color: Colors.black)),
          child: FractionallySizedBox(
            alignment: Alignment.centerLeft,
            widthFactor: percentage,
            child: Container(color: isWarning ? Colors.red : Colors.black),
          ),
        ),
      ],
    );
  }
}
