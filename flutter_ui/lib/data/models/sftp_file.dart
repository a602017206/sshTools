/// 远端文件/目录条目（对应 Go 侧 ssh.FileInfo 的 JSON）。
class SftpFile {
  final String name;
  final String path;
  final int size;
  final String mode;
  final DateTime? modTime;
  final bool isDir;
  final bool isSymlink;

  const SftpFile({
    required this.name,
    required this.path,
    required this.size,
    required this.mode,
    required this.modTime,
    required this.isDir,
    required this.isSymlink,
  });

  factory SftpFile.fromJson(Map<String, dynamic> json) {
    return SftpFile(
      name: json['name'] as String? ?? '',
      path: json['path'] as String? ?? '',
      size: (json['size'] as num?)?.toInt() ?? 0,
      mode: json['mode'] as String? ?? '',
      modTime: switch (json['mod_time']) {
        final String s => DateTime.tryParse(s),
        _ => null,
      },
      isDir: json['is_dir'] as bool? ?? false,
      isSymlink: json['is_symlink'] as bool? ?? false,
    );
  }
}
