import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/transactions/presentation/transactions_provider.dart';
import 'package:intl/intl.dart';

class TransactionsScreen extends ConsumerWidget {
  const TransactionsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final txsAsync = ref.watch(transactionsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('SYSTEM_LOG // TRANSACTIONS')),
      body: txsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator(color: Colors.black)),
        error: (err, stack) => Center(child: Text('ERR_LOAD: $err')),
        data: (transactions) {
          if (transactions.isEmpty) return const Center(child: Text('BUFFER_EMPTY'));
          
          return ListView.separated(
            itemCount: transactions.length,
            separatorBuilder: (context, index) => const Divider(color: Colors.black, height: 2, thickness: 2),
            itemBuilder: (context, index) {
              final tx = transactions[index];
              final isCredit = tx.direction == 'credit';
              final color = isCredit ? const Color(0xFF008000) : const Color(0xFFD00000);
              final sign = isCredit ? '+' : '-';
              
              String formattedDate = tx.date;
              try {
                final dt = DateTime.parse(tx.date);
                formattedDate = DateFormat('dd/MM').format(dt);
              } catch (_) {}

              return Container(
                padding: const EdgeInsets.all(16),
                color: const Color(0xFFF4F1EA),
                child: Row(
                  children: [
                    Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.black, width: 2),
                        color: isCredit ? color : Colors.transparent,
                      ),
                      alignment: Alignment.center,
                      child: Text(
                        sign,
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                          color: isCredit ? Colors.white : color,
                        ),
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(tx.description, style: const TextStyle(fontWeight: FontWeight.w900), maxLines: 1, overflow: TextOverflow.ellipsis),
                          const SizedBox(height: 4),
                          Text('DATE: $formattedDate | CAT: ${tx.categoryName ?? "NULL"}', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$').format(tx.amount),
                      style: TextStyle(fontWeight: FontWeight.w900, fontSize: 16, color: color),
                    ),
                  ],
                ),
              );
            },
          );
        },
      ),
    );
  }
}
