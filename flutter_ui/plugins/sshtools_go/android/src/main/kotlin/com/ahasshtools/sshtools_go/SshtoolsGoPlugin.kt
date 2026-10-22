package com.ahasshtools.sshtools_go

import android.os.Handler
import android.os.Looper
import io.flutter.embedding.engine.plugins.FlutterPlugin
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import io.flutter.plugin.common.MethodChannel.MethodCallHandler
import io.flutter.plugin.common.MethodChannel.Result
import mobile.Mobile

/**
 * 把 gomobile 生成的 Go 入口（mobile.Mobile）桥接到 Dart 侧 MethodChannel。
 *
 * Go 的 Start 内部会等健康检查通过，耗时可达数秒，必须放到后台线程执行，
 * 结果再回主线程交付给 Flutter。
 */
class SshtoolsGoPlugin : FlutterPlugin, MethodCallHandler {
    private lateinit var channel: MethodChannel
    private val mainHandler = Handler(Looper.getMainLooper())

    override fun onAttachedToEngine(binding: FlutterPlugin.FlutterPluginBinding) {
        channel = MethodChannel(binding.binaryMessenger, "sshtools_go")
        channel.setMethodCallHandler(this)
    }

    override fun onMethodCall(call: MethodCall, result: Result) {
        when (call.method) {
            "start" -> {
                val dataDir = call.argument<String>("dataDir")
                if (dataDir.isNullOrEmpty()) {
                    result.error("INVALID_ARGUMENT", "dataDir is empty", null)
                    return
                }
                Thread {
                    try {
                        Mobile.start(dataDir)
                        val port = Mobile.port()
                        mainHandler.post { result.success(port.toInt()) }
                    } catch (e: Exception) {
                        mainHandler.post {
                            result.error("GO_START_FAILED", e.message, null)
                        }
                    }
                }.start()
            }
            "stop" -> {
                Mobile.stop()
                result.success(null)
            }
            "isRunning" -> result.success(Mobile.isRunning())
            else -> result.notImplemented()
        }
    }

    override fun onDetachedFromEngine(binding: FlutterPlugin.FlutterPluginBinding) {
        channel.setMethodCallHandler(null)
    }
}
