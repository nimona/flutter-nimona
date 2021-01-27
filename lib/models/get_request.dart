import 'dart:convert';

import 'package:flutter/foundation.dart';

class GetRequest {
  String lookup;
  String orderBy;
  String orderDir;
  int limit;
  int offset;

  GetRequest({
    this.lookup,
    this.orderBy,
    this.orderDir,
    this.limit,
    this.offset,
  });

  GetRequest copyWith({
    String lookup,
    String orderBy,
    String orderDir,
    int limit,
    int offset,
  }) {
    return GetRequest(
      lookup: lookup ?? this.lookup,
      orderBy: orderBy ?? this.orderBy,
      orderDir: orderDir ?? this.orderDir,
      limit: limit ?? this.limit,
      offset: offset ?? this.offset,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'lookup': lookup,
      'orderBy': orderBy,
      'orderDir': orderDir,
      'limit': limit,
      'offset': offset,
    };
  }

  factory GetRequest.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;

    return GetRequest(
      lookup: map['lookup'],
      orderBy: map['orderBy'],
      orderDir: map['orderDir'],
      limit: map['limit'],
      offset: map['offset'],
    );
  }

  String toJson() => json.encode(toMap());

  factory GetRequest.fromJson(String source) =>
      GetRequest.fromMap(json.decode(source));

  @override
  String toString() {
    return 'GetRequest(lookup: $lookup, orderBy: $orderBy, orderDir: $orderDir, limit: $limit, offset: $offset)';
  }

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;

    return o is GetRequest &&
        o.lookup == lookup &&
        o.orderBy == orderBy &&
        o.orderDir == orderDir &&
        o.limit == limit &&
        o.offset == offset;
  }

  @override
  int get hashCode {
    return lookup.hashCode ^
        orderBy.hashCode ^
        orderDir.hashCode ^
        limit.hashCode ^
        offset.hashCode;
  }
}

class GetResponse {
  List<String> objectBodies;

  GetResponse({
    this.objectBodies,
  });

  GetResponse copyWith({
    List<String> objectBodies,
  }) {
    return GetResponse(
      objectBodies: objectBodies ?? this.objectBodies,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'objectBodies': objectBodies,
    };
  }

  factory GetResponse.fromMap(Map<String, dynamic> map) {
    if (map == null) return null;
  
    return GetResponse(
      objectBodies: List<String>.from(map['objectBodies']),
    );
  }

  String toJson() => json.encode(toMap());

  factory GetResponse.fromJson(String source) => GetResponse.fromMap(json.decode(source));

  @override
  String toString() => 'GetResponse(objectBodies: $objectBodies)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;
  
    return o is GetResponse &&
      listEquals(o.objectBodies, objectBodies);
  }

  @override
  int get hashCode => objectBodies.hashCode;
}