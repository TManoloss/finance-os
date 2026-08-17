import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/data/dashboard_models.dart';
import 'package:finance_os/features/dashboard/data/summary_model.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/daily_spending_chart.dart';
import 'package:finance_os/features/dashboard/presentation/widgets/recent_transactions.dart';

final recentTxProvider = FutureProvider<List<dynamic>>((ref) async {
  final response = await ref.watch(apiClientProvider).dio.get('/transactions?page=1&page_size=4');
  return (response.data['data']['transactions'] ?? []) as List<dynamic>;
});

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summary = ref.watch(summaryProvider);
    final transactions = ref.watch(recentTxProvider);
    final format = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      body: SafeArea(
        child: RefreshIndicator(
          color: BlueprintTheme.accentPurple,
          onRefresh: () async {
            ref.invalidate(summaryProvider);
            ref.invalidate(recentTxProvider);
            await ref.read(summaryProvider.future);
          },
          child: summary.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (_, __) => _ErrorState(onRetry: () => ref.invalidate(summaryProvider)),
            data: (data) => ListView(
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 104),
              children: [
                _Header(onSettings: () => context.push('/settings')),
                const SizedBox(height: 22),
                _BalanceCard(summary: data, format: format),
                const SizedBox(height: 14),
                _QuickActions(
                  onPortfolio: () => context.go('/cards'),
                  onHistory: () => context.go('/transactions'),
                  onPierre: () => context.push('/chat'),
                  onReports: () => context.push('/reports'),
                ),
                const SizedBox(height: 22),
                _SectionTitle(title: 'Visão deste período', action: 'Mês'),
                const SizedBox(height: 10),
                _PeriodPicker(current: ref.watch(periodProvider), onChanged: (value) => ref.read(periodProvider.notifier).state = value),
                const SizedBox(height: 14),
                _Stats(summary: data, format: format),
                const SizedBox(height: 22),
                const _SectionTitle(title: 'Fluxo de caixa'),
                const SizedBox(height: 10),
                _ChartCard(data: data),
                const SizedBox(height: 22),
                _SectionTitle(title: 'Últimas transações', action: 'Ver tudo', onAction: () => context.go('/transactions')),
                const SizedBox(height: 10),
                Container(
                  decoration: neoBrutalCard(),
                  child: transactions.when(
                    data: (items) => RecentTransactions(transactions: items),
                    loading: () => const Padding(padding: EdgeInsets.all(28), child: Center(child: CircularProgressIndicator())),
                    error: (_, __) => const Padding(padding: EdgeInsets.all(20), child: Text('Não foi possível carregar as transações.')),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _Header extends StatelessWidget {
  final VoidCallback onSettings;
  const _Header({required this.onSettings});
  @override
  Widget build(BuildContext context) => Row(
        children: [
          const CircleAvatar(radius: 21, backgroundColor: BlueprintTheme.elevated, child: Icon(LucideIcons.user, size: 19)),
          const SizedBox(width: 11),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('Bom dia', style: terminalLabel(fontSize: 12)),
            Text('FinanceOS', style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w700)),
          ])),
          _RoundButton(icon: LucideIcons.bell, onTap: () {}),
          const SizedBox(width: 8),
          _RoundButton(icon: LucideIcons.settings2, onTap: onSettings),
        ],
      );
}

class _BalanceCard extends StatelessWidget {
  final FinancialSummary summary;
  final NumberFormat format;
  const _BalanceCard({required this.summary, required this.format});
  @override
  Widget build(BuildContext context) {
    final balance = summary.checkingBalance - summary.closedInvoice - summary.monthInstallments;
    return Container(
      height: 184,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        gradient: const LinearGradient(colors: [Color(0xFF3A1230), Color(0xFF902A10), Color(0xFFE1840C)], begin: Alignment.topLeft, end: Alignment.bottomRight),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        const Text('Saldo disponível', style: TextStyle(color: Color(0xDDF4F4F5), fontSize: 13)),
        const SizedBox(height: 6),
        Text(format.format(balance), style: const TextStyle(color: Colors.white, fontSize: 31, fontWeight: FontWeight.w700, letterSpacing: -1.5)),
        const Spacer(),
        Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
          const Text('FINANCEOS •••• 4688', style: TextStyle(color: Color(0xDDF4F4F5), fontSize: 11, letterSpacing: .5)),
          Text('Atualizado hoje', style: terminalLabel(color: const Color(0xDDF4F4F5), fontSize: 10)),
        ]),
      ]),
    );
  }
}

