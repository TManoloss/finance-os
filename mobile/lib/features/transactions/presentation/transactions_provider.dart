import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/transactions/data/transaction_model.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

final transactionsProvider = FutureProvider<List<TransactionModel>>((ref) async {
  final api = ref.watch(apiClientProvider);
  final resp = await api.dio.get('/transactions?page=1&page_size=50');
  final List<dynamic> data = resp.data['data']['transactions'] ?? [];
  return data.map((json) => TransactionModel.fromJson(json)).toList();
});
