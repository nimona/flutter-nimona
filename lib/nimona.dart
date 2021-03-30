import 'dart:async';

import 'package:nimona/bridge/binding.dart';
import 'package:nimona/models/nimona_connection_info.dart';
import 'package:nimona/models/get_request.dart';
import 'package:nimona/models/init_request.dart';
import 'package:nimona/models/subscribe_request.dart';
import 'package:nimona/unmarshal.dart';

class Nimona {
  static Future<void> init(InitRequest req) {
    return Binding().init(req);
  }

  static Future<List<String>> get(GetRequest req) {
    return Binding().get(req);
  }

  static Future<String> subscribe(SubscribeRequest req) {
    return Binding().subscribe(req);
  }

  static Future<String> version() {
    return Binding().version();
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

  static Future<ConnectionInfo> getConnectionInfo() async {
    final res = await Binding().getConnectionInfo();
    final typ = unmarshal(res);
    print(">>>>" + res);
    if (typ is ConnectionInfo) {
      print("+++"+ typ.publicKey!);
      return typ;
    } else {
      throw 'getConnectionInfo() ERROR';
    }
  }
}
