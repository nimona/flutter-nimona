import 'package:nimona/models/nimona_basic_object.dart';
import 'package:nimona/models/nimona_connection_info.dart';
import 'package:nimona/models/typed.dart';
import 'package:nimona/models/types.dart';
 
NimonaTyped unmarshal(String body) {
  final typed = BasicObject.fromJson(body);
  switch (typed.type) {
    case ConnectionInfoType:
      return ConnectionInfo.fromJson(body);
    default:
      throw ('unknown object type');
  }
}
