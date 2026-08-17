import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_pluggy_connect/flutter_pluggy_connect.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';

class PluggyConnectScreen extends ConsumerStatefulWidget {
  final String connectToken;
  const PluggyConnectScreen({super.key, required this.connectToken});

  @override
  ConsumerState<PluggyConnectScreen> createState() =>
      _PluggyConnectScreenState();
}

class _PluggyConnectScreenState extends ConsumerState<PluggyConnectScreen> {
  bool _syncing = false;

  Future<void> _onSuccess(dynamic data) async {
    final item = data is Map ? data['item'] : null;
    final itemId = item is Map ? item['id']?.toString() : null;
    if (itemId == null || _syncing) return;
    setState(() => _syncing = true);
    try {
      await ref
          .read(apiClientProvider)
          .dio
          .post('/accounts/sync', data: {'item_id': itemId});
      ref.invalidate(connectedAccountsProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text('Conta conectada. A sincronização foi iniciada.')));
        context.pop();
      }
    } catch (_) {
      if (mounted)
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text(
                'A conta foi conectada, mas não foi possível iniciar a sincronização.')));
    } finally {
      if (mounted) setState(() => _syncing = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        backgroundColor: BlueprintTheme.background,
        appBar: AppBar(title: const Text('Conectar conta PJ')),
        body: Stack(children: [
          PluggyConnect(
              connectToken: widget.connectToken,
              language: 'pt',
              onSuccess: _onSuccess,
              onError: (_) => ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                      content: Text('Não foi possível concluir a conexão.')))),
          if (_syncing)
            const ColoredBox(
                color: Color(0x99000000),
                child: Center(child: CircularProgressIndicator())),
        ]),
      );
}
