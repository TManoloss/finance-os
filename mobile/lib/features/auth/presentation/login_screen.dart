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
  bool _obscurePassword = true;

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
          content: Text('PREENCHA TODOS OS CAMPOS',
              style: terminalLabel(color: Colors.white)),
          backgroundColor: BlueprintTheme.danger,
        ),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final success =
          await ref.read(authProvider.notifier).login(email, password);
      if (!success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('ERRO: CREDENCIAIS INVÁLIDAS',
                style: terminalLabel(color: Colors.white)),
            backgroundColor: BlueprintTheme.danger,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('ERRO_SISTEMA: $e',
                style: terminalLabel(color: Colors.white)),
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
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(24, 40, 24, 28),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Align(
                alignment: Alignment.centerLeft,
                child: Container(
                  width: 52,
                  height: 52,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(17),
                    gradient: const LinearGradient(colors: [
                      BlueprintTheme.accentPurple,
                      Color(0xFFB195FF)
                    ]),
                  ),
                  child: const Center(
                      child: Text('F',
                          style: TextStyle(
                              fontSize: 25,
                              fontWeight: FontWeight.w800,
                              color: Colors.white))),
                ),
              ),
              const SizedBox(height: 36),
              const Text('Bem-vindo de volta',
                  style: TextStyle(
                      fontSize: 30,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -1)),
              const SizedBox(height: 8),
              Text('Entre para acompanhar sua vida financeira.',
                  style: TextStyle(
                      color: BlueprintTheme.textSecondary.withValues(alpha: .9),
                      fontSize: 15)),
              const SizedBox(height: 34),
              Container(
                padding: const EdgeInsets.all(22),
                decoration: BoxDecoration(
                  color: BlueprintTheme.surface,
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(color: BlueprintTheme.border),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    _buildField(
                        'E-mail', _emailController, false, LucideIcons.mail),
                    const SizedBox(height: 18),
                    _buildField(
                        'Senha', _passwordController, true, LucideIcons.lock),
                    const SizedBox(height: 24),
                    FilledButton(
                      onPressed: _isLoading ? null : _handleLogin,
                      style: FilledButton.styleFrom(
                        minimumSize: const Size.fromHeight(54),
                        backgroundColor: BlueprintTheme.accentPurple,
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16)),
                      ),
                      child: _isLoading
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                  color: Colors.white, strokeWidth: 2))
                          : const Text('Entrar',
                              style: TextStyle(
                                  fontWeight: FontWeight.w700, fontSize: 16)),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text('Ainda não tem uma conta?',
                      style: TextStyle(color: BlueprintTheme.textSecondary)),
                  TextButton(
                      onPressed: () => context.go('/register'),
                      child: const Text('Criar conta')),
                ],
              ),
              const SizedBox(height: 12),
              Text('Sua sessão fica salva com segurança neste dispositivo.',
                  textAlign: TextAlign.center,
                  style: terminalLabel(
                      color: BlueprintTheme.textSecondary, fontSize: 10)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildField(String label, TextEditingController controller,
      bool isPassword, IconData icon) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 10, color: BlueprintTheme.textSecondary),
            const SizedBox(width: 6),
            Text(
              label,
              style: terminalLabel(
                  color: BlueprintTheme.textSecondary, fontSize: 8),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: BlueprintTheme.elevated,
            borderRadius: BorderRadius.circular(12),
          ),
          child: TextField(
            controller: controller,
            autofillHints: isPassword
                ? const [AutofillHints.password]
                : const [AutofillHints.username, AutofillHints.email],
            keyboardType: isPassword
                ? TextInputType.visiblePassword
                : TextInputType.emailAddress,
            textInputAction:
                isPassword ? TextInputAction.done : TextInputAction.next,
            onSubmitted: isPassword ? (_) => _handleLogin() : null,
            obscureText: isPassword && _obscurePassword,
            style: const TextStyle(fontSize: 13),
            decoration: InputDecoration(
              border: InputBorder.none,
              isDense: true,
              hintText: isPassword ? 'Digite sua senha' : 'voce@exemplo.com',
              suffixIcon: isPassword
                  ? IconButton(
                      icon: Icon(
                          _obscurePassword
                              ? LucideIcons.eye
                              : LucideIcons.eyeOff,
                          size: 18),
                      onPressed: () =>
                          setState(() => _obscurePassword = !_obscurePassword),
                    )
                  : null,
              contentPadding: const EdgeInsets.symmetric(vertical: 14),
            ),
          ),
        ),
      ],
    );
  }
}
