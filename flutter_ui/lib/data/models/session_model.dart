/// SSH session model
class SessionModel {
  final String sessionId;
  final String host;
  final int port;
  final String user;
  final String authType;
  final int cols;
  final int rows;
  final DateTime connectedAt;
  bool isActive;

  SessionModel({
    required this.sessionId,
    required this.host,
    required this.port,
    required this.user,
    required this.authType,
    required this.cols,
    required this.rows,
    required this.connectedAt,
    this.isActive = true,
  });

  String get displayName => '$user@$host:$port';

  SessionModel copyWith({
    String? sessionId,
    String? host,
    int? port,
    String? user,
    String? authType,
    int? cols,
    int? rows,
    DateTime? connectedAt,
    bool? isActive,
  }) {
    return SessionModel(
      sessionId: sessionId ?? this.sessionId,
      host: host ?? this.host,
      port: port ?? this.port,
      user: user ?? this.user,
      authType: authType ?? this.authType,
      cols: cols ?? this.cols,
      rows: rows ?? this.rows,
      connectedAt: connectedAt ?? this.connectedAt,
      isActive: isActive ?? this.isActive,
    );
  }
}
