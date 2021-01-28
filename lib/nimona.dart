import 'dart:async';

import 'package:nimona/bridge/binding.dart';
import 'package:nimona/models/get_request.dart';

class Nimona {
  static void init() {
    return Binding().init();
  }

  static Future<List<String>> get(GetRequest req) {
    return Binding().get(req);
  }

  static Future<String> subscribe(String lookup) {
    return Binding().subscribe(lookup);
  }

  static Future<void> cancel(String key) {
    return Binding().cancel(key);
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
