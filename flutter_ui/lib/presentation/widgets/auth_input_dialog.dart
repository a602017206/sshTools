import 'package:flutter/material.dart';

import '../../data/models/connection_model.dart';

/// 确保拿到一次连接可用的凭据：
/// - 已有保存密码（或私钥认证且路径已存）→ 直接返回空串，
///   服务端按 connection_id 回退已保存凭据；
/// - 否则弹窗索要一次临时凭据，返回用户输入；
/// - 用户取消 → 返回 null，调用方应中止本次操作。
///
/// 连接成功/测试成功后服务端会把临时密码回存，后续不再弹窗。
Future<String?> requestAuthValue(
  BuildContext context,
  ConnectionModel connection,
) async {
  final canUseSaved = connection.hasPassword ||
      (connection.authType == 'key' && (connection.keyPath?.isNotEmpty ?? false));
  if (canUseSaved) {
    return '';
  }
  return showDialog<String>(
    context: context,
    builder: (context) => AuthInputDialog(connection: connection),
  );
}

/// 无已保存密码时的一次性凭据输入框
/// （连接/测试成功后由服务端回存，下次不再弹）。
class AuthInputDialog extends StatefulWidget {
  final ConnectionModel connection;

  const AuthInputDialog({super.key, required this.connection});

  @override
  State<AuthInputDialog> createState() => _AuthInputDialogState();
}

class _AuthInputDialogState extends State<AuthInputDialog> {
  final _controller = TextEditingController();
  bool _obscureText = true;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isPassword = widget.connection.authType == 'password';

    return AlertDialog(
      title: Text('输入凭据：${widget.connection.name}'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isPassword ? '该连接还没有保存密码，请输入（连接成功后自动保存）：' : '请输入私钥路径：',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _controller,
            obscureText: isPassword && _obscureText,
            autofocus: true,
            textCapitalization: TextCapitalization.none,
            autocorrect: false,
            enableSuggestions: false,
            decoration: InputDecoration(
              labelText: isPassword ? '密码' : '私钥路径',
              border: const OutlineInputBorder(),
              prefixIcon: Icon(isPassword ? Icons.lock : Icons.vpn_key),
              suffixIcon: isPassword
                  ? IconButton(
                      icon: Icon(
                        _obscureText ? Icons.visibility : Icons.visibility_off,
                      ),
                      onPressed: () {
                        setState(() => _obscureText = !_obscureText);
                      },
                    )
                  : null,
            ),
            onSubmitted: (_) => _submit(),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: _submit,
          child: const Text('确定'),
        ),
      ],
    );
  }

  void _submit() {
    final value = _controller.text;
    if (value.isNotEmpty) {
      Navigator.pop(context, value);
    }
  }
}
