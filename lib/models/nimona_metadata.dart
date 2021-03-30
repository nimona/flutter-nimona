import 'dart:convert';

import 'package:flutter/foundation.dart';

class MetadataM {
  String? owner;
  List<String>? parents;
  String? stream;
  String? datetime;

  MetadataM({
    this.owner,
    this.parents,
    this.stream,
    this.datetime,
  });

  MetadataM copyWith({
    String? owner,
    List<String>? parents,
    String? stream,
    String? datetime,
  }) {
    return MetadataM(
      owner: owner ?? this.owner,
      parents: parents ?? this.parents,
      stream: stream ?? this.stream,
      datetime: datetime ?? this.datetime,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'owner:s': owner,
      'parents:as': parents,
      'stream:s': stream,
      'datetime:s': datetime,
    };
  }

  factory MetadataM.fromMap(Map<String, dynamic>? map) {
    return MetadataM(
      owner: map == null ? null : map['owner:s'],
      stream: map == null ? null : map['stream:s'],
      datetime: map == null ? null : map['datetime:s'],
    );
  }

  String toJson() => json.encode(toMap());

  factory MetadataM.fromJson(String source) => MetadataM.fromMap(json.decode(source));

  @override
  String toString() => 'MetadataM(owner: $owner, parents: $parents, stream: $stream)';

  @override
  bool operator ==(Object o) {
    if (identical(this, o)) return true;
  
    return o is MetadataM &&
      o.owner == owner &&
      listEquals(o.parents, parents) &&
      o.stream == stream;
  }

  @override
  int get hashCode => owner.hashCode ^ parents.hashCode ^ stream.hashCode;
}
