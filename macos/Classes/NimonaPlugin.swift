import Cocoa
import FlutterMacOS

public class NimonaPlugin: NSObject, FlutterPlugin {
  public static func register(with registrar: FlutterPluginRegistrar) {
    let channel = FlutterMethodChannel(name: "nimona", binaryMessenger: registrar.messenger)
    let instance = NimonaPlugin()
    registrar.addMethodCallDelegate(instance, channel: channel)
  }

  public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
      result(FlutterMethodNotImplemented)
  }
}
