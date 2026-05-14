class UserModel {
  final String id;
  final String name;
  final String email;
  final String? pluggyClientId;

  UserModel({
    required this.id,
    required this.name,
    required this.email,
    this.pluggyClientId,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'],
      name: json['name'],
      email: json['email'],
      pluggyClientId: json['pluggy_client_id'],
    );
  }

  bool get isPluggyConfigured => pluggyClientId != null && pluggyClientId!.isNotEmpty;
}
