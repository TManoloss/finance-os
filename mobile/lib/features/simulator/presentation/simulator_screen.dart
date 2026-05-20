import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

class SimulatorScreen extends ConsumerStatefulWidget {
  const SimulatorScreen({super.key});

  @override
  ConsumerState<SimulatorScreen> createState() => _SimulatorScreenState();
}

class _SimulatorScreenState extends ConsumerState<SimulatorScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final _amountController = TextEditingController();
  final _installmentsController = TextEditingController(text: '1');
  final _cutAmountController = TextEditingController();
  final _cutNameController = TextEditingController();
  bool _loading = false;
  Map<String, dynamic>? _result;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    _amountController.dispose();
    _installmentsController.dispose();
    _cutAmountController.dispose();
    _cutNameController.dispose();
    super.dispose();
  }

  Future<void> _simulatePurchase() async {
    final amount = double.tryParse(_amountController.text) ?? 0;
    final installments = int.tryParse(_installmentsController.text) ?? 1;
    if (amount <= 0) return;

    setState(() { _loading = true; _result = null; });
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.post('/simulator/purchase', data: {
        'amount': amount,
        'installments': installments,
      });
      setState(() => _result = resp.data['data'] as Map<String, dynamic>?);
    } catch (e) {
      _showError('FALHA_NA_SIMULAÇÃO');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _simulateCut() async {
    final amount = double.tryParse(_cutAmountController.text) ?? 0;
    if (amount <= 0) return;

    setState(() { _loading = true; _result = null; });
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.post('/simulator/cut', data: {
        'amount': amount,
        'name': _cutNameController.text.trim(),
      });
      setState(() => _result = resp.data['data'] as Map<String, dynamic>?);
    } catch (e) {
      _showError('FALHA_NA_SIMULAÇÃO');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _showError(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), backgroundColor: BlueprintTheme.danger),
    );
  }

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      appBar: AppBar(
        title: const Text('SIMULADOR'),
        bottom: TabBar(
          controller: _tabController,
          labelStyle: const TextStyle(fontWeight: FontWeight.w900, fontSize: 10),
          indicatorColor: BlueprintTheme.accentPurple,
          labelColor: BlueprintTheme.accentPurple,
          unselectedLabelColor: BlueprintTheme.textSecondary,
          tabs: const [
            Tab(text: 'SIMULAR_COMPRA'),
            Tab(text: 'SIMULAR_CORTE'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildPurchaseTab(fmt),
          _buildCutTab(fmt),
        ],
      ),
    );
  }

  Widget _buildPurchaseTab(NumberFormat fmt) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('QUANTO CUSTA ESSA COMPRA?', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w900)),
          const SizedBox(height: 4),
          const Text('Simule o impacto real no seu orçamento', style: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary)),
          const SizedBox(height: 24),
          _buildInput('VALOR_DA_COMPRA (R\$)', _amountController, keyboardType: TextInputType.number),
          const SizedBox(height: 16),
          _buildInput('PARCELAS', _installmentsController, keyboardType: TextInputType.number),
          const SizedBox(height: 24),
          _buildActionButton('SIMULAR_IMPACTO', _loading ? null : _simulatePurchase),
          if (_result != null) ...[
            const SizedBox(height: 24),
            _buildResultCard(fmt),
          ],
        ],
      ),
    );
  }

  Widget _buildCutTab(NumberFormat fmt) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('E SE VOCÊ CORTASSE ESSE GASTO?', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w900)),
          const SizedBox(height: 4),
          const Text('Veja quanto você economizaria ao longo do tempo', style: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary)),
          const SizedBox(height: 24),
          _buildInput('NOME_DO_GASTO', _cutNameController),
          const SizedBox(height: 16),
          _buildInput('VALOR_MENSAL (R\$)', _cutAmountController, keyboardType: TextInputType.number),
          const SizedBox(height: 24),
          _buildActionButton('SIMULAR_ECONOMIA', _loading ? null : _simulateCut),
          if (_result != null) ...[
            const SizedBox(height: 24),
            _buildResultCard(fmt),
          ],
        ],
      ),
    );
  }

  Widget _buildResultCard(NumberFormat fmt) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.accentPurple.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('RESULTADO_DA_SIMULAÇÃO', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
          const SizedBox(height: 16),
          if (_result!['message'] != null)
            Text(_result!['message'].toString(), style: const TextStyle(fontSize: 14, height: 1.5)),
          if (_result!['monthly_impact'] != null) ...[
            const SizedBox(height: 12),
            _buildResultRow('IMPACTO_MENSAL', fmt.format(_result!['monthly_impact'])),
          ],
          if (_result!['yearly_savings'] != null)
            _buildResultRow('ECONOMIA_ANUAL', fmt.format(_result!['yearly_savings'])),
          if (_result!['commitment_after'] != null)
            _buildResultRow('COMPROMETIMENTO', '${(_result!['commitment_after'] as num).toStringAsFixed(1)}%'),
          if (_result!['verdict'] != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: BlueprintTheme.accentTeal.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(_result!['verdict'].toString(), style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: BlueprintTheme.accentTeal)),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildResultRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: BlueprintTheme.textSecondary)),
          Text(value, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w900, fontFamily: 'monospace')),
        ],
      ),
    );
  }

  Widget _buildInput(String label, TextEditingController controller, {TextInputType? keyboardType}) {
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
            keyboardType: keyboardType,
            style: const TextStyle(fontSize: 14),
            decoration: const InputDecoration(border: InputBorder.none),
          ),
        ),
      ],
    );
  }

  Widget _buildActionButton(String label, VoidCallback? onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: double.infinity,
        height: 52,
        decoration: BoxDecoration(
          color: onTap == null ? BlueprintTheme.accentPurple.withValues(alpha: 0.5) : BlueprintTheme.accentPurple,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Center(
          child: _loading
            ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
            : Text(label, style: const TextStyle(fontWeight: FontWeight.w900, color: Colors.white, fontSize: 12, letterSpacing: 0.5)),
        ),
      ),
    );
  }
}
