import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmController = TextEditingController();
  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _confirmController.dispose();
    super.dispose();
  }

  Future<void> _handleRegister() async {
    final name = _nameController.text.trim();
    final email = _emailController.text.trim();
    final password = _passwordController.text.trim();
    final confirm = _confirmController.text.trim();

    if (name.isEmpty || email.isEmpty || password.isEmpty) {
      setState(() => _error = 'PREENCHA TODOS OS CAMPOS OBRIGATÓRIOS');
      return;
    }
    if (password.length < 8) {
      setState(() => _error = 'SENHA DEVE TER NO MÍNIMO 8 CARACTERES');
      return;
    }
    if (password != confirm) {
      setState(() => _error = 'AS SENHAS NÃO COINCIDEM');
      return;
    }

    setState(() { _isLoading = true; _error = null; });

    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.post('/auth/register', data: {
        'name': name,
        'email': email,
        'password': password,
      });

      if (resp.statusCode == 200 || resp.statusCode == 201) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('REGISTRO CONCLUÍDO COM SUCESSO', style: terminalLabel(color: Colors.white)),
              backgroundColor: BlueprintTheme.success,
            ),
          );
          context.go('/login');
        }
      }
    } catch (e) {
      setState(() => _error = 'FALHA NO REGISTRO. EMAIL EM USO.');
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
            padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 20.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Logo Neo-Brutalist (Teal)
                Center(
                  child: Container(
                    width: 64,
                    height: 64,
                    decoration: const BoxDecoration(color: BlueprintTheme.accentTeal, shape: BoxShape.circle),
                    child: const Center(
                      child: Icon(LucideIcons.userPlus, color: Colors.white, size: 28),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                const Text(
                  'Crie sua conta',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.w900,
                    letterSpacing: -1,
                    color: BlueprintTheme.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Comece a organizar sua vida financeira.',
                  textAlign: TextAlign.center,
                  style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 9),
                ),
                const SizedBox(height: 36),

                // Form Container
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(color: BlueprintTheme.surface, borderRadius: BorderRadius.circular(20)),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _buildField('Nome completo', _nameController, false, LucideIcons.user),
                      const SizedBox(height: 14),
                      _buildField('E-mail', _emailController, false, LucideIcons.mail),
                      const SizedBox(height: 14),
                      _buildField('Senha', _passwordController, true, LucideIcons.lock),
                      const SizedBox(height: 14),
                      _buildField('Confirmar senha', _confirmController, true, LucideIcons.checkSquare),
                      
                      if (_error != null) ...[
                        const SizedBox(height: 16),
                        Container(
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: BlueprintTheme.danger.withValues(alpha: 0.1),
                            border: Border.all(color: BlueprintTheme.danger, width: 2),
                          ),
                          child: Text(
                            'ERROR: $_error',
                            style: terminalLabel(color: BlueprintTheme.danger, fontSize: 8),
                          ),
                        ),
                      ],

                      const SizedBox(height: 24),

                      // Button
                      GestureDetector(
                        onTap: _isLoading ? null : _handleRegister,
                        child: Container(
                          height: 52,
                          decoration: BoxDecoration(color: BlueprintTheme.accentTeal, borderRadius: BorderRadius.circular(14)),
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
                                    'Criar conta',
                                    style: TextStyle(
                                      fontWeight: FontWeight.w900,
                                      color: Colors.white,
                                      fontSize: 11,
                                      letterSpacing: 1,
                                    ),
                                  ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 24),
                TextButton(
                  onPressed: () => context.go('/login'),
                  child: Text(
                    'Já tem conta? Entrar',
                    style: terminalLabel(color: BlueprintTheme.textSecondary, fontSize: 9),
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
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: BlueprintTheme.elevated, borderRadius: BorderRadius.circular(12),
          ),
          child: TextField(
            controller: controller,
            obscureText: isPassword,
            style: const TextStyle(fontSize: 12),
            decoration: const InputDecoration(
              border: InputBorder.none,
              isDense: true,
              contentPadding: EdgeInsets.symmetric(vertical: 12),
            ),
          ),
        ),
      ],
    );
  }
}
