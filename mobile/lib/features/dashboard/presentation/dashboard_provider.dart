import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'package:intl/intl.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:finance_os/core/api/api_client.dart';
import 'package:finance_os/features/dashboard/data/summary_model.dart';

final dioProvider = Provider((ref) => Dio());
final storageProvider = Provider((ref) => const FlutterSecureStorage());

final apiClientProvider = Provider((ref) => ApiClient(
      dio: ref.watch(dioProvider),
      storage: ref.watch(storageProvider),
    ));

final periodProvider = StateProvider<String>((ref) => 'month');

final summaryProvider = FutureProvider<FinancialSummary>((ref) async {
  final api = ref.watch(apiClientProvider);
  final period = ref.watch(periodProvider);
  
  String from = '';
  String to = '';
  final now = DateTime.now();

  if (period == 'month') {
    from = DateFormat('yyyy-MM-01').format(now);
    to = DateFormat('yyyy-MM-dd').format(now);
  } else if (period == 'quarter') {
    final startOfQuarter = DateTime(now.year, ((now.month - 1) ~/ 3) * 3 + 1, 1);
    from = DateFormat('yyyy-MM-dd').format(startOfQuarter);
    to = DateFormat('yyyy-MM-dd').format(now);
  } else if (period == 'semester') {
    final startOfSemester = now.subtract(const Duration(days: 180));
    from = DateFormat('yyyy-MM-dd').format(startOfSemester);
    to = DateFormat('yyyy-MM-dd').format(now);
  } else if (period == 'year') {
    from = DateFormat('yyyy-01-01').format(now);
    to = DateFormat('yyyy-MM-dd').format(now);
  }

  String url = '/transactions/summary';
  if (from.isNotEmpty && to.isNotEmpty && period != 'all') {
    url += '?from_date=$from&to_date=$to';
  }

  final resp = await api.dio.get(url);
  return FinancialSummary.fromJson(resp.data['data']);
});

final inflationProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/personal-inflation');
    return resp.data['data'] as Map<String, dynamic>;
  } catch (e) {
    return {};
  }
});
