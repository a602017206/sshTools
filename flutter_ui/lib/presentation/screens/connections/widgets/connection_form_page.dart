import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../data/models/connection_model.dart';

/// 新增/编辑连接的全屏表单页。
///
/// 早期版本是桌面式弹窗，且没有密码字段：移动端小屏上既挤又存不下密码，
/// 保存后的连接既不能删也不能改（空 ID 问题已在服务端修复）。
/// 这里与桌面端字段对齐：名称、主机、端口、用户名、认证方式、密码、
/// 私钥路径、标签，保存即经服务端加密存储密码。
class ConnectionFormPage extends StatefulWidget {
  final ConnectionModel? connection;

  /// 克隆模式：以 [connection] 预填，但保存走新增（ID 清空、名称加「副本」，
  /// 密码不复制——服务端从不下发密码，必须重新输入）。
  final bool clone;

  /// 已有分组名（来自现有连接），供分组输入框下方快选
  final List<String> existingGroups;

  const ConnectionFormPage({
    super.key,
    this.connection,
    this.clone = false,
    this.existingGroups = const [],
  });

  @override
  State<ConnectionFormPage> createState() => _ConnectionFormPageState();
}

class _ConnectionFormPageState extends State<ConnectionFormPage> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameController;
  late final TextEditingController _hostController;
  late final TextEditingController _portController;
  late final TextEditingController _userController;
  late final TextEditingController _passwordController;
  late final TextEditingController _keyPathController;
  late final TextEditingController _groupController;
  late final TextEditingController _tagsController;

  String _authType = 'password';
  bool _obscurePassword = true;

  bool get _isEdit => widget.connection != null && !widget.clone;
  bool get _hasSavedPassword => !_isEdit ? false : (widget.connection?.hasPassword ?? false);

  @override
  void initState() {
    super.initState();
    final conn = widget.connection;
    _nameController = TextEditingController(
      text: widget.clone && conn != null ? '${conn.name} 副本' : (conn?.name ?? ''),
    );
    _hostController = TextEditingController(text: conn?.host ?? '');
    _portController = TextEditingController(text: conn?.port.toString() ?? '22');
    _userController = TextEditingController(text: conn?.user ?? '');
    _passwordController = TextEditingController();
    _keyPathController = TextEditingController(text: conn?.keyPath ?? '');
    // 分组与桌面端同一约定：tags 的第一个就是分组，其余才是普通标签
    final tags = conn?.tags ?? const <String>[];
    _groupController = TextEditingController(
      text: tags.isNotEmpty ? tags.first : '',
    );
    _tagsController = TextEditingController(
      text: tags.length > 1 ? tags.sublist(1).join(', ') : '',
    );
    _authType = conn?.authType ?? 'password';
  }

  @override
  void dispose() {
    _nameController.dispose();
    _hostController.dispose();
    _portController.dispose();
    _userController.dispose();
    _passwordController.dispose();
    _keyPathController.dispose();
    _groupController.dispose();
    _tagsController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.clone ? '克隆连接' : (_isEdit ? '编辑连接' : '新增连接')),
        actions: [
          TextButton(
            onPressed: _saveConnection,
            child: const Text('保存'),
          ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(
                labelText: '连接名称',
                hintText: '例如：生产服务器',
                prefixIcon: Icon(Icons.label),
                border: OutlineInputBorder(),
              ),
              textCapitalization: TextCapitalization.none,
              autocorrect: false,
              validator: (value) {
                if (value == null || value.trim().isEmpty) {
                  return '请输入连接名称';
                }
                return null;
              },
            ),
            const SizedBox(height: 16),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  flex: 3,
                  child: TextFormField(
                    controller: _hostController,
                    decoration: const InputDecoration(
                      labelText: '主机地址',
                      hintText: '192.168.1.100 或域名',
                      prefixIcon: Icon(Icons.dns),
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.url,
                    textCapitalization: TextCapitalization.none,
                    autocorrect: false,
                    enableSuggestions: false,
                    validator: (value) {
                      if (value == null || value.trim().isEmpty) {
                        return '请输入主机地址';
                      }
                      return null;
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  flex: 1,
                  child: TextFormField(
                    controller: _portController,
                    decoration: const InputDecoration(
                      labelText: '端口',
                      hintText: '22',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType: TextInputType.number,
                    inputFormatters: [
                      FilteringTextInputFormatter.digitsOnly,
                    ],
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return '必填';
                      }
                      final port = int.tryParse(value);
                      if (port == null || port < 1 || port > 65535) {
                        return '端口无效';
                      }
                      return null;
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _userController,
              decoration: const InputDecoration(
                labelText: '用户名',
                hintText: 'root',
                prefixIcon: Icon(Icons.person),
                border: OutlineInputBorder(),
              ),
              // 用户名大小写敏感：关掉移动键盘的自动大写与纠错，
              // 否则 root 会被改成 Root 导致登录失败/看起来改不了
              textCapitalization: TextCapitalization.none,
              autocorrect: false,
              enableSuggestions: false,
              validator: (value) {
                if (value == null || value.trim().isEmpty) {
                  return '请输入用户名';
                }
                return null;
              },
            ),
            const SizedBox(height: 16),
            InputDecorator(
              decoration: const InputDecoration(
                labelText: '认证方式',
                prefixIcon: Icon(Icons.security),
                border: OutlineInputBorder(),
              ),
              child: DropdownButtonHideUnderline(
                child: DropdownButton<String>(
                  value: _authType,
                  isExpanded: true,
                  items: const [
                    DropdownMenuItem(value: 'password', child: Text('密码')),
                    DropdownMenuItem(value: 'key', child: Text('私钥')),
                  ],
                  onChanged: (value) {
                    if (value != null) {
                      setState(() => _authType = value);
                    }
                  },
                ),
              ),
            ),
            const SizedBox(height: 16),
            if (_authType == 'password') ...[
              TextFormField(
                controller: _passwordController,
                obscureText: _obscurePassword,
                decoration: InputDecoration(
                  labelText: '密码',
                  hintText: _isEdit && _hasSavedPassword ? '已保存，留空不修改' : '请输入密码',
                  prefixIcon: const Icon(Icons.lock),
                  border: const OutlineInputBorder(),
                  suffixIcon: IconButton(
                    icon: Icon(
                      _obscurePassword ? Icons.visibility : Icons.visibility_off,
                    ),
                    onPressed: () {
                      setState(() => _obscurePassword = !_obscurePassword);
                    },
                  ),
                ),
                textCapitalization: TextCapitalization.none,
                autocorrect: false,
                enableSuggestions: false,
                validator: (value) {
                  // 新增/克隆时必须填密码（克隆不复制原密码：服务端从不下发）；
                  // 编辑时留空表示沿用已保存密码
                  if (!_isEdit && (value == null || value.isEmpty)) {
                    return widget.clone ? '克隆不会复制原密码，请重新输入' : '请输入密码';
                  }
                  if (_isEdit && !_hasSavedPassword && (value == null || value.isEmpty)) {
                    return '该连接尚未保存密码，请输入';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 8),
              Text(
                '密码经服务端加密后保存在本机，不会明文写入配置文件。',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context)
                          .colorScheme
                          .onSurface
                          .withValues(alpha: 0.6),
                    ),
              ),
            ] else ...[
              TextFormField(
                controller: _keyPathController,
                decoration: const InputDecoration(
                  labelText: '私钥路径',
                  hintText: '/path/to/id_rsa',
                  prefixIcon: Icon(Icons.vpn_key),
                  border: OutlineInputBorder(),
                ),
                textCapitalization: TextCapitalization.none,
                autocorrect: false,
                enableSuggestions: false,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return '请输入私钥路径';
                  }
                  return null;
                },
              ),
            ],
            const SizedBox(height: 16),
            TextFormField(
              controller: _groupController,
              decoration: const InputDecoration(
                labelText: '分组（可选）',
                hintText: '例如：生产环境',
                prefixIcon: Icon(Icons.folder_open),
                border: OutlineInputBorder(),
              ),
              textCapitalization: TextCapitalization.none,
              autocorrect: false,
            ),
            if (widget.existingGroups.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: [
                  for (final g in widget.existingGroups)
                    ActionChip(
                      label: Text(g),
                      onPressed: () => _groupController.text = g,
                    ),
                ],
              ),
            ],
            const SizedBox(height: 16),
            TextFormField(
              controller: _tagsController,
              decoration: const InputDecoration(
                labelText: '其它标签（逗号分隔，可选）',
                hintText: 'Web, 跳板机',
                prefixIcon: Icon(Icons.local_offer),
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 32),
            FilledButton.icon(
              onPressed: _saveConnection,
              icon: const Icon(Icons.save),
              label: Text(widget.clone ? '克隆为新连接' : (_isEdit ? '保存修改' : '添加连接')),
              style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(48),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _saveConnection() {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    // 分组占 tags 首位（桌面端约定），其余标签跟随其后并剔掉与分组同名的
    final group = _groupController.text.trim();
    final otherTags = _tagsController.text
        .split(',')
        .map((t) => t.trim())
        .where((t) => t.isNotEmpty && t != group)
        .toList();
    final tags = [if (group.isNotEmpty) group, ...otherTags];

    final password = _passwordController.text;
    final connection = ConnectionModel(
      id: _isEdit ? (widget.connection?.id ?? '') : '',
      name: _nameController.text.trim(),
      host: _hostController.text.trim(),
      port: int.parse(_portController.text),
      user: _userController.text.trim(),
      authType: _authType,
      keyPath: _authType == 'key' ? _keyPathController.text.trim() : null,
      hasPassword: _hasSavedPassword,
      password: _authType == 'password' && password.isNotEmpty ? password : null,
      tags: tags,
    );

    Navigator.pop(context, connection);
  }
}
