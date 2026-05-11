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
