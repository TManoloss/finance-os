# Mobile Dashboard V2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the new dashboard V2 for the mobile app, including daily evolution chart, merchant ranking, and hybrid assets/liabilities view.

**Architecture:** Update the backend repository to return separate balances by account type, update the mobile model, and create modular widgets for each new dashboard section.

**Tech Stack:** Go (Echo + pgx), Flutter (Riverpod + fl_chart).

---

### Task 1: Update Backend Summary Model

**Files:**
- Modify: `backend/internal/repository/transaction_repository.go`

- [ ] **Step 1: Update TransactionSummary struct**

```go
type TransactionSummary struct {
	TotalSpent      float64           `json:"total_spent"`
	TotalReceived   float64           `json:"total_received"`
	CheckingBalance float64           `json:"checking_balance"`
	CreditBalance   float64           `json:"credit_balance"`
	ByCategory      []CategorySummary `json:"by_category"`
	ByDay           []DaySummary      `json:"by_day"`
	TopMerchants    []MerchantSummary `json:"top_merchants"`
}
```

- [ ] **Step 2: Update GetSummary implementation**
Update the `GetSummary` method to calculate `CheckingBalance` (checking/savings) and `CreditBalance` (credit).

```go
	// 5. Saldos por Tipo de Conta
	balanceQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN account_type IN ('checking', 'savings') THEN balance ELSE 0 END), 0) as checking,
			COALESCE(SUM(CASE WHEN account_type = 'credit' THEN balance ELSE 0 END), 0) as credit
		FROM connected_accounts
		WHERE user_id = $1
	`
	err = r.db.QueryRow(ctx, balanceQuery, userID).Scan(&summary.CheckingBalance, &summary.CreditBalance)
	if err != nil {
		return nil, err
	}
```

- [ ] **Step 3: Recompile and Verify**
Run: `cd backend && go build -o server cmd/server/main.go`
Verify that `curl http://localhost:8080/api/v1/transactions/summary` (with valid token) now returns the new fields.

- [ ] **Step 4: Commit**
`git add backend/internal/repository/transaction_repository.go && git commit -m "feat(backend): add balances by type to summary"`

---

### Task 2: Update Mobile Summary Model

**Files:**
- Modify: `mobile/lib/features/dashboard/data/summary_model.dart`

- [ ] **Step 1: Update FinancialSummary class**

```dart
class FinancialSummary {
  final double totalSpent;
  final double totalReceived;
  final double checkingBalance;
  final double creditBalance;
  final List<dynamic> byCategory;
  final List<dynamic> byDay;
  final List<dynamic> topMerchants;

  FinancialSummary({
    required this.totalSpent,
    required this.totalReceived,
    required this.checkingBalance,
    required this.creditBalance,
    required this.byCategory,
    required this.byDay,
    required this.topMerchants,
  });

  factory FinancialSummary.fromJson(Map<String, dynamic> json) {
    return FinancialSummary(
      totalSpent: (json['total_spent'] as num).toDouble(),
      totalReceived: (json['total_received'] as num).toDouble(),
      checkingBalance: (json['checking_balance'] as num).toDouble(),
      creditBalance: (json['credit_balance'] as num).toDouble(),
      byCategory: json['by_category'] ?? [],
      byDay: json['by_day'] ?? [],
      topMerchants: json['top_merchants'] ?? [],
    );
  }
}
```

- [ ] **Step 2: Commit**
`git add mobile/lib/features/dashboard/data/summary_model.dart && git commit -m "feat(mobile): update summary model with new fields"`

---

### Task 3: Implement CommitmentBar Widget

**Files:**
- Create: `mobile/lib/features/dashboard/presentation/widgets/commitment_bar.dart`

- [ ] **Step 1: Create the widget**
Implement a bar that shows `creditBalance / checkingBalance` percentage.

```dart
import 'package:flutter/material.dart';

class CommitmentBar extends StatelessWidget {
  final double checking;
  final double credit;

  const CommitmentBar({super.key, required this.checking, required this.credit});

  @override
  Widget build(BuildContext context) {
    final double percentage = (checking > 0) ? (credit.abs() / checking).clamp(0.0, 1.0) : 1.0;
    final bool isWarning = percentage > 0.7;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text('INDICE_COMPROMETIMENTO', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
            Text('${(percentage * 100).toStringAsFixed(0)}%', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: isWarning ? Colors.red : Colors.black)),
          ],
        ),
        const SizedBox(height: 4),
        Container(
          height: 12,
          decoration: BoxDecoration(border: Border.all(color: Colors.black)),
          child: FractionallySizedBox(
            alignment: Alignment.centerLeft,
            widthFactor: percentage,
            child: Container(color: isWarning ? Colors.red : Colors.black),
          ),
        ),
      ],
    );
  }
}
```

