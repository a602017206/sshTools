import Flutter
import UIKit
import Mobile

/// 把 gomobile 生成的 Go 入口（Mobile 模块）桥接到 Dart 侧 MethodChannel。
///
/// Go 的 MobileStart 内部会等健康检查通过，耗时可达数秒，必须放到后台队列
/// 执行，结果再回主队列交付给 Flutter。
public class SshtoolsGoPlugin: NSObject, FlutterPlugin {
    public static func register(with registrar: FlutterPluginRegistrar) {
        let channel = FlutterMethodChannel(name: "sshtools_go", binaryMessenger: registrar.messenger())
        let instance = SshtoolsGoPlugin()
        registrar.addMethodCallDelegate(instance, channel: channel)
    }

    public func handle(_ call: FlutterMethodCall, result: @escaping FlutterResult) {
        switch call.method {
        case "start":
            let args = call.arguments as? [String: Any]
            let dataDir = args?["dataDir"] as? String ?? ""
            if dataDir.isEmpty {
                result(FlutterError(code: "INVALID_ARGUMENT", message: "dataDir is empty", details: nil))
                return
            }
            DispatchQueue.global(qos: .userInitiated).async {
                var nsError: NSError?
                let ok = MobileStart(dataDir, &nsError)
                DispatchQueue.main.async {
                    if ok {
                        result(MobilePort())
                    } else {
                        result(FlutterError(
                            code: "GO_START_FAILED",
                            message: nsError?.localizedDescription ?? "unknown Go error",
                            details: nil))
                    }
                }
            }
        case "stop":
            MobileStop()
            result(nil)
        case "isRunning":
            result(MobileIsRunning())
        default:
            result(FlutterMethodNotImplemented)
        }
    }
}