class _QuickActions extends StatelessWidget {
  final VoidCallback onPortfolio, onHistory, onPierre, onReports;
  const _QuickActions({required this.onPortfolio, required this.onHistory, required this.onPierre, required this.onReports});
  @override
  Widget build(BuildContext context) => Row(children: [
        _Action(icon: LucideIcons.creditCard, label: 'Cartões', onTap: onPortfolio),
        _Action(icon: LucideIcons.arrowLeftRight, label: 'Extrato', onTap: onHistory),
        _Action(icon: LucideIcons.messageCircle, label: 'Pierre', onTap: onPierre),
        _Action(icon: LucideIcons.sparkles, label: 'Insights', onTap: onReports),
      ]);
}

class _Action extends StatelessWidget {
  final IconData icon; final String label; final VoidCallback onTap;
  const _Action({required this.icon, required this.label, required this.onTap});
  @override
  Widget build(BuildContext context) => Expanded(child: InkWell(
    onTap: onTap,
    borderRadius: BorderRadius.circular(14),
    child: Column(children: [
      Container(width: 52, height: 52, decoration: const BoxDecoration(color: BlueprintTheme.elevated, shape: BoxShape.circle), child: Icon(icon, size: 19)),
      const SizedBox(height: 7), Text(label, style: terminalLabel(fontSize: 10)),
    ]),
  ));
}

class _PeriodPicker extends StatelessWidget {
  final String current; final ValueChanged<String> onChanged;
  const _PeriodPicker({required this.current, required this.onChanged});
  @override
  Widget build(BuildContext context) {
    const periods = [('month', '1M'), ('quarter', '3M'), ('semester', '6M'), ('year', '1A')];
    return Row(children: periods.map((period) => Expanded(child: Padding(
      padding: const EdgeInsets.only(right: 7),
      child: ChoiceChip(label: Text(period.$2), selected: current == period.$1, onSelected: (_) => onChanged(period.$1)),
    ))).toList());
  }
}

class _Stats extends StatelessWidget {
  final FinancialSummary summary; final NumberFormat format;
  const _Stats({required this.summary, required this.format});
  @override
  Widget build(BuildContext context) => Row(children: [
    _Stat(label: 'Hoje', value: format.format(summary.todaySpent), color: BlueprintTheme.danger),
    const SizedBox(width: 10),
    _Stat(label: 'Esta semana', value: format.format(summary.weeklySpent), color: BlueprintTheme.warning),
  ]);
}

class _Stat extends StatelessWidget {
  final String label, value; final Color color;
  const _Stat({required this.label, required this.value, required this.color});
  @override
  Widget build(BuildContext context) => Expanded(child: Container(
    padding: const EdgeInsets.all(15), decoration: neoBrutalCard(),
    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [Text(label, style: terminalLabel()), const SizedBox(height: 7), Text(value, style: moneyStyle(color: color, fontSize: 18))]),
  ));
}

class _ChartCard extends StatelessWidget {
  final FinancialSummary data; const _ChartCard({required this.data});
  @override
  Widget build(BuildContext context) {
    final byDay = data.byDay.whereType<Map<String, dynamic>>().map(DailyBalance.fromJson).toList();
    return Container(height: 235, padding: const EdgeInsets.fromLTRB(8, 16, 16, 10), decoration: neoBrutalCard(), child: DailySpendingChart(byDay: byDay));
  }
}

class _SectionTitle extends StatelessWidget {
  final String title; final String? action; final VoidCallback? onAction;
  const _SectionTitle({required this.title, this.action, this.onAction});
  @override
  Widget build(BuildContext context) => Row(children: [
    Expanded(child: Text(title, style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700))),
    if (action != null) TextButton(onPressed: onAction, child: Text(action!, style: terminalLabel(color: BlueprintTheme.accentPurple, fontSize: 10))),
  ]);
}

class _RoundButton extends StatelessWidget {
  final IconData icon; final VoidCallback onTap; const _RoundButton({required this.icon, required this.onTap});
  @override
  Widget build(BuildContext context) => InkResponse(onTap: onTap, radius: 23, child: Container(width: 42, height: 42, decoration: const BoxDecoration(color: BlueprintTheme.elevated, shape: BoxShape.circle), child: Icon(icon, size: 18)));
}

class _ErrorState extends StatelessWidget {
  final VoidCallback onRetry; const _ErrorState({required this.onRetry});
  @override
  Widget build(BuildContext context) => Center(child: Padding(padding: const EdgeInsets.all(24), child: Column(mainAxisSize: MainAxisSize.min, children: [const Icon(LucideIcons.wifiOff, size: 36), const SizedBox(height: 12), const Text('Não foi possível atualizar seus dados.'), const SizedBox(height: 14), ElevatedButton(onPressed: onRetry, child: const Text('Tentar novamente'))])));
}
