import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/models/connection_model.dart';
import '../../data/repositories/connection_repository.dart';
import 'api_providers.dart';

/// State for connection list
class ConnectionState {
  final List<ConnectionModel> connections;
  final bool isLoading;
  final String? error;

  const ConnectionState({
    this.connections = const [],
    this.isLoading = false,
    this.error,
  });

  ConnectionState copyWith({
    List<ConnectionModel>? connections,
    bool? isLoading,
    String? error,
  }) {
    return ConnectionState(
      connections: connections ?? this.connections,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

/// Connection list notifier
class ConnectionNotifier extends StateNotifier<ConnectionState> {
  final ConnectionRepository _repository;

  ConnectionNotifier(this._repository) : super(const ConnectionState()) {
    loadConnections();
  }

  /// Load all connections
  Future<void> loadConnections() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final connections = await _repository.getConnections();
      state = state.copyWith(
        connections: connections,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Add a new connection
  Future<bool> addConnection(ConnectionModel connection) async {
    try {
      final newConnection = await _repository.addConnection(connection);
      state = state.copyWith(
        connections: [...state.connections, newConnection],
      );
      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  /// Update an existing connection
  Future<bool> updateConnection(ConnectionModel connection) async {
    try {
      final updatedConnection = await _repository.updateConnection(connection);
      final connections = state.connections.map((c) {
        return c.id == updatedConnection.id ? updatedConnection : c;
      }).toList();
      state = state.copyWith(connections: connections);
      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  /// Delete a connection
  Future<bool> deleteConnection(String id) async {
    try {
      await _repository.deleteConnection(id);
      final connections = state.connections.where((c) => c.id != id).toList();
      state = state.copyWith(connections: connections);
      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  /// Test a connection
  Future<bool> testConnection({
    required String host,
    required int port,
    required String user,
    required String authType,
    required String authValue,
    String? passphrase,
  }) async {
    try {
      return await _repository.testConnection(
        host: host,
        port: port,
        user: user,
        authType: authType,
        authValue: authValue,
        passphrase: passphrase,
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  /// Clear error
  void clearError() {
    state = state.copyWith(error: null);
  }
}

/// Connection provider
final connectionProvider =
    StateNotifierProvider<ConnectionNotifier, ConnectionState>((ref) {
  return ConnectionNotifier(ref.watch(connectionRepositoryProvider));
});
