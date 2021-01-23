import 'dart:async';

import 'package:nimona/bridge/binding.dart';

class Nimona {
  static Future<String> subscribe(String lookup) {
    return Binding().subscribe(lookup);
  }

  static Stream<String> pop(String key) {
    return Binding().pop(key);
  }

  static Future<void> requestStream(String rootHash) {
    return Binding().requestStream(rootHash);
  }

  static Future<String> put(String objectJSON) {
    return Binding().put(objectJSON);
  }

  static Future<String> getFeedRootHash(String streamRootType) {
    return Binding().getFeedRootHash(streamRootType);
  }
}
