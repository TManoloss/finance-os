import '../../../core/api/api_client.dart';

class AuthRepository {
  final ApiClient apiClient;

  AuthRepository({required this.apiClient});

  Future<bool> login(String email, String password) async {
    try {
      final response = await apiClient.dio.post(
        '/auth/login',
        data: {
          'email': email,
          'password': password,
        },
      );

      if (response.statusCode == 200) {
        final accessToken = response.data['data']['access_token'];
        if (accessToken != null) {
          await apiClient.storage.write(key: 'access_token', value: accessToken);
          return true;
        }
      }
      return false;
    } catch (e) {
      return false;
    }
  }
}
