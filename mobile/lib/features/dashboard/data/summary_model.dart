class FinancialSummary {
  final double totalSpent;
  final double totalReceived;
  final double checkingBalance;
  final double creditBalance;
  final double closedInvoice;
  final double monthInstallments;
  final double currentInvoice;
  final double todaySpent;
  final double weeklySpent;
  final List<dynamic> byCategory;
  final List<dynamic> byDay;
  final List<dynamic> topMerchants;

  FinancialSummary({
    required this.totalSpent,
    required this.totalReceived,
    required this.checkingBalance,
    required this.creditBalance,
    required this.closedInvoice,
    required this.monthInstallments,
    this.currentInvoice = 0.0,
    this.todaySpent = 0.0,
    this.weeklySpent = 0.0,
    required this.byCategory,
    required this.byDay,
    required this.topMerchants,
  });

  factory FinancialSummary.fromJson(Map<String, dynamic> json) {
    return FinancialSummary(
      totalSpent: (json['total_spent'] as num?)?.toDouble() ?? 0.0,
      totalReceived: (json['total_received'] as num?)?.toDouble() ?? 0.0,
      checkingBalance: (json['checking_balance'] as num?)?.toDouble() ?? 0.0,
      creditBalance: (json['credit_balance'] as num?)?.toDouble() ?? 0.0,
      closedInvoice: (json['closed_invoice'] as num?)?.toDouble() ?? 0.0,
      monthInstallments: (json['month_installments'] as num?)?.toDouble() ?? 0.0,
      currentInvoice: (json['current_invoice'] as num?)?.toDouble() ?? 0.0,
      todaySpent: (json['today_spent'] as num?)?.toDouble() ?? 0.0,
      weeklySpent: (json['weekly_spent'] as num?)?.toDouble() ?? 0.0,
      byCategory: json['by_category'] ?? [],
      byDay: json['by_day'] ?? [],
      topMerchants: json['top_merchants'] ?? [],
    );
  }
}
