import 'package:json_annotation/json_annotation.dart';

part 'connection_model.g.dart';

/// SSH connection configuration model
@JsonSerializable()
class ConnectionModel {
  final String id;
  final String name;
  final String host;
  final int port;
  final String user;
  @JsonKey(name: 'auth_type')
  final String authType; // 'password' or 'key'
  @JsonKey(name: 'key_path')
  final String? keyPath;

  /// 服务端是否已保存该连接的密码。密码本身永不下发，这里只回一个标记，
  /// 供界面决定测试/连接时是否还要向用户索要密码。
  @JsonKey(name: 'has_password', includeToJson: false)
  final bool hasPassword;

  /// 临时密码：仅在新增/更新时由表单带给仓库层发送给服务端加密保存，
  /// 不参与 JSON 序列化（仓库手动拼进请求体），也不会从服务端回读。
  @JsonKey(includeFromJson: false, includeToJson: false)
  final String? password;

  final List<String>? tags;
  final Map<String, String>? metadata;

  ConnectionModel({
    required this.id,
    required this.name,
    required this.host,
    required this.port,
    required this.user,
    required this.authType,
    this.keyPath,
    this.hasPassword = false,
    this.password,
    this.tags,
    this.metadata,
  });

  factory ConnectionModel.fromJson(Map<String, dynamic> json) =>
      _$ConnectionModelFromJson(json);

  Map<String, dynamic> toJson() => _$ConnectionModelToJson(this);

  ConnectionModel copyWith({
    String? id,
    String? name,
    String? host,
    int? port,
    String? user,
    String? authType,
    String? keyPath,
    bool? hasPassword,
    String? password,
    List<String>? tags,
    Map<String, String>? metadata,
  }) {
    return ConnectionModel(
      id: id ?? this.id,
      name: name ?? this.name,
      host: host ?? this.host,
      port: port ?? this.port,
      user: user ?? this.user,
      authType: authType ?? this.authType,
      keyPath: keyPath ?? this.keyPath,
      hasPassword: hasPassword ?? this.hasPassword,
      password: password ?? this.password,
      tags: tags ?? this.tags,
      metadata: metadata ?? this.metadata,
    );
  }
}
