import 'dart:async';
import 'dart:convert';
import 'dart:ffi' as ffi;
import 'dart:io' show Platform;
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'package:ffi/ffi.dart';

import 'package:nimona/bridge/ffi.dart';
import 'package:nimona/bridge/isolate.dart';
import 'package:nimona/models/get_request.dart';
import 'package:nimona/models/init_request.dart';
import 'package:nimona/models/subscribe_request.dart';

typedef StartWorkType = ffi.Void Function(ffi.Int64 port);
typedef StartWorkFunc = void Function(int port);

class Binding {
  static final String _callFuncName = 'NimonaBridgeCall';
  static final String _libraryName = 'libnimona';
  static final Binding _singleton = Binding._internal();

  ffi.DynamicLibrary? _library;

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

    StreamSubscription? subscription;
    // TODO
    subscription = port.listen((message) async {
      try {
        await subscription?.cancel();
        if (message == null) {
          completer.complete();
        } else {
          completer.complete(message);
        }
      } catch (e) {
        print('++ callAsync(' + name + ') ERROR err=' + e.toString());
        throw e;
      }
    });
    return completer.future;
  }

  static callBridge(IsolateArguments args) async {
    var result = await Binding().call(args.name, args.payload);
    args.port.send(result);
  }

  Future<Uint8List> call(String name, Uint8List payload) async {
    final callable = _library!
        .lookup<ffi.NativeFunction<call_func>>(_callFuncName)
        .asFunction<Call>();

    final pointer = malloc<ffi.Uint8>(payload.length);

    // https://github.com/dart-lang/ffi/issues/27
    // https://github.com/objectbox/objectbox-dart/issues/69
    for (var i = 0; i < payload.length; i++) {
      pointer[i] = payload[i];
    }
    final payloadPointer = pointer.cast<ffi.Void>();
    final namePointer = toUtf8(name);

    final result = callable(namePointer, payloadPointer, payload.length);

    malloc.free(namePointer);
    malloc.free(payloadPointer);

    handleError(result.ref.error, result);

    final output =
        result.ref.message.cast<ffi.Uint8>().asTypedList(result.ref.size);
    freeResult(result);
    return output;
  }

  Future<void> init(InitRequest req) async {
    try {
      String reqJSON = req.toJson();
      Uint8List body = await callAsync(
        'init',
        stringToBytes(reqJSON),
      );
      if (body == null || body.isEmpty) {
        throw 'got empty response';
      }
      return;
    } on Exception catch (e) {
      throw e;
    }
  }

  Future<List<String>> get(GetRequest req) async {
    try {
      String reqJSON = req.toJson();
      Uint8List body = await callAsync(
        'get',
        stringToBytes(reqJSON),
      );
      if (body == null || body.isEmpty) {
        throw 'got empty response';
      }
      GetResponse resp = GetResponse.fromJson(
        bytesToString(body),
      );
      return resp.objectBodies;
    } on Exception catch (e) {
      throw e;
    }
  }

  Future<String> version() async {
    try {
      Uint8List r = await callAsync(
        'version',
        Uint8List(0),
      );
      if (r == null || r.isEmpty) {
        throw 'got empty response';
      }
      return bytesToString(r);
    } on Exception catch (e) {
      throw e;
    }
  }

  Future<String> subscribe(SubscribeRequest req) async {
    try {
      String reqJSON = req.toJson();
      Uint8List r = await callAsync(
        'subscribe',
        stringToBytes(reqJSON),
      );
      if (r == null || r.isEmpty) {
        throw 'got empty response';
      }
      return bytesToString(r);
    } on Exception catch (e) {
      throw e;
    }
  }

  Future<void> cancel(String key) async {
    callAsync("cancel", stringToBytes(key));
  }

  Stream<String> pop(String key) async* {
    while (true) {
      Uint8List r = await callAsync("pop", stringToBytes(key));
      yield bytesToString(r);
    }
  }

  Future<void> requestStream(String rootHash) async {
    await callAsync("requestStream", stringToBytes(rootHash));
  }

  Future<String> put(String objectJSON) async {
    try {
      Uint8List r = await callAsync(
        'put',
        stringToBytes(objectJSON),
      );
      if (r == null || r.isEmpty) {
        throw 'got empty response';
      }
      return bytesToString(r);
    } on Exception catch (e) {
      throw e;
    }
  }

  Future<String> getFeedRootHash(String feedRoothash) async {
    Uint8List r =
        await callAsync("getFeedRootHash", stringToBytes(feedRoothash));
    return bytesToString(r);
  }

  Future<String> getConnectionInfo() async {
    Uint8List r = await callAsync("getConnectionInfo", Uint8List(0));
    return bytesToString(r);
  }

  void handleError(
      ffi.Pointer<Utf8> error, ffi.Pointer<FFIBytesReturn> result) {
    if (error.address != ffi.nullptr.address) {
      var message = fromUtf8(error);
      freeResult(result);
      throw message;
    }
  }

  ffi.Pointer<Utf8> toUtf8(String? text) {
    return text == null ? "".toNativeUtf8() : text.toNativeUtf8();
  }

  String fromUtf8(ffi.Pointer<Utf8>? text) {
    return text == null ? "" : text.toDartString();
  }

  void freeResult(ffi.Pointer<FFIBytesReturn> result) {
    if (!Platform.isWindows) {
      malloc.free(result);
    }
  }

  ffi.DynamicLibrary openLib() {
    if (Platform.isMacOS) {
      return ffi.DynamicLibrary.open("$_libraryName.dylib");
    }
    if (Platform.isWindows) {
      return ffi.DynamicLibrary.open("$_libraryName.dll");
    }
    if (Platform.isIOS) {
      return ffi.DynamicLibrary.process();
    }
    if (Platform.isLinux) {
      return ffi.DynamicLibrary.open("$_libraryName.so");
    }
    return ffi.DynamicLibrary.open("$_libraryName.so");
  }
}

Uint8List stringToBytes(String s) {
  return Uint8List.fromList(utf8.encode(s));
}

String bytesToString(Uint8List b) {
  return utf8.decode(b);
}
