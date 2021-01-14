import 'dart:async';

import 'package:nimona/bridge/binding.dart';

class Nimona {
  static void subscribe(String name) {
    return Binding().subscribe(name);
  }

  static Stream<String> pop(String name) {
    return Binding().pop(name);
  }
}
