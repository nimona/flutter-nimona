import 'dart:convert';

import 'package:nimona/models/nimona_metadata.dart';
import 'package:nimona/models/typed.dart';
import 'package:nimona/models/types.dart';

class ConnectionInfo implements NimonaTyped {
  String? cid;
  String type;
  MetadataM? metadata;
  final List<String>? addresses;
  final List<String>? objectFormats;
  final String? publicKey;
  final int? version;

  ConnectionInfo({
    this.cid,
    required this.type,
    this.metadata,
    this.addresses,
    this.objectFormats,
    this.publicKey,
    this.version,
  });

  String getType() {
    return this.type;
  }

  Map<String, dynamic> toMap() {
    return {
      '_cid:s': cid,
      '@type:s': ConnectionInfoType,
      '@metadata:m': metadata?.toMap(),
      'addresses:as': addresses,
      'objectFormats:as': objectFormats,
      'publicKey:s': publicKey,
      'version:i': version,
    };
  }

  factory ConnectionInfo.fromMap(Map<String, dynamic> map) {
    return ConnectionInfo(
      cid: map['_cid:s'],
      type: map['@type:s'],
      metadata: MetadataM.fromMap(map['@metadata:m']),
      addresses: List<String>.from(map['addresses:as']),
      objectFormats: List<String>.from(map['objectFormats:as']),
      publicKey: map['publicKey:s'],
      version: map['version:i'],
    );
  }

  String toJson() => json.encode(toMap());

  factory ConnectionInfo.fromJson(String source) => ConnectionInfo.fromMap(json.decode(source));
}
