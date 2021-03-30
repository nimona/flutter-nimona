import 'dart:convert';

import 'package:nimona/models/nimona_metadata.dart';
import 'package:nimona/models/typed.dart';

class BasicObject implements NimonaTyped {
  String? cid;
  String type;
  MetadataM? metadata;

  BasicObject({
    this.cid,
    required this.type,
    this.metadata,
  });

  Map<String, dynamic> toMap() {
    return {
      '_cid:s': cid,
      '@type:s': type,
      '@metadata:m': metadata?.toMap(),
    };
  }

  String getType() {
    return this.type;
  }

  factory BasicObject.fromMap(Map<String, dynamic> map) {
    return BasicObject(
      cid: map['_cid:s'],
      type: map['@type:s'],
      metadata: MetadataM.fromMap(map['@metadata:m']),
    );
  }

  String toJson() => json.encode(toMap());

  factory BasicObject.fromJson(String source) => BasicObject.fromMap(json.decode(source));
}
