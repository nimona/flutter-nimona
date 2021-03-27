import 'dart:convert';

class InitRequest {
  String configPath;

  InitRequest({
    required this.configPath,
  });

  Map<String, dynamic> toMap() {
    return {
      'configPath': configPath,
    };
  }

  factory InitRequest.fromMap(Map<String, dynamic> map) {
    return InitRequest(
      configPath: map['configPath'],
    );
  }

  String toJson() => json.encode(toMap());

  factory InitRequest.fromJson(String source) => InitRequest.fromMap(json.decode(source));
}
