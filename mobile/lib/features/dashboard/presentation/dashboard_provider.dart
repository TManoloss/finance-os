import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:finance_os/core/api/api_client.dart';
import 'package:finance_os/features/dashboard/data/summary_model.dart';

final dioProvider = Provider((ref) => Dio());
final storageProvider = Provider((ref) => const FlutterSecureStorage());

final apiClientProvider = Provider((ref) => ApiClient(
      dio: ref.watch(dioProvider),
      storage: ref.watch(storageProvider),
    ));

final summaryProvider = FutureProvider<FinancialSummary>((ref) async {
  final api = ref.watch(apiClientProvider);
  final resp = await api.dio.get('/transactions/summary');
  return FinancialSummary.fromJson(resp.data['data']);
});
