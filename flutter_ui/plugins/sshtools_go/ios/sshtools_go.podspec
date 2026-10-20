#
# gomobile 生成的 Mobile.xcframework 由 CI 在打包前放入 Frameworks/
# （见 .github/workflows/mobile.yml），因此该目录不纳入版本管理。
#
Pod::Spec.new do |s|
  s.name             = 'sshtools_go'
  s.version          = '0.1.0'
  s.summary          = 'AHaSSHTools 内嵌 Go 核心的 Flutter 桥接插件'
  s.description      = <<-DESC
在应用进程内启动 gomobile 编译的 Go API 服务器，供 Flutter 侧通过本机回环调用。
                       DESC
  s.homepage         = 'https://github.com/a602017206/sshTools'
  s.license          = { :type => 'MIT' }
  s.author           = { 'AHaSSHTools' => '602017206@qq.com' }
  s.source           = { :path => '.' }
  s.source_files     = 'Classes/**/*'
  # Go 内嵌二进制（cgo DNS 解析）需要 libresolv，否则链接报 _res_9_* 未定义
  s.libraries        = 'resolv'
  s.dependency 'Flutter'
  s.platform         = :ios, '13.0'
  s.vendored_frameworks = 'Frameworks/Mobile.xcframework'
  s.pod_target_xcconfig = { 'DEFINES_MODULE' => 'YES', 'EXCLUDED_ARCHS[sdk=iphonesimulator*]' => 'i386' }
  s.swift_version    = '5.0'
end
