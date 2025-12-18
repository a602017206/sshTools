import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../data/models/connection_model.dart';
import '../../providers/connection_provider.dart' as conn_provider;
import 'widgets/connection_form_dialog.dart';

/// Connection list screen
class ConnectionListScreen extends ConsumerWidget {
  const ConnectionListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final connectionState = ref.watch(conn_provider.connectionProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('SSH Connections'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              ref.read(conn_provider.connectionProvider.notifier).loadConnections();
            },
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: _buildBody(context, ref, connectionState),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showConnectionDialog(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Add Connection'),
      ),
    );
  }

  Widget _buildBody(
    BuildContext context,
    WidgetRef ref,
    conn_provider.ConnectionState state,
  ) {
    // Show error if present
    if (state.error != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(state.error!),
            backgroundColor: Colors.red,
            action: SnackBarAction(
              label: 'Dismiss',
              textColor: Colors.white,
              onPressed: () {
                ref.read(conn_provider.connectionProvider.notifier).clearError();
              },
            ),
          ),
        );
        ref.read(conn_provider.connectionProvider.notifier).clearError();
      });
    }

    // Show loading
    if (state.isLoading) {
      return const Center(
        child: CircularProgressIndicator(),
      );
    }

    // Show empty state
    if (state.connections.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.storage_outlined,
              size: 80,
              color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.5),
            ),
            const SizedBox(height: 16),
            Text(
              'No Connections',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              'Add your first SSH connection to get started',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withValues(alpha: 0.6),
                  ),
            ),
          ],
        ),
      );
    }

    // Show connection list
    return ListView.builder(
      padding: const EdgeInsets.all(8),
      itemCount: state.connections.length,
      itemBuilder: (context, index) {
        final connection = state.connections[index];
        return _ConnectionCard(
          connection: connection,
          onTap: () => _onConnectionTap(context, ref, connection),
          onEdit: () => _showConnectionDialog(context, ref, connection),
          onDelete: () => _deleteConnection(context, ref, connection),
          onTest: () => _testConnection(context, ref, connection),
        );
      },
    );
  }

  void _onConnectionTap(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) {
    // TODO: Navigate to terminal screen
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Connect to ${connection.name}')),
    );
  }

  Future<void> _showConnectionDialog(
    BuildContext context,
    WidgetRef ref, [
    ConnectionModel? connection,
  ]) async {
    final result = await showDialog<ConnectionModel>(
      context: context,
      builder: (context) => ConnectionFormDialog(connection: connection),
    );

    if (result != null) {
      final notifier = ref.read(conn_provider.connectionProvider.notifier);
      final success = connection == null
          ? await notifier.addConnection(result)
          : await notifier.updateConnection(result);

      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              success
                  ? '${connection == null ? 'Added' : 'Updated'} ${result.name}'
                  : 'Failed to ${connection == null ? 'add' : 'update'} connection',
            ),
            backgroundColor: success ? Colors.green : Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _deleteConnection(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Connection'),
        content: Text('Are you sure you want to delete "${connection.name}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      final success =
          await ref.read(conn_provider.connectionProvider.notifier).deleteConnection(
                connection.id,
              );

      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              success
                  ? 'Deleted ${connection.name}'
                  : 'Failed to delete connection',
            ),
            backgroundColor: success ? Colors.green : Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _testConnection(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) async {
    // Show password/key input dialog
    final authValue = await showDialog<String>(
      context: context,
      builder: (context) => _AuthInputDialog(
        authType: connection.authType,
        connectionName: connection.name,
      ),
    );

    if (authValue == null || authValue.isEmpty) {
      return; // User cancelled
    }

    if (!context.mounted) return;

    // Show loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(
        child: Card(
          child: Padding(
            padding: EdgeInsets.all(24.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                CircularProgressIndicator(),
                SizedBox(height: 16),
                Text('Testing connection...'),
              ],
            ),
          ),
        ),
      ),
    );

    final success =
        await ref.read(conn_provider.connectionProvider.notifier).testConnection(
              host: connection.host,
              port: connection.port,
              user: connection.user,
              authType: connection.authType,
              authValue: authValue,
            );

    if (context.mounted) {
      Navigator.pop(context); // Close loading dialog

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            success
                ? 'Connection successful!'
                : 'Connection failed. Check credentials.',
          ),
          backgroundColor: success ? Colors.green : Colors.red,
        ),
      );
    }
  }
}

