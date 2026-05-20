import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
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
            const SnackBar(content: Text('REGISTRO CONCLUÍDO COM SUCESSO')),
          );
          context.go('/login');
        }
      }
    } catch (e) {
      setState(() => _error = 'FALHA NO REGISTRO. EMAIL JÁ PODE ESTAR EM USO.');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(32.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: BlueprintTheme.accentTeal,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Center(
                  child: Icon(Icons.person_add_rounded, color: Colors.white, size: 32),
                ),
              ),
              const SizedBox(height: 24),
              const Text(
                'NOVO_OPERADOR',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 24, fontWeight: FontWeight.w900, letterSpacing: -1),
              ),
              const Text(
                'CRIAR_CONTA_NO_SISTEMA',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 1.5),
              ),
              const SizedBox(height: 40),

              _buildField('NOME_COMPLETO', _nameController, false),
              const SizedBox(height: 16),
              _buildField('EMAIL_ADDRESS', _emailController, false),
              const SizedBox(height: 16),
              _buildField('PASSWORD_KEY', _passwordController, true),
              const SizedBox(height: 16),
              _buildField('CONFIRM_PASSWORD', _confirmController, true),

              if (_error != null) ...[
                const SizedBox(height: 16),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.danger.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: BlueprintTheme.danger.withValues(alpha: 0.3)),
                  ),
                  child: Text(
                    'ERROR: $_error',
                    style: const TextStyle(color: BlueprintTheme.danger, fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                ),
              ],

              const SizedBox(height: 32),
              GestureDetector(
                onTap: _isLoading ? null : _handleRegister,
                child: Container(
                  height: 56,
                  decoration: BoxDecoration(
                    color: BlueprintTheme.accentTeal,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Center(
                    child: _isLoading
                      ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                      : const Text('REGISTRAR_OPERADOR', style: TextStyle(fontWeight: FontWeight.w900, color: Colors.white, letterSpacing: 1)),
                  ),
                ),
              ),

              const SizedBox(height: 24),
              TextButton(
                onPressed: () => context.go('/login'),
                child: const Text(
                  'JÁ POSSUI ACESSO? AUTENTICAR',
                  style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildField(String label, TextEditingController controller, bool isPassword) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 1)),
        const SizedBox(height: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          decoration: BoxDecoration(
            color: BlueprintTheme.elevated,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: BlueprintTheme.border, width: 1),
          ),
          child: TextField(
            controller: controller,
            obscureText: isPassword,
            style: const TextStyle(fontSize: 14),
            decoration: const InputDecoration(border: InputBorder.none),
          ),
        ),
      ],
    );
  }
}
