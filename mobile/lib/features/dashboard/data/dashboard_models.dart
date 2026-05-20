/// Modelos tipados para dados do dashboard

class CategorySummary {
  final String categoryId;
  final String categoryName;
  final String? color;
  final double total;
  final double percentage;
  final int transactionCount;

  CategorySummary({
    required this.categoryId,
    required this.categoryName,
    this.color,
    required this.total,
    required this.percentage,
    required this.transactionCount,
  });

  factory CategorySummary.fromJson(Map<String, dynamic> json) {
    return CategorySummary(
      categoryId: (json['category_id'] ?? '').toString(),
      categoryName: (json['category_name'] ?? json['name'] ?? '').toString(),
      color: json['color']?.toString(),
      total: (json['total'] as num?)?.toDouble() ?? 0.0,
      percentage: (json['percentage'] as num?)?.toDouble() ?? 0.0,
      transactionCount: (json['transaction_count'] as num?)?.toInt() ?? 0,
    );
  }
}

class DailyBalance {
  final DateTime date;
  final double totalSpent;
  final double totalReceived;

  DailyBalance({
    required this.date,
    required this.totalSpent,
    required this.totalReceived,
  });

  factory DailyBalance.fromJson(Map<String, dynamic> json) {
    return DailyBalance(
      date: DateTime.tryParse(json['date']?.toString() ?? '') ?? DateTime.now(),
      totalSpent: (json['total_spent'] as num?)?.toDouble() ?? 0.0,
      totalReceived: (json['total_received'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class MerchantSummary {
  final String merchantName;
  final double total;
  final int count;

  MerchantSummary({
    required this.merchantName,
    required this.total,
    required this.count,
  });

  factory MerchantSummary.fromJson(Map<String, dynamic> json) {
    return MerchantSummary(
      merchantName: (json['merchant_name'] ?? json['merchant'] ?? '').toString(),
      total: (json['total'] as num?)?.toDouble() ?? 0.0,
      count: (json['count'] as num?)?.toInt() ?? 0,
    );
  }
}
