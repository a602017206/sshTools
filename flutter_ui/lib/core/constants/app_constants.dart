/// Application-wide constants
class AppConstants {
  // App Info
  static const String appName = 'SSH Tools';
  static const String appVersion = '1.0.0';

  // Terminal Defaults
  static const int defaultTerminalCols = 80;
  static const int defaultTerminalRows = 24;

  // SSH Defaults
  static const int defaultSSHPort = 22;
  static const String defaultAuthType = 'password';

  // Storage Keys
  static const String storageKeyTheme = 'theme';
  static const String storageKeyLastConnection = 'last_connection';
  static const String storageKeyAutoConnect = 'auto_connect';

  // UI Constants
  static const double defaultSidebarWidth = 300.0;
  static const double minSidebarWidth = 200.0;
  static const double maxSidebarWidth = 500.0;

  // Timeouts
  static const Duration sshConnectTimeout = Duration(seconds: 30);
  static const Duration fileTransferTimeout = Duration(minutes: 5);

  // Monitoring
  static const Duration monitorRefreshInterval = Duration(seconds: 3);
}
