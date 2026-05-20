import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

/// Provider para listar contas conectadas do usuário
final connectedAccountsProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/accounts');
    final data = resp.data['data'];
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return [];
  } catch (e) {
    return [];
  }
});

/// Provider para buscar top merchants
final merchantsProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/merchants?limit=20');
    final data = resp.data['data'];
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return [];
  } catch (e) {
    return [];
  }
});

/// Provider para health score
final healthScoreProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports/health-score');
    return resp.data['data'] as Map<String, dynamic>;
  } catch (e) {
    return {};
  }
});

/// Provider para relatórios dos agentes
final agentReportsProvider = FutureProvider.family<Map<String, dynamic>, String>((ref, agentType) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/reports?type=$agentType');
    return resp.data['data'] as Map<String, dynamic>;
  } catch (e) {
    return {};
  }
});

/// Provider para cartões e parcelamentos
final installmentsProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/cards/installments');
    final data = resp.data['data'];
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return [];
  } catch (e) {
    return [];
  }
});

/// Provider para assinaturas detectadas
final subscriptionsProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final api = ref.watch(apiClientProvider);
  try {
    final resp = await api.dio.get('/cards/subscriptions');
    final data = resp.data['data'];
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return [];
  } catch (e) {
    return [];
  }
});
