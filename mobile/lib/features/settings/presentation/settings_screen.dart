import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';
import 'package:finance_os/shared/widgets/premium_page.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/core/api/api_client.dart';

final settingsApiProvider =
    Provider((ref) => _SettingsApi(ref.read(apiClientProvider)));

class _SettingsApi {
  final ApiClient api;
  _SettingsApi(this.api);
  Future<void> deleteAccount(String id) => api.dio.delete('/accounts/$id');
}

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final accounts = ref.watch(connectedAccountsProvider);
    return PremiumPage(title: 'Configurações', children: [
      const PremiumTitle(
          title: 'Seu FinanceOS',
          subtitle: 'Tudo é controlado manualmente por você.'),
      const Text('Contas e cartões',
          style: TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
      const SizedBox(height: 10),
      accounts.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => const PremiumCard(
            child: Text('Não foi possível carregar suas contas.')),
        data: (items) => items.isEmpty
            ? const PremiumCard(
                child: Text('Crie uma conta ou cartão na aba Patrimônio.'))
            : Column(
                children: items
                    .map((account) => Padding(
                        padding: const EdgeInsets.only(bottom: 10),
                        child: PremiumCard(
                            child: Row(children: [
                          const CircleAvatar(
                              backgroundColor: BlueprintTheme.elevated,
                              child: Icon(LucideIcons.walletCards, size: 18)),
                          const SizedBox(width: 12),
                          Expanded(
                              child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                Text(
                                    (account['institution_name'] ?? 'Conta')
                                        .toString(),
                                    style: const TextStyle(
                                        fontWeight: FontWeight.w700)),
                                Text((account['account_type'] ?? '').toString(),
                                    style: terminalLabel(fontSize: 10))
                              ])),
                          IconButton(
                              icon: const Icon(LucideIcons.trash2,
                                  color: BlueprintTheme.danger),
                              onPressed: () async {
                                await ref
                                    .read(settingsApiProvider)
                                    .deleteAccount(account['id'].toString());
                                ref.invalidate(connectedAccountsProvider);
                              })
                        ]))))
                    .toList()),
      ),
      const SizedBox(height: 20),
      const Text('Inteligência artificial',
          style: TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
      const SizedBox(height: 10),
      PremiumCard(
          child: ListTile(
              leading: const Icon(LucideIcons.sparkles),
              title: const Text('Credenciais de IA'),
              subtitle: Text(
                  'Configure Groq ou Gemini para conversar com Pierre.',
                  style: terminalLabel(fontSize: 11)),
              trailing: const Icon(LucideIcons.chevronRight))),
      const SizedBox(height: 20),
      const Text('Open Finance',
          style: TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
      const SizedBox(height: 10),
      PremiumCard(
          child: ListTile(
              onTap: () => _connectPluggy(context, ref),
              leading: const Icon(LucideIcons.landmark),
              title: const Text('Conectar conta PJ'),
              subtitle: Text('Use a conexão segura da Pluggy.',
                  style: terminalLabel(fontSize: 11)),
              trailing: const Icon(LucideIcons.chevronRight))),
      const SizedBox(height: 20),
      PremiumCard(
          child: ListTile(
              leading: const Icon(LucideIcons.info),
              title: const Text('Modo manual'),
              subtitle: Text('Sem integração Open Finance.',
                  style: terminalLabel(fontSize: 11)))),
    ]);
  }

  Future<void> _connectPluggy(BuildContext context, WidgetRef ref) async {
    try {
      final response =
          await ref.read(apiClientProvider).dio.post('/accounts/connect-token');
      final token = response.data['data']['accessToken']?.toString();
      if (token == null || !context.mounted) throw StateError('token ausente');
      context.push('/pluggy-connect', extra: token);
    } catch (_) {
      if (context.mounted)
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text(
                'Não foi possível iniciar a Pluggy. Verifique suas credenciais.')));
    }
  }
}