- [ ] **Step 2: Commit**
`git add mobile/lib/features/dashboard/presentation/widgets/commitment_bar.dart && git commit -m "feat(mobile): add commitment bar widget"`

---

### Task 4: Implement MerchantRankingWidget

**Files:**
- Create: `mobile/lib/features/dashboard/presentation/widgets/merchant_ranking_widget.dart`

- [ ] **Step 1: Create the widget**

```dart
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class MerchantRankingWidget extends StatelessWidget {
  final List<dynamic> merchants;

  const MerchantRankingWidget({super.key, required this.merchants});

  @override
  Widget build(BuildContext context) {
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (var i = 0; i < merchants.length; i++)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 4.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('${i + 1}. ${merchants[i]['merchant_name']}', style: const TextStyle(fontSize: 12)),
                Text(currencyFormat.format(merchants[i]['total']), style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
      ],
    );
  }
}
```

- [ ] **Step 2: Commit**
`git add mobile/lib/features/dashboard/presentation/widgets/merchant_ranking_widget.dart && git commit -m "feat(mobile): add merchant ranking widget"`

---

### Task 5: Implement DailySpendingChart Widget

**Files:**
- Create: `mobile/lib/features/dashboard/presentation/widgets/daily_spending_chart.dart`

- [ ] **Step 1: Create the widget using fl_chart**

```dart
import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';

class DailySpendingChart extends StatelessWidget {
  final List<dynamic> byDay;

  const DailySpendingChart({super.key, required this.byDay});

  @override
  Widget build(BuildContext context) {
    return LineChart(
      LineChartData(
        gridData: const FlGridData(show: false),
        titlesData: const FlTitlesData(show: false),
        borderData: FlBorderData(show: false),
        lineBarsData: [
          LineChartBarData(
            spots: byDay.asMap().entries.map((e) {
              return FlSpot(e.key.toDouble(), (e.value['spent'] as num).toDouble());
            }).toList(),
            isCurved: false,
            color: Colors.black,
            barWidth: 2,
            dotData: const FlDotData(show: false),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Commit**
`git add mobile/lib/features/dashboard/presentation/widgets/daily_spending_chart.dart && git commit -m "feat(mobile): add daily spending chart widget"`

---

### Task 6: Assemble New Dashboard

**Files:**
- Modify: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`

- [ ] **Step 1: Update layout**
Include `DailySpendingChart`, `MerchantRankingWidget`, `CommitmentBar` and the refresh button.

```dart
// ... imports
import 'widgets/daily_spending_chart.dart';
import 'widgets/merchant_ranking_widget.dart';
import 'widgets/commitment_bar.dart';

// Na AppBar:
actions: [
  IconButton(
    icon: const Icon(Icons.refresh),
    onPressed: () => ref.refresh(summaryProvider),
  ),
],

// No body:
// 1. Estado Líquido
BlueprintCard(
  label: 'ESTADO_LIQUIDO_ATUAL',
  child: Text(
    currencyFormat.format(summary.checkingBalance - summary.creditBalance),
    style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
  ),
),
const SizedBox(height: 16),
// 2. Detalhamento Ativos vs Passivos
Row(
  children: [
    Expanded(
      child: BlueprintCard(
        label: 'EM_CONTA',
        child: Text(currencyFormat.format(summary.checkingBalance)),
      ),
    ),
    const SizedBox(width: 16),
    Expanded(
      child: BlueprintCard(
        label: 'EM_CARTÕES',
        child: Text(currencyFormat.format(summary.creditBalance)),
      ),
    ),
  ],
),
const SizedBox(height: 16),
BlueprintCard(
  child: CommitmentBar(checking: summary.checkingBalance, credit: summary.creditBalance),
),
const SizedBox(height: 24),
// 3. Telemetria Diária
BlueprintCard(
  label: 'TELEMETRIA_DIARIA',
  height: 150,
  child: DailySpendingChart(byDay: summary.byDay),
),
const SizedBox(height: 24),
// 4. Ranking
BlueprintCard(
  label: 'RANKING_MERCANTES',
  child: MerchantRankingWidget(merchants: summary.topMerchants),
),
```

- [ ] **Step 2: Commit**
`git add mobile/lib/features/dashboard/presentation/dashboard_screen.dart && git commit -m "feat(mobile): finalize dashboard V2 assembly"`
