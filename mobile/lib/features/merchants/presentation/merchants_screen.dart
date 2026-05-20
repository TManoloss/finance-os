import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';

class MerchantsScreen extends ConsumerWidget {
  const MerchantsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final merchantsAsync = ref.watch(merchantsProvider);
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      appBar: AppBar(title: const Text('MERCHANTS')),
      body: merchantsAsync.when(
        data: (merchants) {
          if (merchants.isEmpty) {
            return const Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.store_rounded, size: 64, color: BlueprintTheme.textSecondary),
                  SizedBox(height: 16),
                  Text('NENHUM_MERCHANT_ENCONTRADO', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: BlueprintTheme.textSecondary)),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () async => ref.refresh(merchantsProvider),
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: merchants.length,
              itemBuilder: (context, index) {
                final m = merchants[index];
                final name = (m['merchant'] ?? m['merchant_name'] ?? '?').toString();
                final total = (m['total'] as num?)?.toDouble() ?? 0;
                final count = (m['count'] as num?)?.toInt() ?? 0;
                final avgTicket = count > 0 ? total / count : 0.0;
                final firstChar = name.isNotEmpty ? name[0].toUpperCase() : '?';

                // Cor baseada no índice (alterna entre púrpura, teal, warning)
                final colors = [BlueprintTheme.accentPurple, BlueprintTheme.accentTeal, BlueprintTheme.warning, BlueprintTheme.danger, BlueprintTheme.success];
                final color = colors[index % colors.length];

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
                      // Avatar
                      Container(
                        width: 48, height: 48,
                        decoration: BoxDecoration(
                          color: color.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Center(
                          child: Text(firstChar, style: TextStyle(fontWeight: FontWeight.w900, fontSize: 20, color: color)),
                        ),
                      ),
                      const SizedBox(width: 16),
                      // Info
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(name.toUpperCase(), style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 13), overflow: TextOverflow.ellipsis),
                            const SizedBox(height: 4),
                            Row(
                              children: [
                                _buildMiniTag('$count OPERAÇÕES', BlueprintTheme.textSecondary),
                                const SizedBox(width: 8),
                                _buildMiniTag('TICKET: ${fmt.format(avgTicket)}', BlueprintTheme.accentTeal),
                              ],
                            ),
                          ],
                        ),
                      ),
                      // Total
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          const Text('TOTAL', style: TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary)),
                          Text(fmt.format(total), style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 15, fontFamily: 'monospace', color: BlueprintTheme.danger)),
                        ],
                      ),
                    ],
                  ),
                );
              },
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('ERRO: $e', style: const TextStyle(color: BlueprintTheme.danger))),
      ),
    );
  }

  Widget _buildMiniTag(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(label, style: TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: color)),
    );
  }
}
