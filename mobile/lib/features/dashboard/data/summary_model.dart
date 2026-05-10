class FinancialSummary {
  final double totalSpent;
  final double totalReceived;
  final List<dynamic> byCategory;

  FinancialSummary({
    required this.totalSpent,
    required this.totalReceived,
    required this.byCategory,
  });

  factory FinancialSummary.fromJson(Map<String, dynamic> json) {
    return FinancialSummary(
      totalSpent: (json['total_spent'] as num).toDouble(),
      totalReceived: (json['total_received'] as num).toDouble(),
      byCategory: json['by_category'] ?? [],
    );
  }
}
