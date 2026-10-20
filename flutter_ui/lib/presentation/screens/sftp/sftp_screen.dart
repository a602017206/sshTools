import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'package:uuid/uuid.dart';

import '../../../data/models/connection_model.dart';
import '../../../data/models/sftp_file.dart';
import '../../providers/api_providers.dart';

/// SFTP 文件管理页：自己开一条专用 SSH 会话（凭 connection_id 回退
/// 已保存密码），浏览/新建/重命名/删除/上传/下载/文本查看编辑。
/// 退出页面即关这条会话，不占用终端的会话。
class SftpScreen extends ConsumerStatefulWidget {
  final ConnectionModel connection;
  final String authValue;

  const SftpScreen({
    super.key,
    required this.connection,
    this.authValue = '',
  });

  @override
  ConsumerState<SftpScreen> createState() => _SftpScreenState();
}

class _SftpScreenState extends ConsumerState<SftpScreen> {
  final String _sessionId = 'mobile-sftp-${const Uuid().v4()}';

  bool _ready = false;
  bool _loading = false;
  bool _busy = false;
  String? _error;
  String _currentPath = '';
  List<SftpFile> _files = const [];
  bool _disposed = false;

  @override
  void initState() {
    super.initState();
    _init();
  }

  Future<void> _init() async {
    try {
      await ref.read(sessionRepositoryProvider).connectSSH(
            sessionId: _sessionId,
            host: widget.connection.host,
            port: widget.connection.port,
            user: widget.connection.user,
            authType: widget.connection.authType,
            authValue: widget.authValue,
            connectionId: widget.connection.id,
            cols: 80,
            rows: 24,
          );
      _ready = true;
      await _load('');
    } catch (e) {
      if (mounted) setState(() => _error = '连接失败：$e');
    }
  }

  Future<void> _load(String path) async {
    if (!_ready || _disposed) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final repo = ref.read(sftpRepositoryProvider);
      final files = await repo.listFiles(_sessionId, path);
      var current = path;
      if (current.isEmpty) {
        // 首次进入：从条目路径反推当前目录；空目录时向服务端问 pwd
        if (files.isNotEmpty) {
          current = _dirname(files.first.path);
        } else {
          current = await repo.pwd(_sessionId);
        }
      }
      if (!mounted) return;
      setState(() {
        _files = files;
        _currentPath = current;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  String _dirname(String path) {
    final trimmed = path.endsWith('/') && path.length > 1
        ? path.substring(0, path.length - 1)
        : path;
    final idx = trimmed.lastIndexOf('/');
    if (idx < 0) return '.';
    if (idx == 0) return '/';
    return trimmed.substring(0, idx);
  }

  String _join(String dir, String name) {
    if (dir.isEmpty || dir == '.') return name;
    return dir.endsWith('/') ? '$dir$name' : '$dir/$name';
  }

  void _toast(String msg, {bool error = false}) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: error ? Colors.red : null,
        duration: Duration(seconds: error ? 6 : 3),
      ),
    );
  }

