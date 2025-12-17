# Connection Management UI Implementation

This document describes the connection management UI implementation for the Flutter frontend.

## Implemented Components

### 1. State Management (`lib/presentation/providers/connection_provider.dart`)

**ConnectionState**: Holds the connection list, loading state, and error messages.

**ConnectionNotifier**: Manages connection CRUD operations:
- `loadConnections()` - Fetches all connections from API
- `addConnection(ConnectionModel)` - Adds a new connection
- `updateConnection(ConnectionModel)` - Updates an existing connection
- `deleteConnection(String id)` - Deletes a connection
- `testConnection(...)` - Tests connection credentials
- `clearError()` - Clears error state

**Provider**:
```dart
final connectionProvider = StateNotifierProvider<ConnectionNotifier, ConnectionState>((ref) {
  return ConnectionNotifier(ref.watch(connectionRepositoryProvider));
});
```

### 2. Connection List Screen (`lib/presentation/screens/connections/connection_list_screen.dart`)

**Features**:
- Displays all saved SSH connections in a card list
- Empty state with helpful message when no connections exist
- Loading indicator during API calls
- Error display with dismissible snackbar
- Refresh button to reload connections
- Floating action button to add new connections

**Connection Card**:
- Shows connection name, host, port, and username
- Displays tags as chips
- Three-dot menu with actions:
  - Test Connection - Validates credentials
  - Edit - Opens edit dialog
  - Delete - Shows confirmation dialog
- Tap card to connect (TODO: navigate to terminal)

**User Interactions**:
- Add connection → Opens form dialog
- Edit connection → Opens pre-filled form dialog
- Delete connection → Shows confirmation dialog
- Test connection → Shows loading, then success/failure message
- Refresh → Reloads connections from API

### 3. Connection Form Dialog (`lib/presentation/screens/connections/widgets/connection_form_dialog.dart`)

**Form Fields**:
- Connection Name (required)
- Host (required)
- Port (required, numeric, 1-65535)
- Username (required)
- Authentication Type (dropdown: Password / Private Key)
- Tags (comma-separated, optional)

**Validation**:
- All required fields validated on submit
- Port must be a valid number (1-65535)
- Empty strings trimmed automatically

**Behavior**:
- Used for both adding and editing connections
- Pre-fills fields when editing
- Returns `ConnectionModel` on save, `null` on cancel

### 4. Router Configuration (`lib/core/router/app_router.dart`)

**Routes**:
- `/` - Home screen (feature showcase)
- `/connections` - Connection list (initial route)
- `/terminal` - Terminal screen (TODO)
- `/files` - File manager screen (TODO)
- `/monitor` - System monitoring screen (TODO)
- `/settings` - Settings screen (TODO)

Uses `go_router` for declarative routing with type-safe navigation.

### 5. Home Screen (`lib/presentation/screens/home_screen.dart`)

Placeholder screen showing SSH Tools branding and feature list.

## Usage Examples

### Loading Connections

The connection list is loaded automatically when the screen is created:

```dart
final connectionState = ref.watch(connectionProvider);

if (connectionState.isLoading) {
  return CircularProgressIndicator();
}

final connections = connectionState.connections;
```

### Adding a Connection

```dart
final notifier = ref.read(connectionProvider.notifier);

final newConnection = ConnectionModel(
  id: '',
  name: 'Production Server',
  host: '192.168.1.100',
  port: 22,
  user: 'admin',
  authType: 'password',
  tags: ['production', 'web'],
);

final success = await notifier.addConnection(newConnection);
```

### Testing a Connection

```dart
final notifier = ref.read(connectionProvider.notifier);

final success = await notifier.testConnection(
  host: '192.168.1.100',
  port: 22,
  user: 'admin',
  authType: 'password',
  authValue: 'password123',
);

if (success) {
  print('Connection successful!');
}
```

## Running the App

Before running, ensure Flutter is installed and the API server is running.

### Prerequisites

1. **Start the API server**:
```bash
cd /Users/dingwei/go/sshTools
./build/bin/sshtools-api
```

2. **Install Flutter dependencies**:
```bash
cd flutter_ui
flutter pub get
```

### Run on Different Platforms

```bash
# Desktop (macOS)
flutter run -d macos

# Mobile simulator
flutter run -d ios        # iOS Simulator
flutter run -d android    # Android Emulator

# Web
flutter run -d chrome

# List available devices
flutter devices
```

## Integration with API Server

The connection UI integrates with the following API endpoints:

- `GET /api/v1/connections` - List all connections
- `POST /api/v1/connections` - Add new connection
- `PUT /api/v1/connections/:id` - Update connection
- `DELETE /api/v1/connections/:id` - Delete connection
- `POST /api/v1/connections/test` - Test connection

**API Configuration**:
- Base URL: `http://localhost:8080`
- Configured in: `lib/core/constants/api_constants.dart`
- HTTP Client: Dio with 30s timeout
- Error Handling: Automatic retry and user-friendly messages

## Architecture Pattern

This implementation follows **Clean Architecture**:

```
Presentation Layer (UI + State)
  ├── Screens (ConnectionListScreen)
  ├── Widgets (ConnectionFormDialog, ConnectionCard)
  └── Providers (ConnectionNotifier)
           ↓
Data Layer (API + Models)
  ├── Repositories (ConnectionRepository)
  ├── DataSources (ApiClient)
  └── Models (ConnectionModel)
```

**Benefits**:
- Separation of concerns
- Testable business logic
- Easy to mock for testing
- Platform-agnostic data layer

## Next Steps

1. **Implement Terminal Screen** - SSH session with xterm.dart
2. **Add Password Input** - Secure credential entry when connecting
3. **Integrate Credential Store** - Save/retrieve passwords securely
4. **Add Connection Groups** - Organize connections by tags
5. **Implement Search/Filter** - Find connections quickly
6. **Add Import/Export** - Backup and restore connections

## Files Created

1. `lib/presentation/providers/connection_provider.dart` - State management
2. `lib/presentation/screens/connections/connection_list_screen.dart` - Main screen
3. `lib/presentation/screens/connections/widgets/connection_form_dialog.dart` - Form dialog
4. `lib/presentation/screens/home_screen.dart` - Home screen
5. `lib/core/router/app_router.dart` - Routing configuration
6. `lib/main.dart` - Updated to use router

## Testing

To test the connection UI:

1. Start the API server: `./build/bin/sshtools-api`
2. Run Flutter app: `flutter run -d macos`
3. Add a test connection with your SSH server details
4. Test the connection to verify API integration
5. Try editing and deleting connections

**Note**: The "Test Connection" feature requires valid SSH credentials. The `authValue` (password/key) is currently empty in the test connection flow - this will be integrated with the credential store in the next phase.

## Known Limitations

1. **No Password Input**: Test connection doesn't prompt for password (needs credential store integration)
2. **No Terminal Navigation**: Tapping a connection shows a snackbar instead of opening terminal
3. **No Connection Status**: No visual indicator for active connections
4. **No Recent Connections**: No quick access to recently used connections
5. **No Connection Sorting**: Connections displayed in API order only

These will be addressed in subsequent implementation phases.
