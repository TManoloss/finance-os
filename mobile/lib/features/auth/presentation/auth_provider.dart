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
    final refreshToken = await storage.read(key: 'refresh_token');
    final token = await storage.read(key: 'access_token');

    if (refreshToken != null) {
      try {
        final api = ref.read(apiClientProvider);
        final resp = await api.dio
            .post('/auth/refresh', data: {'refresh_token': refreshToken});
        final data = resp.data['data'] as Map<String, dynamic>;
        await _saveSession(data);
        state = AuthState(
            user: UserModel.fromJson(data['user']), isAuthenticated: true);
        return;
      } catch (_) {
        await storage.delete(key: 'refresh_token');
      }
    }

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

  Future<bool> login(String email, String password) async {
    state = state.copyWith(isLoading: true);
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.post('/auth/login', data: {
        'email': email,
        'password': password,
      });

      if (resp.statusCode == 200) {
        final data = resp.data['data'] as Map<String, dynamic>;
        if (data['access_token'] != null && data['refresh_token'] != null) {
          await _saveSession(data);
          await fetchUser();
          return true;
        }
      }
      state = state.copyWith(isLoading: false);
      return false;
    } catch (e) {
      state = state.copyWith(isLoading: false);
      return false;
    }
  }

  Future<void> logout() async {
    final storage = ref.read(storageProvider);
    await storage.delete(key: 'access_token');
    await storage.delete(key: 'refresh_token');
    state = AuthState();
  }

  Future<void> _saveSession(Map<String, dynamic> data) async {
    final storage = ref.read(storageProvider);
    await storage.write(
        key: 'access_token', value: data['access_token'].toString());
    await storage.write(
        key: 'refresh_token', value: data['refresh_token'].toString());
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref);
});
