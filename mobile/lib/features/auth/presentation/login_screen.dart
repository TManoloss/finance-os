import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'auth_provider.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _isLoading = false;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _handleLogin() async {
    final email = _emailController.text.trim();
    final password = _passwordController.text.trim();

    if (email.isEmpty || password.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('PREENCHA TODOS OS CAMPOS', style: terminalLabel(color: Colors.white)),
          backgroundColor: BlueprintTheme.danger,
        ),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final success = await ref.read(authProvider.notifier).login(email, password);
      if (!success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('ERRO: CREDENCIAIS INVÁLIDAS', style: terminalLabel(color: Colors.white)),
            backgroundColor: BlueprintTheme.danger,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('ERRO_SISTEMA: $e', style: terminalLabel(color: Colors.white)),
            backgroundColor: BlueprintTheme.danger,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
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
                // Logo Neo-Brutalist
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
                      child: Text(
                        'F',
                        style: TextStyle(
                          fontSize: 32,
                          fontWeight: FontWeight.w900,
                          color: Colors.white,
                          fontFamily: 'monospace',
                        ),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                const Text(
                  'FINANCE_OS',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 26,
                    fontWeight: FontWeight.w900,
                    fontFamily: 'monospace',
                    letterSpacing: -1,
                    color: BlueprintTheme.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'OPERATIONAL_SYSTEM_V1.1',
                  textAlign: TextAlign.center,
                  style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 9),
                ),
                const SizedBox(height: 48),

                // Form Container Neo-Brutalist
                Container(
                  padding: const EdgeInsets.all(24),
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
                      _buildField('EMAIL_ADDRESS', _emailController, false, LucideIcons.mail),
                      const SizedBox(height: 20),
                      _buildField('PASSWORD_KEY', _passwordController, true, LucideIcons.lock),
                      const SizedBox(height: 28),

                      // Login Button
                      GestureDetector(
                        onTap: _isLoading ? null : _handleLogin,
                        child: Container(
                          height: 52,
                          decoration: BoxDecoration(
                            color: BlueprintTheme.accentPurple,
                            border: Border.all(color: BlueprintTheme.border, width: 2),
                            boxShadow: _isLoading
                                ? null
                                : const [
                                    BoxShadow(
                                      color: BlueprintTheme.border,
                                      offset: Offset(3, 3),
                                    ),
                                  ],
                          ),
                          child: Center(
                            child: _isLoading
                                ? const SizedBox(
                                    width: 18,
                                    height: 18,
                                    child: CircularProgressIndicator(
                                      color: Colors.white,
                                      strokeWidth: 2,
                                    ),
                                  )
                                : const Text(
                                    'AUTENTICAR',
                                    style: TextStyle(
                                      fontWeight: FontWeight.w900,
                                      color: Colors.white,
                                      fontFamily: 'monospace',
                                      fontSize: 12,
                                      letterSpacing: 1,
                                    ),
                                  ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 36),
                
                // Solicitar acesso
                Center(
                  child: GestureDetector(
                    onTap: () => context.go('/register'),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      decoration: BoxDecoration(
                        color: BlueprintTheme.elevated,
                        border: Border.all(color: BlueprintTheme.border, width: 2),
                      ),
                      child: Text(
                        'SOLICITAR_NOVO_ACESSO',
                        style: terminalLabel(color: BlueprintTheme.textPrimary, fontSize: 9),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildField(String label, TextEditingController controller, bool isPassword, IconData icon) {
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
        const SizedBox(height: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: BlueprintTheme.elevated,
            border: Border.all(color: BlueprintTheme.border, width: 2),
          ),
          child: TextField(
            controller: controller,
            obscureText: isPassword,
            style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
            decoration: const InputDecoration(
              border: InputBorder.none,
              isDense: true,
              contentPadding: EdgeInsets.symmetric(vertical: 14),
            ),
          ),
        ),
      ],
    );
  }
}
