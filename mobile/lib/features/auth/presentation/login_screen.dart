import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
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

  Future<void> _handleLogin() async {
    final email = _emailController.text.trim();
    final password = _passwordController.text.trim();

    if (email.isEmpty || password.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('PREENCHA TODOS OS CAMPOS')),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final success = await ref.read(authProvider.notifier).login(email, password);
      if (!success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('ERRO: CREDENCIAIS INVÁLIDAS')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('ERRO_SISTEMA: $e')),
        );
      }
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
              // Logo Placeholder
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: BlueprintTheme.accentPurple,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Center(
                  child: Text(
                    'F', 
                    style: TextStyle(
                      fontSize: 32, 
                      fontWeight: FontWeight.w900, 
                      color: Colors.white
                    )
                  ),
                ),
              ),
              const SizedBox(height: 24),
              const Text(
                'FINANCE_OS',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 24, 
                  fontWeight: FontWeight.w900, 
                  letterSpacing: -1
                ),
              ),
              Text(
                'OPERATIONAL_SYSTEM_V1.1',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 10, 
                  fontWeight: FontWeight.bold, 
                  color: BlueprintTheme.textSecondary,
                  letterSpacing: 1.5
                ),
              ),
              const SizedBox(height: 48),

              // Inputs
              _buildField('EMAIL_ADDRESS', _emailController, false),
              const SizedBox(height: 16),
              _buildField('PASSWORD_KEY', _passwordController, true),
              const SizedBox(height: 32),

              // Login Button
              GestureDetector(
                onTap: _isLoading ? null : _handleLogin,
                child: Container(
                  height: 56,
                  decoration: BoxDecoration(
                    color: BlueprintTheme.accentPurple,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Center(
                    child: _isLoading 
                      ? const SizedBox(
                          width: 20, 
                          height: 20, 
                          child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2)
                        )
                      : const Text(
                          'AUTENTICAR', 
                          style: TextStyle(
                            fontWeight: FontWeight.w900, 
                            color: Colors.white, 
                            letterSpacing: 1
                          )
                        ),
                  ),
                ),
              ),
              
              const SizedBox(height: 24),
              TextButton(
                onPressed: () => context.go('/register'),
                child: Text(
                  'SOLICITAR_ACESSO',
                  style: TextStyle(
                    fontSize: 10, 
                    fontWeight: FontWeight.bold, 
                    color: BlueprintTheme.textSecondary
                  ),
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
        Text(
          label,
          style: TextStyle(
            fontSize: 8, 
            fontWeight: FontWeight.bold, 
            color: BlueprintTheme.textSecondary,
            letterSpacing: 1
          ),
        ),
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
            decoration: const InputDecoration(
              border: InputBorder.none,
            ),
          ),
        ),
      ],
    );
  }
}
