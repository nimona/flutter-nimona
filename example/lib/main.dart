import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/widgets.dart';

import 'package:path_provider/path_provider.dart';

import 'package:nimona/nimona.dart';
import 'package:nimona/models/init_request.dart';
import 'package:nimona/models/nimona_connection_info.dart';

void main() {
  if (!kIsWeb && (Platform.isLinux || Platform.isWindows)) {
    debugDefaultTargetPlatformOverride = TargetPlatform.fuchsia;
  }

  runApp(new MyApp());
}

class MyApp extends StatefulWidget {
  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  String _version = "unable to get nimona version";
  String _peerKey = "unable to get peer key";

  @override
  void initState() {
    super.initState();
    getApplicationDocumentsDirectory().then((directory) {
      String configPath = directory.path;
      if (Platform.isLinux || Platform.isMacOS) {
        configPath = configPath + "/.nimona-example";
      } else if (Platform.isWindows) {
        configPath = configPath + "\\.nimona-example";
      }
      final req = InitRequest(
        configPath: configPath,
      );
      Nimona.init(req).then((value) {
        Nimona.getConnectionInfo().then((ConnectionInfo value) {
          setState(() {
            _peerKey = value.publicKey;
          });
        });
      });
    });
    Nimona.version().then((version) {
      setState(() {
        _version = version;
      });
    });
  }

  @override
  void dispose() {
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        body: Container(
          child: Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Text("Nimona library version: " + _version),
                Text("Peer Public Key: " + _peerKey),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
