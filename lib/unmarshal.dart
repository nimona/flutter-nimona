import 'package:nimona/models/basic_object.dart';
import 'package:nimona/models/connection_info.dart';
import 'package:nimona/models/typed.dart';
import 'package:nimona/models/types.dart';
 
NimonaTyped unmarshal(String body) {
  final typed = BasicObject.fromJson(body);
  switch (typed.typeS) {
    case ConnectionInfoType:
      return connectionInfoFromJson(body);
    default:
      throw ('unknown object type');
  }
}