  Future<void> _runBusy(Future<void> Function() action, String okMsg) async {
    setState(() => _busy = true);
    try {
      await action();
      _toast(okMsg);
      await _load(_currentPath);
    } catch (e) {
      _toast(e.toString(), error: true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _createFolder() async {
    final name = await _promptText('新建文件夹', '文件夹名称');
    if (name == null || name.trim().isEmpty) return;
    await _runBusy(
      () => ref
          .read(sftpRepositoryProvider)
          .createDirectory(_sessionId, _join(_currentPath, name.trim())),
      '已创建 ${name.trim()}',
    );
  }

  Future<void> _rename(SftpFile file) async {
    final name = await _promptText('重命名', '新名称', initial: file.name);
    if (name == null || name.trim().isEmpty || name.trim() == file.name) return;
    await _runBusy(
      () => ref.read(sftpRepositoryProvider).rename(
            _sessionId,
            file.path,
            _join(_dirname(file.path), name.trim()),
          ),
      '已重命名为 ${name.trim()}',
    );
  }

  Future<void> _delete(SftpFile file) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('删除'),
        content: Text(
          file.isDir
              ? '确定删除文件夹「${file.name}」及其内容吗？此操作不可恢复。'
              : '确定删除文件「${file.name}」吗？此操作不可恢复。',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('删除'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    await _runBusy(
      () => ref.read(sftpRepositoryProvider).delete(_sessionId, file.path),
      '已删除 ${file.name}',
    );
  }

  Future<void> _upload() async {
    final picked = await FilePicker.platform.pickFiles();
    final file = picked?.files.single;
    if (file?.path == null) return;
    await _runBusy(
      () => ref.read(sftpRepositoryProvider).uploadFile(
            _sessionId,
            _currentPath,
            file!.path!,
            file.name,
          ),
      '已上传 ${file!.name}',
    );
  }

  Future<void> _download(SftpFile file) async {
    setState(() => _busy = true);
    try {
      final docs = await getApplicationDocumentsDirectory();
      final saved = await ref.read(sftpRepositoryProvider).downloadFile(
            _sessionId,
            file.path,
            '${docs.path}/downloads',
          );
      _toast('已下载到 $saved');
    } catch (e) {
      _toast(e.toString(), error: true);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _openText(SftpFile file) async {
    try {
      final content =
          await ref.read(sftpRepositoryProvider).readText(_sessionId, file.path);
      if (!mounted) return;
      final saved = await Navigator.push<bool>(
        context,
        MaterialPageRoute(
          builder: (context) => _TextEditorPage(
            title: file.name,
            initialContent: content,
            onSave: (text) => ref
                .read(sftpRepositoryProvider)
                .writeText(_sessionId, file.path, text),
          ),
        ),
      );
      if (saved == true) _toast('已保存 ${file.name}');
    } catch (e) {
      _toast('无法打开（可能不是文本文件）：$e', error: true);
    }
  }

  Future<String?> _promptText(String title, String label, {String initial = ''}) {
    final controller = TextEditingController(text: initial);
    return showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: TextField(
          controller: controller,
          autofocus: true,
          textCapitalization: TextCapitalization.none,
          autocorrect: false,
          enableSuggestions: false,
          decoration: InputDecoration(
            labelText: label,
            border: const OutlineInputBorder(),
          ),
          onSubmitted: (v) => Navigator.pop(context, v),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, controller.text),
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _showFileActions(SftpFile file) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              title: Text(file.name, style: const TextStyle(fontWeight: FontWeight.bold)),
              subtitle: Text(file.path),
            ),
            const Divider(height: 1),
            if (!file.isDir) ...[
              ListTile(
                leading: const Icon(Icons.download),
                title: const Text('下载到本机'),
                onTap: () {
                  Navigator.pop(context);
                  _download(file);
                },
              ),
              ListTile(
                leading: const Icon(Icons.edit_note),
                title: const Text('查看 / 编辑文本'),
                onTap: () {
                  Navigator.pop(context);
                  _openText(file);
                },
              ),
            ],
            ListTile(
              leading: const Icon(Icons.drive_file_rename_outline),
              title: const Text('重命名'),
              onTap: () {
                Navigator.pop(context);
                _rename(file);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('删除', style: TextStyle(color: Colors.red)),
              onTap: () {
                Navigator.pop(context);
                _delete(file);
              },
            ),
          ],
        ),
      ),
    );
  }

  @override
  void dispose() {
    _disposed = true;
    if (_ready) {
      // 页面销毁即关专用会话；不在 dispose 里 await
      ref.read(sessionRepositoryProvider).closeSession(_sessionId).catchError((_) {});
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final canGoUp = _currentPath.isNotEmpty &&
        _currentPath != '/' &&
        _currentPath != '.';

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('文件 · ${widget.connection.name}',
                style: const TextStyle(fontSize: 16)),
            Text(
              _currentPath.isEmpty ? '正在打开...' : _currentPath,
              style: const TextStyle(fontSize: 12),
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.arrow_upward),
            tooltip: '上级目录',
            onPressed: canGoUp ? () => _load(_dirname(_currentPath)) : null,
          ),
          IconButton(
            icon: const Icon(Icons.create_new_folder),
            tooltip: '新建文件夹',
            onPressed: _ready ? _createFolder : null,
          ),
          IconButton(
            icon: const Icon(Icons.upload_file),
            tooltip: '上传文件',
            onPressed: _ready ? _upload : null,
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: '刷新',
            onPressed: _ready ? () => _load(_currentPath) : null,
          ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_error != null && !_ready) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Text(_error!, style: const TextStyle(color: Colors.red)),
        ),
      );
    }
    if (!_ready || _loading && _files.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    return Stack(
      children: [
        if (_files.isEmpty)
          const Center(child: Text('此目录为空'))
        else
          RefreshIndicator(
            onRefresh: () => _load(_currentPath),
            child: ListView.builder(
              itemCount: _files.length,
              itemBuilder: (context, index) {
                final file = _files[index];
                return ListTile(
                  leading: Icon(
                    file.isDir
                        ? Icons.folder
                        : Icons.insert_drive_file_outlined,
                    color: file.isDir
                        ? Colors.amber
                        : Theme.of(context).colorScheme.primary,
                  ),
                  title: Text(file.name),
                  subtitle: Text(
                    file.isDir ? '文件夹' : _formatSize(file.size),
                  ),
                  trailing: IconButton(
                    icon: const Icon(Icons.more_horiz),
                    onPressed: () => _showFileActions(file),
                  ),
                  onTap: file.isDir
                      ? () => _load(file.path)
                      : () => _showFileActions(file),
                  onLongPress: () => _showFileActions(file),
                );
              },
            ),
          ),
        if (_busy || _loading)
          const Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: LinearProgressIndicator(),
          ),
        if (_error != null && _ready)
          Positioned(
            left: 12,
            right: 12,
            bottom: 12,
            child: Material(
              color: Colors.red.withValues(alpha: 0.9),
              borderRadius: BorderRadius.circular(8),
              child: Padding(
                padding: const EdgeInsets.all(10),
                child: Text(
                  _error!,
                  style: const TextStyle(color: Colors.white, fontSize: 12),
                ),
              ),
            ),
          ),
      ],
    );
  }

  String _formatSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }
}

/// 远端文本查看/编辑页：保存即回写远端文件。
class _TextEditorPage extends StatefulWidget {
  final String title;
  final String initialContent;
  final Future<void> Function(String content) onSave;

  const _TextEditorPage({
    required this.title,
    required this.initialContent,
    required this.onSave,
  });

  @override
  State<_TextEditorPage> createState() => _TextEditorPageState();
}

class _TextEditorPageState extends State<_TextEditorPage> {
  late final TextEditingController _controller;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialContent);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      await widget.onSave(_controller.text);
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('保存失败：$e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        actions: [
          TextButton(
            onPressed: _saving ? null : _save,
            child: const Text('保存'),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(12),
        child: TextField(
          controller: _controller,
          maxLines: null,
          expands: true,
          textAlignVertical: TextAlignVertical.top,
          keyboardType: TextInputType.multiline,
          textCapitalization: TextCapitalization.none,
          autocorrect: false,
          enableSuggestions: false,
          style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
          decoration: const InputDecoration(border: OutlineInputBorder()),
        ),
      ),
    );
  }
}