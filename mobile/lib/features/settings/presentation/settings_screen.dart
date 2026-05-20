import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/features/auth/presentation/auth_provider.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  final _pluggyIdController = TextEditingController();
  final _pluggySecretController = TextEditingController();
  final _groqKeyController = TextEditingController();
  final _geminiKeyController = TextEditingController();
  bool _savingPluggy = false;
  bool _savingLLM = false;

  @override
  void dispose() {
    _pluggyIdController.dispose();
    _pluggySecretController.dispose();
    _groqKeyController.dispose();
    _geminiKeyController.dispose();
    super.dispose();
  }

  Future<void> _savePluggyKeys() async {
    if (_pluggyIdController.text.trim().isEmpty || _pluggySecretController.text.trim().isEmpty) {
      _showSnack('PREENCHA AMBOS OS CAMPOS', isError: true);
      return;
    }
    setState(() => _savingPluggy = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/accounts/keys', data: {
        'client_id': _pluggyIdController.text.trim(),
        'client_secret': _pluggySecretController.text.trim(),
      });
      _showSnack('CREDENCIAIS PLUGGY SALVAS COM SUCESSO');
      _pluggySecretController.clear();
      ref.invalidate(connectedAccountsProvider);
      await ref.read(authProvider.notifier).fetchUser();
    } catch (e) {
      _showSnack('FALHA AO SALVAR CREDENCIAIS', isError: true);
    } finally {
      if (mounted) setState(() => _savingPluggy = false);
    }
  }

  Future<void> _saveLLMKeys() async {
    setState(() => _savingLLM = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/accounts/llm-keys', data: {
        'groq_api_key': _groqKeyController.text.trim(),
        'gemini_api_key': _geminiKeyController.text.trim(),
      });
      _showSnack('CREDENCIAIS IA SALVAS COM SUCESSO');
    } catch (e) {
      _showSnack('FALHA AO SALVAR CREDENCIAIS IA', isError: true);
    } finally {
      if (mounted) setState(() => _savingLLM = false);
    }
  }

  Future<void> _syncAccount(String itemId) async {
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/accounts/sync', data: {'item_id': itemId});
      _showSnack('SINCRONIZAÇÃO INICIADA');
    } catch (e) {
      _showSnack('FALHA NA SINCRONIZAÇÃO', isError: true);
    }
  }

  Future<void> _deleteAccount(String accountId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: BlueprintTheme.surface,
        title: const Text('CONFIRMAR REMOÇÃO', style: TextStyle(fontWeight: FontWeight.w900, fontSize: 16)),
        content: const Text('Deseja realmente desconectar esta fonte de dados?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('CANCELAR')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('REMOVER', style: TextStyle(color: BlueprintTheme.danger)),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.delete('/accounts/$accountId');
      ref.invalidate(connectedAccountsProvider);
      _showSnack('FONTE REMOVIDA COM SUCESSO');
    } catch (e) {
      _showSnack('FALHA AO REMOVER', isError: true);
    }
  }

  void _showSnack(String msg, {bool isError = false}) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
        backgroundColor: isError ? BlueprintTheme.danger : BlueprintTheme.accentPurple,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(authProvider).user;
    final accountsAsync = ref.watch(connectedAccountsProvider);
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      appBar: AppBar(
        title: const Text('CONFIGURAÇÕES'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout_rounded, size: 20),
            onPressed: () async {
              await ref.read(authProvider.notifier).logout();
            },
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // --- Seção Perfil ---
          _buildSectionHeader('PERFIL_DO_OPERADOR'),
          const SizedBox(height: 12),
          _buildInfoCard([
            _buildInfoRow('NOME', user?.name ?? '—'),
            _buildInfoRow('EMAIL', user?.email ?? '—'),
            _buildInfoRow('ID', user?.id.substring(0, 8) ?? '—'),
          ]),

          const SizedBox(height: 32),

          // --- Seção Credenciais Pluggy ---
          _buildSectionHeader('CREDENCIAIS_OPEN_FINANCE'),
          const SizedBox(height: 4),
          Text(
            user?.isPluggyConfigured == true ? '✓ CONFIGURADO' : '⚠ NÃO CONFIGURADO',
            style: TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.w900,
              color: user?.isPluggyConfigured == true ? BlueprintTheme.success : BlueprintTheme.danger,
            ),
          ),
          const SizedBox(height: 12),
          _buildInputField('PLUGGY_CLIENT_ID', _pluggyIdController, hintText: user?.pluggyClientId ?? '00000000-0000-...'),
          const SizedBox(height: 12),
          _buildInputField('PLUGGY_CLIENT_SECRET', _pluggySecretController, isPassword: true),
          const SizedBox(height: 16),
          _buildActionButton(
            label: _savingPluggy ? 'SALVANDO...' : 'SALVAR_CREDENCIAIS_PLUGGY',
            onTap: _savingPluggy ? null : _savePluggyKeys,
            color: BlueprintTheme.accentPurple,
          ),

          const SizedBox(height: 32),

          // --- Seção Credenciais IA ---
          _buildSectionHeader('CREDENCIAIS_IA'),
          const SizedBox(height: 12),
          _buildInputField('GROQ_API_KEY', _groqKeyController, hintText: 'gsk_...', isPassword: true),
          const SizedBox(height: 12),
          _buildInputField('GEMINI_API_KEY', _geminiKeyController, hintText: 'AIzaSy...', isPassword: true),
          const SizedBox(height: 16),
          _buildActionButton(
            label: _savingLLM ? 'SALVANDO...' : 'SALVAR_CHAVES_IA',
            onTap: _savingLLM ? null : _saveLLMKeys,
            color: BlueprintTheme.accentTeal,
          ),

          const SizedBox(height: 32),

          // --- Fontes de Dados Conectadas ---
          _buildSectionHeader('FONTES_DE_DADOS'),
          const SizedBox(height: 12),
          accountsAsync.when(
            data: (accounts) {
              if (accounts.isEmpty) {
                return Container(
                  padding: const EdgeInsets.all(32),
                  decoration: BoxDecoration(
                    color: BlueprintTheme.surface,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: BlueprintTheme.border),
                  ),
                  child: const Column(
                    children: [
                      Icon(Icons.account_balance_outlined, size: 40, color: BlueprintTheme.textSecondary),
                      SizedBox(height: 12),
                      Text('NENHUMA_FONTE_CONECTADA', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: BlueprintTheme.textSecondary)),
                    ],
                  ),
                );
              }
              return Column(
                children: accounts.map((acc) => _buildAccountCard(acc, currencyFormat)).toList(),
              );
            },
            loading: () => const Center(child: Padding(padding: EdgeInsets.all(32), child: CircularProgressIndicator())),
            error: (e, _) => Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger, fontSize: 10)),
          ),

          const SizedBox(height: 48),

          // --- Segurança ---
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: BlueprintTheme.success.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: BlueprintTheme.success.withValues(alpha: 0.3)),
            ),
            child: const Row(
              children: [
                Icon(Icons.shield_rounded, color: BlueprintTheme.success, size: 20),
                SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Dados criptografados via AES-256. Tokens efêmeros da Pluggy API.',
                    style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  Widget _buildAccountCard(Map<String, dynamic> acc, NumberFormat fmt) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Row(
        children: [
          Container(
            width: 44, height: 44,
            decoration: BoxDecoration(
              color: BlueprintTheme.elevated,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: BlueprintTheme.border),
            ),
            child: acc['institution_logo'] != null
              ? ClipRRect(borderRadius: BorderRadius.circular(11), child: Image.network(acc['institution_logo'], fit: BoxFit.contain))
              : Center(child: Text((acc['institution_name'] ?? '?').toString().substring(0, 1).toUpperCase(), style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 18))),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text((acc['institution_name'] ?? 'DESCONHECIDO').toString().toUpperCase(), style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12), overflow: TextOverflow.ellipsis),
                const SizedBox(height: 2),
                Text('${acc['account_type'] ?? 'N/A'} • ${fmt.format((acc['balance'] as num?)?.toDouble() ?? 0)}',
                  style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary)),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(Icons.sync_rounded, size: 18),
            color: BlueprintTheme.accentPurple,
            onPressed: () => _syncAccount(acc['pluggy_item_id'] ?? ''),
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline_rounded, size: 18),
            color: BlueprintTheme.danger,
            onPressed: () => _deleteAccount(acc['id'] ?? ''),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Text(title, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple));
  }

  Widget _buildInfoCard(List<Widget> children) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Column(children: children),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary)),
          Text(value, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w900, fontFamily: 'monospace')),
        ],
      ),
    );
  }

  Widget _buildInputField(String label, TextEditingController controller, {bool isPassword = false, String? hintText}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary, letterSpacing: 1)),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          decoration: BoxDecoration(
            color: BlueprintTheme.elevated,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: BlueprintTheme.border),
          ),
          child: TextField(
            controller: controller,
            obscureText: isPassword,
            style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
            decoration: InputDecoration(
              border: InputBorder.none,
              hintText: hintText,
              hintStyle: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary.withValues(alpha: 0.4)),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildActionButton({required String label, VoidCallback? onTap, required Color color}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        height: 48,
        decoration: BoxDecoration(
          color: onTap == null ? color.withValues(alpha: 0.5) : color,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Center(
          child: Text(label, style: const TextStyle(fontWeight: FontWeight.w900, color: Colors.white, fontSize: 12, letterSpacing: 0.5)),
        ),
      ),
    );
  }
}
