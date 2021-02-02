import 'dart:convert';

class SubscribeRequest {
  List<String> lookups;

  SubscribeRequest({
    this.lookups,
  });

  SubscribeRequest copyWith({
    String lookups,
    String orderBy,
    String orderDir,
    int limit,
    int offset,
  }) {
    return SubscribeRequest(
      lookups: lookups ?? this.lookups,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'lookups': lookups,
    };
  }

  factory SubscribeRequest.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return SubscribeRequest(
      lookups: map['lookups'],
    );
  }

  String toJson() => json.encode(toMap());

  factory SubscribeRequest.fromJson(String source) =>
      SubscribeRequest.fromMap(json.decode(source));

  @override
  String toString() {
    return 'SubscribeRequest(lookups: $lookups)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is SubscribeRequest &&
        o.lookups == lookups;
  }

  @override
  int get hashCode {
    return lookups.hashCode;
  }
}
