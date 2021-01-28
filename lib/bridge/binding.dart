import 'dart:async';
import 'dart:ffi' as ffi;
import 'dart:io' show Platform;
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:ffi/ffi.dart';
import 'package:nimona/bridge/ffi.dart';
import 'package:nimona/bridge/isolate.dart';
import 'package:nimona/models/get_request.dart';

typedef StartWorkType = ffi.Void Function(ffi.Int64 port);
typedef StartWorkFunc = void Function(int port);

class Binding {
  static final String _callFuncName = 'NimonaBridgeCall';
  static final Binding _singleton = Binding._internal();

  ffi.DynamicLibrary _library;

  factory Binding() {
    return _singleton;
  }

  Binding._internal() {
    _library = openLib();
  }

  Future<Uint8List> callAsync(String name, Uint8List payload) async {
    final port = ReceivePort();
    final args = IsolateArguments(name, payload, port.sendPort);

    Isolate.spawn<IsolateArguments>(
      callBridge,
      args,
      onError: port.sendPort,
      onExit: port.sendPort,
    );

    Completer<Uint8List> completer = new Completer();

    StreamSubscription subscription;
    subscription = port.listen((message) async {
      await subscription?.cancel();
      completer.complete(message);
    });
    return completer.future;
  }

  static callBridge(IsolateArguments args) async {
    var result = await Binding().call(args.name, args.payload);
    args.port.send(result);
  }

  Future<Uint8List> call(String name, Uint8List payload) async {
    final callable = _library
        .lookup<ffi.NativeFunction<call_func>>(_callFuncName)
        .asFunction<Call>();

    final pointer = allocate<ffi.Uint8>(count: payload.length);

    // https://github.com/dart-lang/ffi/issues/27
    // https://github.com/objectbox/objectbox-dart/issues/69
    for (var i = 0; i < payload.length; i++) {
      pointer[i] = payload[i];
    }

    final voidStar = pointer.cast<ffi.Void>();
    final nameRef = toUtf8(name);

    final result =
        callable(nameRef, voidStar, payload.length).cast<FFIBytesReturn>().ref;

    free(nameRef);
    free(voidStar);

    handleError(result.error, result.addressOf);

    final output = result.message.cast<ffi.Uint8>().asTypedList(result.size);
    free(result.addressOf);
    return output;
  }

  void init() {
    call("init", Uint8List(0));
  }
 
  Future<List<String>> get(GetRequest req) async {
    String reqJSON = req.toJson();
    Uint8List body = await callAsync("get", Uint8List.fromList(reqJSON.codeUnits));
    GetResponse resp = GetResponse.fromJson(String.fromCharCodes(body));
    return resp.objectBodies;
  }
 
  Future<String> subscribe(String lookup) async {
    Uint8List r = await callAsync("subscribe", Uint8List.fromList(lookup.codeUnits));
    return String.fromCharCodes(r);
  }

  Future<void> cancel(String key) async {
    callAsync("cancel", Uint8List.fromList(key.codeUnits));
  }

  Stream<String> pop(String key) async* {
    while (true) {
      Uint8List r = await callAsync("pop", Uint8List.fromList(key.codeUnits));
      yield String.fromCharCodes(r);
    }
  }
 
  Future<void> requestStream(String rootHash) async {
    await callAsync("requestStream", Uint8List.fromList(rootHash.codeUnits));
  }
 
  Future<String> put(String objectJSON) async {
    Uint8List r = await callAsync("put", Uint8List.fromList(objectJSON.codeUnits));
    return String.fromCharCodes(r);
  }
 
  Future<String> getFeedRootHash(String feedRoothash) async {
    Uint8List r = await callAsync("getFeedRootHash", Uint8List.fromList(feedRoothash.codeUnits));
    return String.fromCharCodes(r);
  }

  void handleError(ffi.Pointer<Utf8> error, ffi.Pointer pointer) {
    if (error.address != ffi.nullptr.address) {
      var message = fromUtf8(error);
      free(pointer);
      throw message;
    }
  }

  ffi.Pointer<Utf8> toUtf8(String text) {
    return text == null ? Utf8.toUtf8("") : Utf8.toUtf8(text);
  }

  String fromUtf8(ffi.Pointer<Utf8> text) {
    return text == null ? "" : Utf8.fromUtf8(text);
  }

  ffi.DynamicLibrary openLib() {
    if (Platform.isMacOS || Platform.isIOS) {
      return ffi.DynamicLibrary.process();
    }
    throw("not implemented");
  }
}
