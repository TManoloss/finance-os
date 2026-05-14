import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/features/auth/presentation/auth_provider.dart';

class PluggySetupScreen extends ConsumerStatefulWidget {
  const PluggySetupScreen({super.key});

  @override
  ConsumerState<PluggySetupScreen> createState() => _PluggySetupScreenState();
}

class _PluggySetupScreenState extends ConsumerState<PluggySetupScreen> {
  final _clientIdController = TextEditingController();
  final _clientSecretController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void dispose() {
    _clientIdController.dispose();
    _clientSecretController.dispose();
    super.dispose();
  }

  Future<void> _handleSave() async {
    final clientId = _clientIdController.text.trim();
    final clientSecret = _clientSecretController.text.trim();

    if (clientId.isEmpty || clientSecret.isEmpty) {
      setState(() {
        _errorMessage = 'PLEASE_FILL_ALL_FIELDS';
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.post(
        '/accounts/keys',
        data: {
          'client_id': clientId,
          'client_secret': clientSecret,
        },
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        // Refresh user info to update pluggyClientId status
        await ref.read(authProvider.notifier).fetchUser();
        if (mounted) {
          context.go('/dashboard');
        }
      } else {
        setState(() {
          _errorMessage = 'FAILED_TO_SAVE_KEYS';
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = 'ERROR_OCCURRED_WHILE_SAVING';
      });
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: const Color(0xFF0A0A0F),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 32.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'PLUGGY_INITIAL_SETUP',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                  letterSpacing: -0.5,
                  color: Colors.white,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'OPEN_FINANCE_INTEGRATION_REQUIRED',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 2,
                  color: Colors.white.withValues(alpha: 0.6),
                ),
              ),
              const SizedBox(height: 48),
              _buildTerminalInput(
                label: 'PLUGGY_CLIENT_ID',
                controller: _clientIdController,
                hintText: 'your_client_id_here',
              ),
              const SizedBox(height: 24),
              _buildTerminalInput(
                label: 'PLUGGY_CLIENT_SECRET',
                controller: _clientSecretController,
                hintText: 'your_client_secret_here',
                obscureText: true,
              ),
              const SizedBox(height: 48),
              _isLoading
                  ? const Center(child: CircularProgressIndicator(color: Color(0xFF7C6FFF)))
                  : ElevatedButton(
                      onPressed: _handleSave,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF7C6FFF),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: const RoundedRectangleBorder(
                          borderRadius: BorderRadius.zero,
                        ),
                      ),
                      child: const Text('INITIALIZE_SYNC_ENGINE'),
                    ),
              const SizedBox(height: 16),
              if (_errorMessage != null)
                Padding(
                  padding: const EdgeInsets.only(bottom: 16),
                  child: Text(
                    'ERROR_SETUP: $_errorMessage',
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      color: Color(0xFFFF6B6B),
                      fontWeight: FontWeight.bold,
                      fontSize: 10,
                    ),
                  ),
                ),
              Text(
                'SYSTEM_STATUS: ${_isLoading ? "CONFIGURING_ADAPTER" : "AWAITING_CONFIG"} // READY',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 10,
                  color: Colors.white.withValues(alpha: 0.4),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTerminalInput({
    required String label,
    required TextEditingController controller,
    String? hintText,
    bool obscureText = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 10,
            fontWeight: FontWeight.bold,
            letterSpacing: 1.5,
            color: Colors.white,
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: controller,
          obscureText: obscureText,
          style: const TextStyle(color: Colors.white),
          cursorColor: const Color(0xFF7C6FFF),
          decoration: InputDecoration(
            hintText: hintText,
            hintStyle: TextStyle(color: Colors.white.withValues(alpha: 0.2)),
            filled: true,
            fillColor: Colors.white.withValues(alpha: 0.05),
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
            enabledBorder: const OutlineInputBorder(
              borderSide: BorderSide(color: Color(0xFF2A2A3A), width: 1),
              borderRadius: BorderRadius.zero,
            ),
            focusedBorder: const OutlineInputBorder(
              borderSide: BorderSide(color: Color(0xFF7C6FFF), width: 1),
              borderRadius: BorderRadius.zero,
            ),
          ),
        ),
      ],
    );
  }
}
