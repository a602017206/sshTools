import 'package:flutter/services.dart';

/// 内嵌 Go 核心的 Dart 侧入口。
///
/// 原生侧（Android AAR / iOS XCFramework，均由 gomobile 从仓库根目录的
/// `mobile` 包生成）负责在应用进程内启动 Go API 服务器；
/// 这里仅通过 MethodChannel 传递沙盒目录并拿回监听端口。
class SshtoolsGo {
  static const MethodChannel _channel = MethodChannel('sshtools_go');

  /// 在 [dataDir]（应用沙盒目录）下启动内嵌 Go 服务器，返回监听端口。
  ///
  /// 服务器仅监听 127.0.0.1 的随机端口。重复调用是幂等的，返回同一端口。
  /// 非移动平台（未注册插件）会抛 [MissingPluginException]，调用方应捕获后
  /// 回退到外部 API 服务器地址。
  static Future<int> start(String dataDir) async {
    final port = await _channel.invokeMethod<int>('start', <String, dynamic>{
      'dataDir': dataDir,
    });
    if (port == null || port <= 0) {
      throw PlatformException(
        code: 'GO_START_FAILED',
        message: 'Embedded Go server returned invalid port: $port',
      );
    }
    return port;
  }

  /// 停止内嵌 Go 服务器。未运行时调用无副作用。
  static Future<void> stop() async {
    await _channel.invokeMethod<void>('stop');
  }

  /// 内嵌 Go 服务器是否正在运行；插件未注册时返回 false。
  static Future<bool> isRunning() async {
    try {
      return await _channel.invokeMethod<bool>('isRunning') ?? false;
    } on MissingPluginException {
      return false;
    }
  }
}
