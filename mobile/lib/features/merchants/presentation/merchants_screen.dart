import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/settings/presentation/settings_provider.dart';
import 'package:finance_os/shared/widgets/premium_page.dart';

class MerchantsScreen extends ConsumerWidget {
  const MerchantsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final data = ref.watch(merchantsProvider);
    final money = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    return PremiumPage(title: 'Estabelecimentos', children: [
      const PremiumTitle(title: 'Onde seu dinheiro vai', subtitle: 'Seus destinos de gastos mais frequentes.'),
      data.when(
        loading: () => const Padding(padding: EdgeInsets.all(36), child: Center(child: CircularProgressIndicator())),
        error: (_, __) => const PremiumCard(child: Text('Não foi possível carregar os estabelecimentos.')),
        data: (items) => items.isEmpty
          ? const PremiumCard(child: Center(child: Padding(padding: EdgeInsets.all(24), child: Text('Ainda não há gastos registrados.'))))
          : Column(children: items.asMap().entries.map((entry) {
              final item = entry.value;
              final name = (item['merchant'] ?? item['merchant_name'] ?? 'Estabelecimento').toString();
              final total = (item['total'] as num?)?.toDouble() ?? 0;
              final count = (item['count'] as num?)?.toInt() ?? 0;
              return Padding(padding: const EdgeInsets.only(bottom: 10), child: PremiumCard(child: Row(children: [
                CircleAvatar(radius: 20, backgroundColor: const Color(0xFF312743), child: Text('${entry.key + 1}', style: const TextStyle(color: Color(0xFFFFA58E), fontWeight: FontWeight.w700))),
                const SizedBox(width: 12), Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [Text(name, style: const TextStyle(fontWeight: FontWeight.w700)), const SizedBox(height: 3), Text('$count compras', style: terminalLabel(fontSize: 11))])),
                Text(money.format(total), style: moneyStyle(color: BlueprintTheme.danger, fontSize: 15)),
              ])));
            }).toList()),
      ),
    ]);
  }
}
