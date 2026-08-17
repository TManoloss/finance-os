import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/features/auth/presentation/auth_provider.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

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
        _errorMessage = 'PREENCHA TODOS OS CAMPOS';
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
        await ref.read(authProvider.notifier).fetchUser();
        if (mounted) {
          context.go('/dashboard');
        }
      } else {
        setState(() {
          _errorMessage = 'FALHA AO SALVAR CREDENCIAIS';
        });
      }
    } catch (e) {
      setState(() {
        _errorMessage = 'ERRO AO SALVAR AS CHAVES';
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
    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Logo/Ícone Open Finance
                Center(
                  child: Container(
                    width: 64,
                    height: 64,
                    decoration: BoxDecoration(
                      color: BlueprintTheme.accentPurple,
                      border: Border.all(color: BlueprintTheme.border, width: 2),
                      boxShadow: const [
                        BoxShadow(
                          color: BlueprintTheme.border,
                          offset: Offset(4, 4),
                        ),
                      ],
                    ),
                    child: const Center(
                      child: Icon(LucideIcons.wallet, color: Colors.white, size: 28),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                const Text(
                  'CONEXÃO_OPEN_FINANCE',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    fontFamily: 'monospace',
                    letterSpacing: -0.5,
                    color: BlueprintTheme.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'INICIALIZAR ADAPTADOR PLUGGY API',
                  textAlign: TextAlign.center,
                  style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 9),
                ),
                const SizedBox(height: 36),

                // Form Container
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.surface,
                    border: Border.all(color: BlueprintTheme.border, width: 2),
                    boxShadow: const [
                      BoxShadow(
                        color: BlueprintTheme.border,
                        offset: Offset(6, 6),
                      ),
                    ],
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _buildTerminalInput(
                        label: 'PLUGGY_CLIENT_ID',
                        controller: _clientIdController,
                        hintText: 'your_client_id_here',
                        icon: LucideIcons.key,
                      ),
                      const SizedBox(height: 16),
                      _buildTerminalInput(
                        label: 'PLUGGY_CLIENT_SECRET',
                        controller: _clientSecretController,
                        hintText: 'your_client_secret_here',
                        obscureText: true,
                        icon: LucideIcons.lock,
                      ),
                      const SizedBox(height: 24),

                      // Botão
                      _isLoading
                          ? const Center(
                              child: CircularProgressIndicator(
                                color: BlueprintTheme.accentPurple,
                              ),
                            )
                          : GestureDetector(
                              onTap: _handleSave,
                              child: Container(
                                height: 52,
                                decoration: BoxDecoration(
                                  color: BlueprintTheme.accentPurple,
                                  border: Border.all(color: BlueprintTheme.border, width: 2),
                                  boxShadow: const [
                                    BoxShadow(
                                      color: BlueprintTheme.border,
                                      offset: Offset(3, 3),
                                    ),
                                  ],
                                ),
                                child: const Center(
                                  child: Text(
                                    'INICIALIZAR_SINCRONIZADOR',
                                    style: TextStyle(
                                      fontWeight: FontWeight.w900,
                                      color: Colors.white,
                                      fontFamily: 'monospace',
                                      fontSize: 11,
                                      letterSpacing: 0.5,
                                    ),
                                  ),
                                ),
                              ),
                            ),
                    ],
                  ),
                ),

                const SizedBox(height: 24),
                if (_errorMessage != null)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: BlueprintTheme.danger.withValues(alpha: 0.1),
                        border: Border.all(color: BlueprintTheme.danger, width: 2),
                      ),
                      child: Text(
                        'ERRO: $_errorMessage',
                        textAlign: TextAlign.center,
                        style: terminalLabel(color: BlueprintTheme.danger, fontSize: 8),
                      ),
                    ),
                  ),

                Text(
                  'SYSTEM_STATUS: ${_isLoading ? "CONFIGURING_ADAPTER" : "AWAITING_CONFIG"}',
                  textAlign: TextAlign.center,
                  style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTerminalInput({
    required String label,
    required TextEditingController controller,
    required IconData icon,
    String? hintText,
    bool obscureText = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 10, color: BlueprintTheme.textSecondary),
            const SizedBox(width: 6),
            Text(
              label,
              style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 8),
            ),
          ],
        ),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: BlueprintTheme.elevated,
            border: Border.all(color: BlueprintTheme.border, width: 2),
          ),
          child: TextField(
            controller: controller,
            obscureText: obscureText,
            style: const TextStyle(fontSize: 12, fontFamily: 'monospace'),
            decoration: InputDecoration(
              border: InputBorder.none,
              hintText: hintText,
              hintStyle: TextStyle(color: BlueprintTheme.textSecondary.withValues(alpha: 0.3)),
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(vertical: 12),
            ),
          ),
        ),
      ],
    );
  }
}
