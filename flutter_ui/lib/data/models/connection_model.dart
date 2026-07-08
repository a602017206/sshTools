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
      tags: tags ?? this.tags,
      metadata: metadata ?? this.metadata,
    );
  }
}
