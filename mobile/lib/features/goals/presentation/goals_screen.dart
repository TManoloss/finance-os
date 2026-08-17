import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

final goalsProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final response = await ref.watch(apiClientProvider).dio.get('/goals');
  return ((response.data['data'] ?? []) as List)
      .whereType<Map<String, dynamic>>()
      .toList();
});

class GoalsScreen extends ConsumerWidget {
  const GoalsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final goals = ref.watch(goalsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Metas')),
      floatingActionButton: Padding(
        padding: const EdgeInsets.only(bottom: 72),
        child: FloatingActionButton(
          onPressed: () => _createGoal(context, ref),
          backgroundColor: const Color(0xFFFFA58E),
          foregroundColor: Colors.black,
          child: const Icon(LucideIcons.plus),
        ),
      ),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(goalsProvider),
        child: goals.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (_, __) => const _GoalsEmpty(error: true),
          data: (items) => items.isEmpty
              ? const _GoalsEmpty()
              : ListView.separated(
                  padding: const EdgeInsets.fromLTRB(20, 12, 20, 104),
                  itemCount: items.length + 1,
                  separatorBuilder: (_, __) => const SizedBox(height: 12),
                  itemBuilder: (_, index) => index == 0
                      ? const _GoalsHeader()
                      : _GoalCard(goal: items[index - 1]),
                ),
        ),
      ),
    );
  }

  Future<void> _createGoal(BuildContext context, WidgetRef ref) async {
    final name = TextEditingController();
    final amount = TextEditingController();
    var type = 'savings';
    await showModalBottomSheet<void>(
      context: context,
      useRootNavigator: true,
      isScrollControlled: true,
      backgroundColor: BlueprintTheme.elevated,
      shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (sheetContext) => Padding(
        padding: EdgeInsets.fromLTRB(
            20, 24, 20, MediaQuery.viewInsetsOf(sheetContext).bottom + 24),
        child: StatefulBuilder(
            builder: (context, setState) => Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Nova meta',
                          style: Theme.of(context)
                              .textTheme
                              .titleLarge
                              ?.copyWith(fontWeight: FontWeight.w700)),
                      const SizedBox(height: 18),
                      TextField(
                          controller: name,
                          textCapitalization: TextCapitalization.sentences,
                          decoration: const InputDecoration(
                              hintText: 'Ex.: Reserva de emergência')),
                      const SizedBox(height: 10),
                      TextField(
                          controller: amount,
                          keyboardType: const TextInputType.numberWithOptions(
                              decimal: true),
                          decoration: const InputDecoration(
                              hintText: 'Valor alvo em R\$')),
                      const SizedBox(height: 10),
                      DropdownButtonFormField<String>(
                        value: type,
                        dropdownColor: BlueprintTheme.elevated,
                        decoration: const InputDecoration(),
                        items: const [
                          DropdownMenuItem(
                              value: 'savings',
                              child: Text('Guardar dinheiro')),
                          DropdownMenuItem(
                              value: 'debt_payoff',
                              child: Text('Quitar dívida')),
                          DropdownMenuItem(
                              value: 'spending_limit',
                              child: Text('Limitar gastos')),
                          DropdownMenuItem(
                              value: 'income_target',
                              child: Text('Meta de renda')),
                        ],
                        onChanged: (value) =>
                            setState(() => type = value ?? type),
                      ),
                      const SizedBox(height: 18),
                      ElevatedButton(
                        onPressed: () async {
                          final target =
                              double.tryParse(amount.text.replaceAll(',', '.'));
                          if (name.text.trim().isEmpty ||
                              target == null ||
                              target <= 0) return;
                          await ref.read(apiClientProvider).dio.post('/goals',
                              data: {
                                'name': name.text.trim(),
                                'goal_type': type,
                                'target_amount': target
                              });
                          ref.invalidate(goalsProvider);
                          if (context.mounted) Navigator.pop(context);
                        },
                        child: const Text('Criar meta'),
                      ),
                    ])),
      ),
    );
  }
}

class _GoalsHeader extends StatelessWidget {
  const _GoalsHeader();
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 4),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text('Seus objetivos',
              style: Theme.of(context)
                  .textTheme
                  .headlineSmall
                  ?.copyWith(fontWeight: FontWeight.w700)),
          const SizedBox(height: 4),
          Text('Acompanhe o que importa para seu dinheiro.',
              style: terminalLabel(fontSize: 12)),
        ]),
      );
}

class _GoalCard extends StatelessWidget {
  final Map<String, dynamic> goal;
  const _GoalCard({required this.goal});
  @override
  Widget build(BuildContext context) {
    final target = (goal['target_amount'] as num?)?.toDouble() ?? 0;
    final current = (goal['current_amount'] as num?)?.toDouble() ?? 0;
    final progress =
        target == 0 ? 0.0 : (current / target).clamp(0, 1).toDouble();
    final format = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    return Container(
      padding: const EdgeInsets.all(17),
      decoration: neoBrutalCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          const CircleAvatar(
              radius: 18,
              backgroundColor: Color(0xFF312743),
              child:
                  Icon(LucideIcons.flame, color: Color(0xFFFFA58E), size: 17)),
          const SizedBox(width: 11),
          Expanded(
              child: Text((goal['name'] ?? 'Meta').toString(),
                  style: const TextStyle(
                      fontWeight: FontWeight.w700, fontSize: 16))),
          Text('${(progress * 100).round()}%',
              style:
                  moneyStyle(color: BlueprintTheme.accentPurple, fontSize: 15))
        ]),
        const SizedBox(height: 16),
        ClipRRect(
            borderRadius: BorderRadius.circular(6),
            child: LinearProgressIndicator(
                value: progress,
                minHeight: 8,
                color: const Color(0xFFFFA58E),
                backgroundColor: BlueprintTheme.border)),
        const SizedBox(height: 9),
        Text('${format.format(current)} de ${format.format(target)}',
            style: terminalLabel(fontSize: 11)),
      ]),
    );
  }
}

class _GoalsEmpty extends StatelessWidget {
  final bool error;
  const _GoalsEmpty({this.error = false});
  @override
  Widget build(BuildContext context) => ListView(children: [
        Padding(
            padding: const EdgeInsets.all(36),
            child: Column(children: [
              Icon(error ? LucideIcons.wifiOff : LucideIcons.target,
                  size: 42, color: BlueprintTheme.textSecondary),
              const SizedBox(height: 15),
              Text(
                  error
                      ? 'Não foi possível carregar suas metas.'
                      : 'Crie sua primeira meta financeira.',
                  textAlign: TextAlign.center,
                  style: terminalLabel(fontSize: 13))
            ]))
      ]);
}