/// Connection card widget
class _ConnectionCard extends StatelessWidget {
  final ConnectionModel connection;
  final VoidCallback onTap;
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  final VoidCallback onTest;

  const _ConnectionCard({
    required this.connection,
    required this.onTap,
    required this.onEdit,
    required this.onDelete,
    required this.onTest,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.computer,
                    size: 32,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          connection.name,
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '${connection.user}@${connection.host}:${connection.port}',
                          style:
                              Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    color: Theme.of(context)
                                        .colorScheme
                                        .onSurface
                                        .withValues(alpha: 0.7),
                                  ),
                        ),
                      ],
                    ),
                  ),
                  PopupMenuButton<String>(
                    onSelected: (value) {
                      switch (value) {
                        case 'test':
                          onTest();
                          break;
                        case 'edit':
                          onEdit();
                          break;
                        case 'delete':
                          onDelete();
                          break;
                      }
                    },
                    itemBuilder: (context) => [
                      const PopupMenuItem(
                        value: 'test',
                        child: Row(
                          children: [
                            Icon(Icons.play_arrow),
                            SizedBox(width: 8),
                            Text('Test Connection'),
                          ],
                        ),
                      ),
                      const PopupMenuItem(
                        value: 'edit',
                        child: Row(
                          children: [
                            Icon(Icons.edit),
                            SizedBox(width: 8),
                            Text('Edit'),
                          ],
                        ),
                      ),
                      const PopupMenuItem(
                        value: 'delete',
                        child: Row(
                          children: [
                            Icon(Icons.delete, color: Colors.red),
                            SizedBox(width: 8),
                            Text('Delete', style: TextStyle(color: Colors.red)),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              if (connection.tags != null && connection.tags!.isNotEmpty) ...[
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  children: connection.tags!.map((tag) {
                    return Chip(
                      label: Text(tag),
                      labelStyle: const TextStyle(fontSize: 12),
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      visualDensity: VisualDensity.compact,
                    );
                  }).toList(),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

/// Auth input dialog for testing connections
class _AuthInputDialog extends StatefulWidget {
  final String authType;
  final String connectionName;

  const _AuthInputDialog({
    required this.authType,
    required this.connectionName,
  });

  @override
  State<_AuthInputDialog> createState() => _AuthInputDialogState();
}

class _AuthInputDialogState extends State<_AuthInputDialog> {
  final _controller = TextEditingController();
  bool _obscureText = true;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isPassword = widget.authType == 'password';

    return AlertDialog(
      title: Text('Test Connection: ${widget.connectionName}'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isPassword ? 'Enter password:' : 'Enter private key path:',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _controller,
            obscureText: isPassword && _obscureText,
            autofocus: true,
            decoration: InputDecoration(
              labelText: isPassword ? 'Password' : 'Private Key Path',
              hintText: isPassword ? 'Your password' : '/path/to/private/key',
              border: const OutlineInputBorder(),
              prefixIcon: Icon(isPassword ? Icons.lock : Icons.vpn_key),
              suffixIcon: isPassword
                  ? IconButton(
                      icon: Icon(
                        _obscureText ? Icons.visibility : Icons.visibility_off,
                      ),
                      onPressed: () {
                        setState(() {
                          _obscureText = !_obscureText;
                        });
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
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _submit,
          child: const Text('Test'),
        ),
      ],
    );
  }

  void _submit() {
    final value = _controller.text.trim();
    if (value.isNotEmpty) {
      Navigator.pop(context, value);
    }
  }
}
