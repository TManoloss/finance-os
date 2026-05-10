class TransactionModel {
  final String id;
  final String description;
  final double amount;
  final String direction;
  final String date;
  final String? categoryName;

  TransactionModel({
    required this.id,
    required this.description,
    required this.amount,
    required this.direction,
    required this.date,
    this.categoryName,
  });

  factory TransactionModel.fromJson(Map<String, dynamic> json) {
    return TransactionModel(
      id: json['id'] ?? '',
      description: json['description'] ?? '',
      amount: (json['amount'] as num).toDouble(),
      direction: json['direction'] ?? 'debit',
      date: json['date'] ?? '',
      categoryName: json['category'] != null ? json['category']['name'] : null,
    );
  }
}
