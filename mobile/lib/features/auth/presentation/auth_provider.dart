import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/auth/data/user_model.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

class AuthState {
  final UserModel? user;
  final bool isAuthenticated;
  final bool isLoading;

  AuthState({
    this.user,
    this.isAuthenticated = false,
    this.isLoading = false,
  });

  AuthState copyWith({
    UserModel? user,
    bool? isAuthenticated,
    bool? isLoading,
  }) {
    return AuthState(
      user: user ?? this.user,
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      isLoading: isLoading ?? this.isLoading,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  final Ref ref;

  AuthNotifier(this.ref) : super(AuthState()) {
    _checkAuthStatus();
  }

  Future<void> _checkAuthStatus() async {
    state = state.copyWith(isLoading: true);
    final storage = ref.read(storageProvider);
    final token = await storage.read(key: 'access_token');

    if (token != null) {
      await fetchUser();
    } else {
      state = state.copyWith(isLoading: false, isAuthenticated: false);
    }
  }

  Future<void> fetchUser() async {
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.get('/me');
      final user = UserModel.fromJson(resp.data['data']);
      state = state.copyWith(
        user: user,
        isAuthenticated: true,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, isAuthenticated: false);
    }
  }

  Future<void> logout() async {
    final storage = ref.read(storageProvider);
    await storage.delete(key: 'access_token');
    state = AuthState();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref);
});
