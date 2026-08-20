<script>
  import { onMount } from 'svelte';
  import AssetList from './components/AssetList.svelte';
  import TerminalPanel from './components/TerminalPanel.svelte';
  import DevToolsPanel from './components/DevToolsPanel.svelte';
  import UploadTaskDialog from './components/UploadTaskDialog.svelte';
  import AddAssetDialog from './components/AddAssetDialog.svelte';
  import AboutDialog from './components/AboutDialog.svelte';
  import GlobalSettingsDialog from './components/GlobalSettingsDialog.svelte';
  import InputDialog from './components/ui/InputDialog.svelte';
  import ConfirmDialog from './components/ui/ConfirmDialog.svelte';
  import { assetsStore, connectionsStore, activeSessionIdStore, themeStore, uiStore, setSidebarWidth, setRightPanelWidth, setFileManagerHeight, setTheme } from './stores.js';
  import { uploadStore, activeTransfers, completedTransfers } from './stores/uploadStore.js';
  import { formatFileSize, formatSpeed, getTransferPercentage } from './stores/uploadStore.js';
  import { CancelTransfer } from '../wailsjs/go/main/App.js';
  import { WindowToggleMaximise } from '../wailsjs/runtime/runtime.js';
  import { applyAppearanceSettings, getDefaultAppSettings, resolveTheme } from './settings/appearance.js';
  import { isNativeDatabaseType } from './lib/nativeDatabaseTypes.js';
  import { buildJDBCConnectionOptions } from './lib/jdbcConnectionOptions.js';
  import WorkspaceNavigation from './components/WorkspaceNavigation.svelte';
  import SessionToolDock from './components/SessionToolDock.svelte';
  import AIPanel from './components/AIPanel.svelte';
  import DatabaseWorkspaceEmpty from './components/DatabaseWorkspaceEmpty.svelte';
  import { modeForAsset, resolveMode, resolveSshToolTab } from './lib/workspaceTabs.js';
  import { formatConnectionError } from './lib/formatConnectionError.js';
  import { createClonedConnectionFormData } from './lib/connectionFormData.js';
  import { copilotStore } from './stores/copilot.js';

  let isDevToolsOpen = false;
  let isAddDialogOpen = false;
  let isAboutDialogOpen = false;
  let isGlobalSettingsOpen = false;
  let isSidebarCollapsed = false;
  let isRightPanelCollapsed = true;
  let editingAsset = null;
  let cloningAsset = null;
  let connectionDialogRequestVersion = 0;
  let terminalPanelRef;
  let activeMode = 'ssh';
  let sshToolTab = 'files';

  let showDbAuthInput = false;
  let dbAuthInputTitle = '';
  let dbAuthInputMessage = '';
  let dbAuthInputPlaceholder = '';
  let dbAuthInputDefault = '';
  let dbAuthInputType = 'text';
  let dbAuthInputAllowEmpty = false;
  let dbAuthInputTrim = true;
  let resolveDbAuthInput = null;

  let showDbSavePasswordConfirm = false;
  let dbSavePasswordTitle = '';
  let dbSavePasswordMessage = '';
  let dbSavePasswordType = 'warning';
  let resolveDbSavePasswordConfirm = null;
  let showDbErrorDialog = false;
  let dbErrorTitle = '数据库连接失败';
  let dbErrorMessage = '';
  let appSettings = getDefaultAppSettings();
  let settingsDraftSnapshot = null;

  $: connectionsArray = $connectionsStore ? Array.from($connectionsStore.values()) : [];
  $: sshSessions = connectionsArray.filter((session) => session?.type !== 'database');
  $: databaseSessions = connectionsArray.filter((session) => session?.type === 'database');
  $: hasDatabaseSession = databaseSessions.length > 0;
  $: connectedSshSessions = sshSessions.filter((session) => session?.connected);
  $: hasActiveServerSession = connectedSshSessions.length > 0;
  $: boundSshSession =
    connectedSshSessions.find((session) => session.sessionId === $activeSessionIdStore) ||
    connectedSshSessions[connectedSshSessions.length - 1] ||
    null;
  $: boundSessionName = boundSshSession?.connection?.name || boundSshSession?.tabName || '';
  $: showSessionDock = activeMode === 'ssh';
  $: isCopilotOpen = $copilotStore.open;
  $: copilotWidth = $copilotStore.width;
  $: copilotSession = connectionsArray.find((session) => {
    if (!session || session.sessionId !== $activeSessionIdStore) return false;
    return activeMode === 'database' ? session.type === 'database' : session.type !== 'database';
  }) || null;
  $: copilotHasSession = Boolean(copilotSession);
  $: copilotSessionId = copilotSession?.sessionId || null;
  $: themeClass = $themeStore === 'dark' ? 'dark' : '';
  $: isDarkTheme = $themeStore === 'dark';
  $: themeToggleTitle = isDarkTheme ? '切换到亮色模式' : '切换到暗色模式';

  $: if (!hasActiveServerSession && activeMode === 'ssh') {
    isRightPanelCollapsed = true;
  }

  // 从 store 获取面板尺寸
  $: sidebarWidth = $uiStore.sidebarWidth;
  $: rightPanelWidth = $uiStore.rightPanelWidth;
  $: fileManagerHeight = $uiStore.fileManagerHeight;

  // Resize state
  let isResizingSidebar = false;
  let isResizingRightPanel = false;
  let isResizingFileManager = false;
  let isResizingCopilot = false;

  function applyAndSyncSettings(nextSettings) {
    const resolvedTheme = resolveTheme(nextSettings.theme_mode, nextSettings.theme);
    setTheme(resolvedTheme);
    appSettings = {
      ...appSettings,
      ...nextSettings,
      theme: resolvedTheme,
      use_system_theme: nextSettings.theme_mode === 'system'
    };
    applyAppearanceSettings(appSettings);
  }

  async function persistAppSettings(settings) {
    if (!window.wailsBindings || typeof window.wailsBindings.UpdateSettings !== 'function') {
      return;
    }

    const updates = {
      theme: settings.theme,
      theme_mode: settings.theme_mode,
      use_system_theme: settings.use_system_theme,
      font_family: settings.font_family,
      font_size: settings.font_size,
      accent_color: settings.accent_color,
      terminal_font_family: settings.terminal_font_family,
      terminal_font_size: settings.terminal_font_size,
      compact_mode: settings.compact_mode,
      reduced_motion: settings.reduced_motion,
      sidebar_width: $uiStore.sidebarWidth,
      background_image_enabled: Boolean(settings.background_image_enabled),
      background_image_path: settings.background_image_path || '',
      background_image_fit: settings.background_image_fit === 'contain' ? 'contain' : 'cover',
      background_image_opacity: Number(settings.background_image_opacity) || 35,
      copilot_provider: settings.copilot_provider || 'openai_compatible',
      copilot_base_url: settings.copilot_base_url || '',
      copilot_model: settings.copilot_model || ''
    };

    try {
      await window.wailsBindings.UpdateSettings(updates);
    } catch (error) {
      console.error('Failed to update app settings:', error);
    }
  }

  async function loadAppSettings() {
    if (!window.wailsBindings || typeof window.wailsBindings.GetSettings !== 'function') {
      applyAndSyncSettings({
        ...appSettings,
        theme_mode: 'system'
      });
      return;
    }

    try {
      const settings = await window.wailsBindings.GetSettings();
      const merged = {
        ...getDefaultAppSettings(),
        ...settings,
        theme_mode: settings?.theme_mode || (settings?.use_system_theme ? 'system' : settings?.theme || 'dark')
      };
      if (merged.background_image_enabled && merged.background_image_path && typeof window.wailsBindings.GetBackgroundImageDataURL === 'function') {
        try {
          merged.background_image_data_url = await window.wailsBindings.GetBackgroundImageDataURL() || '';
        } catch (error) {
          console.warn('Failed to load background image:', error);
          merged.background_image_data_url = '';
        }
      }
      applyAndSyncSettings(merged);
      if (merged.sidebar_width) {
        setSidebarWidth(merged.sidebar_width);
      }
    } catch (error) {
      console.error('Failed to load app settings:', error);
      applyAndSyncSettings({ ...getDefaultAppSettings(), theme_mode: 'system' });
    }
  }

  async function handleSaveGlobalSettings(nextSettings) {
    const apiKey = typeof nextSettings?.copilot_api_key === 'string'
      ? nextSettings.copilot_api_key.trim()
      : '';
    const settingsWithoutKey = { ...nextSettings };
    delete settingsWithoutKey.copilot_api_key;

    if (apiKey && window.wailsBindings && typeof window.wailsBindings.SetCopilotAPIKey === 'function') {
      try {
        await window.wailsBindings.SetCopilotAPIKey(apiKey);
      } catch (error) {
        console.error('Failed to save copilot API key:', error);
        dbErrorTitle = '设置保存失败';
        dbErrorMessage = '密钥保存失败，请重试。其它设置尚未保存。';
        showDbErrorDialog = true;
        return;
      }
    }

    applyAndSyncSettings(settingsWithoutKey);
    settingsDraftSnapshot = null;
    isGlobalSettingsOpen = false;

    await persistAppSettings(appSettings);
  }

  async function toggleThemeMode() {
    if (isAddDialogOpen) return;
    const nextThemeMode = isDarkTheme ? 'light' : 'dark';
    const nextSettings = {
      ...appSettings,
      theme_mode: nextThemeMode,
      theme: nextThemeMode,
      use_system_theme: false
    };
    applyAndSyncSettings(nextSettings);
    await persistAppSettings(nextSettings);
  }

  function handlePreviewGlobalSettings(nextSettings) {
    const preview = {
      ...appSettings,
      ...nextSettings
    };
    const resolvedTheme = resolveTheme(preview.theme_mode, preview.theme);
    setTheme(resolvedTheme);
    applyAppearanceSettings({
      ...preview,
      theme: resolvedTheme,
      use_system_theme: preview.theme_mode === 'system'
    });
  }

  function openGlobalSettings() {
    settingsDraftSnapshot = { ...appSettings };
    isGlobalSettingsOpen = true;
  }

  function handleCancelGlobalSettings() {
    isGlobalSettingsOpen = false;
    if (settingsDraftSnapshot) {
      applyAndSyncSettings(settingsDraftSnapshot);
      settingsDraftSnapshot = null;
    }
  }

  function toggleDevTools() {
    if (isAddDialogOpen) return;
    isDevToolsOpen = !isDevToolsOpen;
  }

  function ensureSshSessionActive() {
    if (!boundSshSession) return;
    const current = $connectionsStore?.get?.($activeSessionIdStore);
    if (!current || current.type === 'database') {
      activeSessionIdStore.set(boundSshSession.sessionId);
    }
  }

  function selectMode(mode) {
    activeMode = resolveMode(mode);
    if (activeMode === 'database') {
      isRightPanelCollapsed = true;
      return;
    }
    ensureSshSessionActive();
    if (hasActiveServerSession) {
      isRightPanelCollapsed = false;
    }
  }

  function selectSshToolTab(tab) {
    sshToolTab = resolveSshToolTab(tab);
    if (activeMode === 'ssh') {
      ensureSshSessionActive();
      isRightPanelCollapsed = false;
    }
  }

  function openAddConnection() {
    editingAsset = null;
    cloningAsset = null;
    connectionDialogRequestVersion += 1;
    isAddDialogOpen = true;
  }

  function focusResourceTree() {
    isSidebarCollapsed = false;
  }

  function toggleSidebar() {
    if (isAddDialogOpen) return;
    isSidebarCollapsed = !isSidebarCollapsed;
  }

  function toggleRightPanel() {
    if (isAddDialogOpen) return;
    isRightPanelCollapsed = !isRightPanelCollapsed;
  }

  // Sidebar resize handlers
  function startSidebarResize(e) {
    e.preventDefault();
    if (isSidebarCollapsed || isAddDialogOpen) return;
    isResizingSidebar = true;
    document.addEventListener('mousemove', handleSidebarResize);
    document.addEventListener('mouseup', stopSidebarResize);
  }

  function handleSidebarResize(e) {
    if (!isResizingSidebar) return;
    const newWidth = Math.max(200, Math.min(420, e.clientX));
    setSidebarWidth(newWidth);
  }

  function resetSidebarWidth() {
    if (isAddDialogOpen) return;
    setSidebarWidth(260);
  }

  function stopSidebarResize() {
    isResizingSidebar = false;
    document.removeEventListener('mousemove', handleSidebarResize);
    document.removeEventListener('mouseup', stopSidebarResize);
  }

  // Right panel resize handlers
  function startRightPanelResize(e) {
    e.preventDefault();
    if (isAddDialogOpen) return;
    isResizingRightPanel = true;
    document.addEventListener('mousemove', handleRightPanelResize);
    document.addEventListener('mouseup', stopRightPanelResize);
  }

  function handleRightPanelResize(e) {
    if (!isResizingRightPanel) return;
    const containerWidth = window.innerWidth;
    const newWidth = Math.max(300, Math.min(600, containerWidth - e.clientX));
    setRightPanelWidth(newWidth);
  }

  function stopRightPanelResize() {
    isResizingRightPanel = false;
    document.removeEventListener('mousemove', handleRightPanelResize);
    document.removeEventListener('mouseup', stopRightPanelResize);
  }

  function startCopilotResize(e) {
    e.preventDefault();
    if (isAddDialogOpen || !isCopilotOpen) return;
    isResizingCopilot = true;
    document.addEventListener('mousemove', handleCopilotResize);
    document.addEventListener('mouseup', stopCopilotResize);
  }

  function handleCopilotResize(e) {
    if (!isResizingCopilot) return;
    const dockWidth = showSessionDock && !isRightPanelCollapsed ? rightPanelWidth : 0;
    const railWidth = showSessionDock && isRightPanelCollapsed ? 40 : 0;
    const newWidth = Math.max(280, Math.min(520, window.innerWidth - e.clientX - dockWidth - railWidth - 24));
    copilotStore.setWidth(newWidth);
  }

  function stopCopilotResize() {
    isResizingCopilot = false;
    document.removeEventListener('mousemove', handleCopilotResize);
    document.removeEventListener('mouseup', stopCopilotResize);
  }

  // File manager resize handlers
  function startFileManagerResize(e) {
    e.preventDefault();
    if (isAddDialogOpen) return;
    isResizingFileManager = true;
    const rightPanel = document.querySelector('[data-right-panel]');
    if (rightPanel) {
      const rect = rightPanel.getBoundingClientRect();
      fileManagerInitialY = e.clientY;
      fileManagerInitialHeight = fileManagerHeight;
      rightPanelHeight = rect.height;
    }
    document.addEventListener('mousemove', handleFileManagerResize);
    document.addEventListener('mouseup', stopFileManagerResize);
  }

  let fileManagerInitialY = 0;
  let fileManagerInitialHeight = 50;
  let rightPanelHeight = 0;

  function handleFileManagerResize(e) {
    if (!isResizingFileManager) return;
    const deltaY = e.clientY - fileManagerInitialY;
    const deltaPercent = (deltaY / rightPanelHeight) * 100;
    const newHeightPercent = Math.max(20, Math.min(80, fileManagerInitialHeight + deltaPercent));
    setFileManagerHeight(newHeightPercent);
  }

  function stopFileManagerResize() {
    isResizingFileManager = false;
    document.removeEventListener('mousemove', handleFileManagerResize);
    document.removeEventListener('mouseup', stopFileManagerResize);
  }

  // 连接处理 - 检查类型并路由到对应面板
  function handleConnect(asset) {
    activeMode = modeForAsset(asset);

    if (asset.type === 'database') {
      isRightPanelCollapsed = true;
      handleDatabaseConnect({ asset, openPanel: true });
      return;
    }

    activeMode = 'ssh';
    isRightPanelCollapsed = false;
    sshToolTab = 'files';

    if (terminalPanelRef && typeof terminalPanelRef.handleConnect === 'function') {
      terminalPanelRef.handleConnect(asset);
    } else {
      console.error('TerminalPanel not available');
      alert('终端面板未初始化');
    }
  }

  function openDatabaseListPanel(asset, sessionId) {
    activeMode = 'database';
    isRightPanelCollapsed = true;

    if (terminalPanelRef && typeof terminalPanelRef.openDatabaseSession === 'function') {
      terminalPanelRef.openDatabaseSession({ asset, sessionId });
      return;
    }

    console.error('TerminalPanel database session API not available');
    alert('数据库面板未初始化');
  }

  async function handleDatabaseConnect(payload) {
    const asset = payload?.asset || payload;
    const openPanel = payload?.openPanel !== false;

    if (!asset) {
      return;
    }

    activeMode = 'database';
    isRightPanelCollapsed = true;

    if (asset.dbConnected && asset.dbSessionId && openPanel) {
      openDatabaseListPanel(asset, asset.dbSessionId);
      return;
    }

    if (!window.wailsBindings) {
      showDatabaseError('Wails 绑定未加载，请使用 wails dev 运行', '无法连接');
      return;
    }

    try {
      const { ConnectDatabase, ConnectDatabaseWithProfile, ConnectDatabaseWithOptions, TestDatabaseConnection, TestDatabaseConnectionWithOptions, ConnectNativeDatabase, TestNativeDatabaseConnection, HasPassword, GetPassword, SavePassword } = window.wailsBindings;
      const sessionId = asset.dbSessionId || `db-${asset.id}`;
      const host = asset.host;
      const port = asset.port;
      const user = asset.username;
      const dbType = asset.metadata?.db_type || asset.dbType || 'mysql';
      const database = asset.metadata?.database || '';
	  const driverProfileID = asset.metadata?.driver_profile_id || '';
      const jdbcOptions = buildJDBCConnectionOptions(
        dbType,
        database,
        asset.metadata?.oracle_connection_mode,
        asset.metadata?.sqlserver_instance_name
      );

      let password = '';
      try {
        const hasSaved = typeof HasPassword === 'function' && await HasPassword(asset.id);
        if (hasSaved) {
          password = await GetPassword(asset.id);
        } else {
          password = await requestDbInput({
            title: `连接到 ${asset.name}`,
            message: '请输入数据库密码：',
            placeholder: '密码',
            inputType: 'password',
            allowEmpty: false,
            trimValue: false
          });
          if (password === null) {
            return;
          }

          if (password && await requestDbConfirm('保存密码', '是否保存密码以便下次自动连接？', 'warning') && typeof SavePassword === 'function') {
            await SavePassword(asset.id, password);
          }
        }
      } catch (error) {
        console.error('Failed to get saved password:', error);
        password = await requestDbInput({
          title: `连接到 ${asset.name}`,
          message: '请输入数据库密码：',
          placeholder: '密码',
          inputType: 'password',
          allowEmpty: false,
          trimValue: false
        });
        if (password === null) {
          return;
        }
      }

      if (!password) {
        console.log('Database connection cancelled - no password provided');
        return;
      }

      console.log('Connecting to database:', { sessionId, host, port, user, dbType, database });

      const isNativeDatabase = isNativeDatabaseType(dbType);
      if (isNativeDatabase) {
        await TestNativeDatabaseConnection(host, port, user, password, dbType, database);
        await ConnectNativeDatabase(sessionId, host, port, user, password, dbType, database);
      } else {
        if (typeof TestDatabaseConnectionWithOptions === 'function') {
          await TestDatabaseConnectionWithOptions(host, port, user, password, dbType, jdbcOptions.database, jdbcOptions.properties);
        } else if (typeof TestDatabaseConnection === 'function') {
          await TestDatabaseConnection(host, port, user, password, dbType, jdbcOptions.database);
        }
        if (typeof ConnectDatabaseWithOptions === 'function') {
          await ConnectDatabaseWithOptions(sessionId, host, port, user, password, dbType, jdbcOptions.database, driverProfileID, jdbcOptions.properties);
        } else if (typeof ConnectDatabaseWithProfile === 'function') {
          await ConnectDatabaseWithProfile(sessionId, host, port, user, password, dbType, jdbcOptions.database, driverProfileID);
        } else {
          await ConnectDatabase(sessionId, host, port, user, password, dbType, jdbcOptions.database);
        }
      }

      assetsStore.update(items => items.map(item => {
        if (item.id === asset.id) {
          return {
            ...item,
            dbConnected: true,
            dbSessionId: sessionId
          };
        }
        return item;
      }));

      if (openPanel) {
        openDatabaseListPanel(asset, sessionId);
      }
    } catch (error) {
      showDatabaseError(error);
    }
  }

  function requestDbInput({ title, message, placeholder, inputType = 'text', defaultValue = '', allowEmpty = false, trimValue = true }) {
    return new Promise(resolve => {
      dbAuthInputTitle = title;
      dbAuthInputMessage = message;
      dbAuthInputPlaceholder = placeholder;
      dbAuthInputDefault = defaultValue;
      dbAuthInputType = inputType;
      dbAuthInputAllowEmpty = allowEmpty;
      dbAuthInputTrim = trimValue;
      resolveDbAuthInput = resolve;
      showDbAuthInput = true;
    });
  }

  function handleDbAuthInputConfirm(value) {
    showDbAuthInput = false;
    if (resolveDbAuthInput) {
      resolveDbAuthInput(value);
      resolveDbAuthInput = null;
    }
  }

  function handleDbAuthInputCancel() {
    showDbAuthInput = false;
    if (resolveDbAuthInput) {
      resolveDbAuthInput(null);
      resolveDbAuthInput = null;
    }
  }

  function requestDbConfirm(title, message, type = 'warning') {
    return new Promise(resolve => {
      dbSavePasswordTitle = title;
      dbSavePasswordMessage = message;
      dbSavePasswordType = type;
      resolveDbSavePasswordConfirm = resolve;
      showDbSavePasswordConfirm = true;
    });
  }

  function handleDbSavePasswordConfirm() {
    showDbSavePasswordConfirm = false;
    if (resolveDbSavePasswordConfirm) {
      resolveDbSavePasswordConfirm(true);
      resolveDbSavePasswordConfirm = null;
    }
  }

  function handleDbSavePasswordCancel() {
    showDbSavePasswordConfirm = false;
    if (resolveDbSavePasswordConfirm) {
      resolveDbSavePasswordConfirm(false);
      resolveDbSavePasswordConfirm = null;
    }
  }

  function showDatabaseError(error, title = '数据库连接失败') {
    console.error(title, error);
    dbErrorTitle = title;
    dbErrorMessage = formatConnectionError(error, '连接失败，请检查主机、端口、凭据或驱动配置');
    showDbErrorDialog = true;
  }

  function dismissDatabaseError() {
    showDbErrorDialog = false;
  }

  async function handleAddAsset(connectionData) {
    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      return;
    }

    try {
      await window.wailsBindings.AddConnection(connectionData);

      const asset = {
        id: connectionData.id,
        name: connectionData.name,
        host: connectionData.host,
        port: connectionData.port,
        username: connectionData.user,
        group: connectionData.tags?.[0] || '默认分组',
        status: 'online',
        type: connectionData.type || 'ssh',
        auth_type: connectionData.auth_type,
        key_path: connectionData.key_path,
        tags: connectionData.tags || [],
        metadata: connectionData.metadata || {},
        dbType: connectionData.metadata?.db_type,
        dbConnected: false,
        dbSessionId: null
      };

      assetsStore.update(assets => [...assets, asset]);
    } catch (error) {
      console.error('Failed to add asset:', error);
      throw error;
    }
  }

  function handleEditAsset(asset) {
    cloningAsset = null;
    editingAsset = asset;
    connectionDialogRequestVersion += 1;
    isAddDialogOpen = true;
  }

  async function handleCloneAsset(asset) {
    editingAsset = null;
    let credentials = {};
    try {
      const { HasPassword, GetPassword } = window.wailsBindings || {};
      if (typeof HasPassword === 'function' && typeof GetPassword === 'function' && await HasPassword(asset.id)) {
        credentials = { password: await GetPassword(asset.id), savePassword: true };
      }
    } catch (error) {
      console.warn('Failed to load password for cloned connection:', error);
    }
    // 生成独立预填数据，避免后续编辑意外修改资产树；已保存密码会写入新连接的独立凭据记录。
    cloningAsset = createClonedConnectionFormData(asset, new Date(), credentials);
    connectionDialogRequestVersion += 1;
    isAddDialogOpen = true;
  }

  async function handleUpdateAsset(connectionData) {
    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      return;
    }

    try {
      // Update connection is already called in the dialog
      // Just update the local store
      assetsStore.update(assets => {
        return assets.map(asset => {
          if (asset.id === connectionData.id) {
            return {
              ...asset,
              id: connectionData.id,
              name: connectionData.name,
              host: connectionData.host,
              port: connectionData.port,
              username: connectionData.user,
              group: connectionData.tags?.[0] || '默认分组',
              type: connectionData.type || 'ssh',
              auth_type: connectionData.auth_type,
              key_path: connectionData.key_path,
              tags: connectionData.tags || [],
              metadata: connectionData.metadata || {},
              dbType: connectionData.metadata?.db_type,
              dbConnected: false,
              dbSessionId: null
            };
          }
          return asset;
        });
      });
    } catch (error) {
      console.error('Failed to update asset:', error);
      throw error;
    }
  }

  async function loadAssetsFromBackend() {
    if (!window.wailsBindings) {
      console.warn('Wails bindings not loaded yet');
      return;
    }

    try {
      const connections = await window.wailsBindings.GetConnections();
      console.log('Loaded connections from backend:', connections);

      const assets = connections.map(conn => ({
        id: conn.id,
        name: conn.name,
        host: conn.host,
        port: conn.port,
        username: conn.user,
        group: conn.tags?.[0] || '默认分组',
        status: 'online',
        type: conn.type || 'ssh',
        auth_type: conn.auth_type,
        key_path: conn.key_path,
        tags: conn.tags || [],
        metadata: conn.metadata || {},
        dbType: conn.metadata?.db_type,
        dbConnected: false,
        dbSessionId: null
      }));

      assetsStore.set(assets);
    } catch (error) {
      console.error('Failed to load connections:', error);
    }
  }

  function ensureFullHeight() {
    const appElement = document.getElementById('app');
    if (appElement) {
      appElement.style.height = '100vh';
      appElement.style.width = '100vw';
    }
  }

   onMount(async () => {

    let cleanupEvents = null;
    let handleDatabaseConnectEvent = null;
    let handleDatabaseEditEvent = null;

    try {
      const wails = await import('../wailsjs/go/main/App.js');
      // 优先 Wails 运行时注入对象（含全部 Go 导出，包括尚未生成的 Copilot 系列），
      // 再回退到生成的绑定模块，确保设置页等其它调用点也能拿到 Copilot 方法。
      window.wailsBindings = window.go?.main?.App || wails;
      window.dispatchEvent(new CustomEvent('wails-bindings-loaded', {
        detail: Object.keys(wails.default || wails).join(', ')
      }));

      await loadAppSettings();

      await loadAssetsFromBackend();

      handleDatabaseConnectEvent = (event) => {
        handleDatabaseConnect(event.detail);
      };

      handleDatabaseEditEvent = (event) => {
        if (event.detail) {
          handleEditAsset(event.detail);
        }
      };

      window.addEventListener('database:connect', handleDatabaseConnectEvent);
      window.addEventListener('database:edit-connection', handleDatabaseEditEvent);

      // Listen for about dialog event from backend
      const runtime = await import('../wailsjs/runtime/runtime.js');
      cleanupEvents = runtime.EventsOn('app:show-about', () => {
        isAboutDialogOpen = true;
      });

      // Listen for assets changed event (from import)
      window.addEventListener('assets-changed', loadAssetsFromBackend);
    } catch (error) {
      console.warn('Wails bindings not available yet:', error.message);
      applyAndSyncSettings({ ...getDefaultAppSettings(), theme_mode: 'system' });
    }

    ensureFullHeight();
    window.addEventListener('resize', ensureFullHeight);

    return () => {
      window.removeEventListener('resize', ensureFullHeight);
      window.removeEventListener('assets-changed', loadAssetsFromBackend);
      if (handleDatabaseConnectEvent) {
        window.removeEventListener('database:connect', handleDatabaseConnectEvent);
      }
      if (handleDatabaseEditEvent) {
        window.removeEventListener('database:edit-connection', handleDatabaseEditEvent);
      }
      if (cleanupEvents) {
        cleanupEvents();
      }
    };
  });
</script>

<div class="h-screen w-full flex flex-col {themeClass} ops-shell">
  <!-- 顶部标题栏 -->
  <header class="h-12 flex-shrink-0 ops-topbar border-b flex items-center px-4" style="pointer-events: {isAddDialogOpen ? 'none' : 'auto'};">
    <div class="flex items-center gap-3">
      <div class="w-7 h-7 rounded-md flex items-center justify-center font-bold text-xs text-white shadow-sm" style="background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));">
        SSH
      </div>
      <div>
        <div class="font-semibold text-sm header-title" style="color: var(--text-primary);">AHa SSH</div>
        <div class="text-[11px] header-title" style="color: var(--text-secondary);">运维工作台</div>
      </div>
    </div>

    <div class="ml-8 min-w-0 flex-1 flex justify-center">
      <WorkspaceNavigation {activeMode} on:select={(event) => selectMode(event.detail)} />
    </div>

    <div class="ml-4 flex items-center gap-2">
      <div class="relative">
        <button
          on:click={() => uploadStore.togglePanel()}
          class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all"
          title="上传任务"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
        </button>
        {#if $activeTransfers.length > 0}
          <span class="absolute -top-1 -right-1 flex items-center justify-center w-4 h-4 text-[10px] font-bold text-white bg-red-500 rounded-full">
            {$activeTransfers.length}
          </span>
        {/if}
      </div>

      <button
        on:click={toggleThemeMode}
        disabled={isAddDialogOpen}
        class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        title={themeToggleTitle}
        aria-label={themeToggleTitle}
      >
        {#if isDarkTheme}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v2m0 14v2m9-9h-2M5 12H3m15.364-6.364-1.414 1.414M7.05 16.95l-1.414 1.414m12.728 0-1.414-1.414M7.05 7.05 5.636 5.636"></path>
            <circle cx="12" cy="12" r="4" stroke-width="2"></circle>
          </svg>
        {:else}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12.79A8.5 8.5 0 1111.21 3 6.6 6.6 0 0021 12.79z"></path>
          </svg>
        {/if}
      </button>

      <button
        on:click={() => copilotStore.toggle()}
        disabled={isAddDialogOpen}
        class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        class:ops-icon-button--active={isCopilotOpen}
        title="AI Copilot"
        aria-label="AI Copilot"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3l1.2 3.6L17 8l-3.8 1.4L12 13l-1.2-3.6L7 8l3.8-1.4L12 3zM19 14l.7 2 2 .7-2 .7-.7 2-.7-2-2-.7 2-.7.7-2zM5 15l.6 1.6L7.2 17 5.6 17.6 5 19.2l-.6-1.6L2.8 17l1.6-.4L5 15z"></path>
        </svg>
      </button>

      <button
        on:click={openGlobalSettings}
        disabled={isAddDialogOpen}
        class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        title="全局设置"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
        </svg>
      </button>

      <button
        on:click={toggleDevTools}
        disabled={isAddDialogOpen}
        class="flex items-center gap-2 px-3 py-1.5 text-white rounded-md font-medium transition-all shadow-sm hover:brightness-95 focus-visible:outline-none focus-visible:ring-2 disabled:opacity-50 disabled:cursor-not-allowed"
        style="background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
        </svg>
        <span class="text-xs">开发工具</span>
      </button>
      <button on:click={WindowToggleMaximise} disabled={isAddDialogOpen} class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all disabled:opacity-50" title="最大化或还原窗口" aria-label="最大化或还原窗口">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="5" y="5" width="14" height="14" rx="1" stroke-width="2"></rect></svg>
      </button>
    </div>
  </header>

  <!-- 主内容：SSH / 数据库共用壳，TerminalPanel 常驻 -->
  <div class="flex-1 flex overflow-hidden min-h-0 relative">
    {#if isSidebarCollapsed}
    <button
      on:click={toggleSidebar}
      disabled={isAddDialogOpen}
      class="ops-flyout absolute left-0 top-1/2 -translate-y-1/2 z-50 flex items-center justify-center w-8 h-12 rounded-r-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed opacity-0 hover:opacity-100"
      style="color: var(--text-secondary);"
      title="展开资产列表"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
      </svg>
    </button>
    {/if}

    <div
      class="flex-shrink-0 transition-all duration-200 ops-float-panel overflow-hidden"
      class:collapsed={isSidebarCollapsed}
      style="width: {isSidebarCollapsed ? '0' : sidebarWidth}px; min-width: {isSidebarCollapsed ? '0' : sidebarWidth}px; margin: {isSidebarCollapsed ? '0' : '10px 0 10px 10px'};"
    >
      <AssetList
        onConnect={handleConnect}
        onAddClick={openAddConnection}
        onEdit={handleEditAsset}
        onClone={handleCloneAsset}
      />
    </div>

    {#if !isSidebarCollapsed}
    <div
      class="resize-handle-horizontal ops-split-handle ops-split-handle-col flex-shrink-0 relative group"
      role="separator"
      aria-hidden="true"
      style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
      on:mousedown={startSidebarResize}
      on:dblclick={resetSidebarWidth}
      title="拖拽调整宽度，双击复位 260px"
    >
      <button
        on:click={toggleSidebar}
        disabled={isAddDialogOpen}
        class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 ops-icon-button flex items-center justify-center w-7 h-7 rounded-md transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed opacity-0 group-hover:opacity-100"
        title="折叠资产列表"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
      </button>
    </div>
    {/if}

    <div class="flex-1 min-w-0 min-h-0 flex flex-col ops-main-stage relative" style="padding: 10px 8px;">
      <TerminalPanel bind:this={terminalPanelRef} {activeMode} />
      {#if activeMode === 'database' && !hasDatabaseSession}
        <div class="absolute inset-0 z-10 db-empty-overlay">
          <DatabaseWorkspaceEmpty
            onCreateConnection={openAddConnection}
            onFocusSidebar={focusResourceTree}
          />
        </div>
      {/if}
    </div>

    {#if isCopilotOpen}
      <div
        class="resize-handle-horizontal flex-shrink-0 relative group"
        role="separator"
        aria-hidden="true"
        style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; padding: 0 2px; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
        on:mousedown={startCopilotResize}
      >
        <div class="h-full w-full rounded"></div>
      </div>
    {/if}

    <div
      class="flex-shrink-0 flex flex-col overflow-hidden ops-float-panel"
      class:collapsed={!isCopilotOpen}
      style="width: {isCopilotOpen ? copilotWidth : 0}px; min-width: {isCopilotOpen ? '280px' : '0'}; max-width: 520px; margin: {isCopilotOpen ? '10px 8px 10px 0' : '0'};"
    >
      {#if isCopilotOpen}
        <AIPanel
          sessionId={copilotSessionId}
          mode={activeMode === 'database' ? 'database' : 'ssh'}
          hasSession={copilotHasSession}
          onOpenSettings={openGlobalSettings}
          onInsertShell={(id, text) => {
            if (terminalPanelRef && typeof terminalPanelRef.insertCopilotText === 'function') {
              terminalPanelRef.insertCopilotText(id, text);
            }
          }}
        />
      {/if}
    </div>

    {#if showSessionDock}
      {#if !isRightPanelCollapsed}
      <div
        class="resize-handle-horizontal flex-shrink-0 relative group"
        role="separator"
        aria-hidden="true"
        style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; padding: 0 2px; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
        on:mousedown={startRightPanelResize}
      >
        <div class="h-full w-full rounded"></div>
        <button
          on:click={toggleRightPanel}
          disabled={isAddDialogOpen}
          class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 ops-icon-button flex items-center justify-center w-7 h-7 rounded-md transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed opacity-0 group-hover:opacity-100"
          title="折叠会话工具"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
          </svg>
        </button>
      </div>
      {/if}

      <div
        data-right-panel="true"
        class="flex-shrink-0 flex flex-col overflow-hidden ops-float-panel"
        class:collapsed={isRightPanelCollapsed}
        style="width: {isRightPanelCollapsed ? '0' : rightPanelWidth}px; min-width: {isRightPanelCollapsed ? '0' : '300px'}; max-width: 600px; margin: {isRightPanelCollapsed ? '0' : '10px 10px 10px 0'};"
      >
        {#if !isRightPanelCollapsed}
          <SessionToolDock
            activeTab={sshToolTab}
            boundSessionName={boundSessionName}
            hasBoundSession={Boolean(boundSshSession)}
            onSelectTab={selectSshToolTab}
            onConnectHint={focusResourceTree}
          />
        {/if}
      </div>

      {#if isRightPanelCollapsed}
      <div class="ops-rail flex-shrink-0 flex flex-col items-center py-2 gap-2">
        <button
          on:click={() => selectSshToolTab('files')}
          disabled={isAddDialogOpen}
          class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all disabled:opacity-40 disabled:cursor-not-allowed"
          title="文件"
          aria-label="打开文件工具"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7h5l2 2h11v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"></path>
          </svg>
        </button>
        <button
          on:click={() => selectSshToolTab('performance')}
          disabled={isAddDialogOpen}
          class="ops-icon-button flex items-center justify-center w-8 h-8 rounded-md transition-all disabled:opacity-40 disabled:cursor-not-allowed"
          title="性能"
          aria-label="打开性能工具"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 19V5m4 14v-7m4 7V8m4 11v-4m4 4V9"></path>
          </svg>
        </button>
        <div class="mt-auto text-[10px] rotate-90 whitespace-nowrap ops-muted">TOOLS</div>
      </div>
      {/if}
    {/if}
  </div>

  <!-- 对话框 -->
  {#key connectionDialogRequestVersion}
    <AddAssetDialog
      bind:isOpen={isAddDialogOpen}
      bind:editingAsset={editingAsset}
      bind:cloningAsset={cloningAsset}
      dialogRequestVersion={connectionDialogRequestVersion}
      onAdd={handleAddAsset}
      onUpdate={handleUpdateAsset}
    />
  {/key}

  <DevToolsPanel bind:isOpen={isDevToolsOpen} {themeStore} />

  <AboutDialog
    bind:isOpen={isAboutDialogOpen}
    onClose={() => isAboutDialogOpen = false}
    themeStore={themeStore}
  />

  <GlobalSettingsDialog
    bind:isOpen={isGlobalSettingsOpen}
    value={appSettings}
    onPreview={handlePreviewGlobalSettings}
    onSave={handleSaveGlobalSettings}
    onCancel={handleCancelGlobalSettings}
  />

  <InputDialog
    bind:isOpen={showDbAuthInput}
    title={dbAuthInputTitle}
    message={dbAuthInputMessage}
    placeholder={dbAuthInputPlaceholder}
    defaultValue={dbAuthInputDefault}
    inputType={dbAuthInputType}
    allowEmpty={dbAuthInputAllowEmpty}
    trimValue={dbAuthInputTrim}
    confirmText="确定"
    cancelText="取消"
    onConfirm={handleDbAuthInputConfirm}
    onCancel={handleDbAuthInputCancel}
  />

  <ConfirmDialog
    bind:isOpen={showDbSavePasswordConfirm}
    title={dbSavePasswordTitle}
    message={dbSavePasswordMessage}
    type={dbSavePasswordType}
    confirmText="确定"
    cancelText="取消"
    onConfirm={handleDbSavePasswordConfirm}
    onCancel={handleDbSavePasswordCancel}
  />

  <ConfirmDialog
    bind:isOpen={showDbErrorDialog}
    title={dbErrorTitle}
    message={dbErrorMessage}
    type="warning"
    confirmText="知道了"
    cancelText="关闭"
    onConfirm={dismissDatabaseError}
    onCancel={dismissDatabaseError}
  />

  <UploadTaskDialog />
  <!-- 已迁移到 UploadTaskDialog。 -->
  {#if false}
    <div>
      <div>
      <div class="p-4 border-b" style="border-color: var(--glass-border);">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-semibold text-sm">上传任务</h3>
          <button
            on:click={() => uploadStore.setPanelOpen(false)}
            class="ops-icon-button p-1 rounded transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="flex gap-2 p-1 rounded-full" style="border: 1px solid var(--glass-border); background: var(--glass-bg);">
          <button
            on:click={() => uploadStore.setActiveTab('active')}
            class="flex-1 py-1.5 px-3 text-xs font-medium rounded-full transition-colors {$uploadStore.activeTab === 'active'
              ? 'text-white'
              : 'ops-muted'}"
            style={$uploadStore.activeTab === 'active' ? 'background: var(--ops-signal);' : ''}
          >
             进行中 ({$activeTransfers.length})
          </button>
          <button
            on:click={() => uploadStore.setActiveTab('history')}
            class="flex-1 py-1.5 px-3 text-xs font-medium rounded-full transition-colors {$uploadStore.activeTab === 'history'
              ? 'text-white'
              : 'ops-muted'}"
            style={$uploadStore.activeTab === 'history' ? 'background: var(--ops-signal);' : ''}
          >
             历史 ({$completedTransfers.length})
          </button>
        </div>
      </div>

       {#if $uploadStore.activeTab === 'active' && $activeTransfers.length === 0}
        <div class="flex-1 flex flex-col items-center justify-center ops-muted gap-3">
          <svg class="w-12 h-12 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <span class="text-sm">暂无上传任务</span>
        </div>
       {:else if $uploadStore.activeTab === 'active'}
         <div class="flex-1 overflow-y-auto p-4 space-y-3">
           {#each $activeTransfers as transfer (transfer.id)}
            <div class="rounded-xl border p-3" style="background: var(--glass-bg); border-color: var(--glass-border);">
              <div class="flex items-center justify-between mb-2">
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium truncate" title={transfer.filename}>{transfer.filename}</div>
                  <div class="text-xs ops-muted mt-0.5">
                    {formatFileSize(transfer.bytesSent)} / {formatFileSize(transfer.totalBytes)}
                    {#if transfer.speed}
                      • {formatSpeed(transfer.speed)}
                    {/if}
                  </div>
                </div>
                <div class="flex items-center gap-2 ml-3">
                  <span class="text-xs font-medium" style="color: var(--ops-signal);">
                    {Math.round(getTransferPercentage(transfer))}%
                  </span>
                  <button
                    on:click={async () => {
                      await CancelTransfer(transfer.id);
                      uploadStore.cancelTransfer(transfer.id);
                    }}
                    class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-red-500 transition-colors"
                    title="取消上传"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
              <div class="h-2 rounded-full overflow-hidden" style="background: color-mix(in srgb, var(--glass-bg) 60%, #94a3b8);">
                <div
                  class="h-full transition-all duration-300"
                  style={`width: ${Math.min(100, Math.max(0, getTransferPercentage(transfer)))}%; background: var(--ops-signal);`}
                ></div>
              </div>
            </div>
          {/each}
        </div>
       {:else if $uploadStore.activeTab === 'history' && $completedTransfers.length === 0}
        <div class="flex-1 flex flex-col items-center justify-center ops-muted gap-3">
          <svg class="w-12 h-12 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm">暂无历史记录</span>
        </div>
       {:else if $uploadStore.activeTab === 'history'}
         <div class="flex-1 flex flex-col">
           <div class="p-3 border-b" style="border-color: var(--glass-border);">
             <button
               on:click={() => uploadStore.clearCompleted()}
               class="w-full py-2 px-3 text-xs font-medium rounded-lg transition-colors text-red-600 dark:text-red-400"
               style="background: color-mix(in srgb, var(--ops-alert) 8%, transparent);"
             >
               清空历史记录
             </button>
           </div>
           <div class="flex-1 overflow-y-auto p-4 space-y-3">
             {#each $completedTransfers as transfer (transfer.id)}
              <div class="rounded-xl border p-3" style="background: var(--glass-bg); border-color: var(--glass-border);">
                <div class="flex items-center justify-between">
                  <div class="flex-1 min-w-0">
                    <div class="text-sm font-medium truncate" title={transfer.filename}>{transfer.filename}</div>
                    <div class="text-xs ops-muted mt-0.5 flex items-center gap-2">
                      {#if transfer.status === 'completed'}
                        <span class="text-green-500">完成</span>
                      {:else if transfer.status === 'failed'}
                        <span class="text-red-500">失败</span>
                      {:else if transfer.status === 'cancelled'}
                        <span>已取消</span>
                      {/if}
                      <span>•</span>
                      <span>{formatFileSize(transfer.totalBytes)}</span>
                    </div>
                    {#if transfer.status === 'failed' && transfer.error}
                      <div class="text-xs text-red-500 dark:text-red-400 mt-1 truncate" title={transfer.error}>
                        {transfer.error}
                      </div>
                    {/if}
                  </div>
                  <button
                    on:click={() => uploadStore.removeTransfer(transfer.id)}
                    class="ops-icon-button p-1.5 rounded transition-colors ml-3"
                    title="删除记录"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    overflow: hidden;
    height: 100vh;
  }

  :global(html) {
    height: 100%;
  }

  .db-empty-overlay {
    /* 压低背后实心面板观感，露出壳层氛围色 */
    background: color-mix(in srgb, var(--bg-primary) 18%, transparent);
    backdrop-filter: blur(2px);
    -webkit-backdrop-filter: blur(2px);
  }

  .resize-handle-horizontal {
    width: auto;
    height: 100%;
    transition: background-color 0.2s ease;
    z-index: 100;
    background-color: transparent;
    position: relative;
  }

  .resize-handle-horizontal > div {
    width: 1px;
    height: 100%;
    background-color: transparent;
    border-radius: 0;
    transition: background-color 0.2s ease, width 0.2s ease;
    opacity: 0;
  }

  .resize-handle-horizontal:hover > div {
    background-color: var(--border-primary);
    width: 2px;
    opacity: 1;
  }

  .resize-handle-vertical {
    height: auto;
    width: 100%;
    transition: background-color 0.2s ease;
    z-index: 100;
    background-color: transparent;
    position: relative;
  }

  .resize-handle-vertical > div {
    height: 1px;
    width: 100%;
    background-color: transparent;
    border-radius: 0;
    transition: background-color 0.2s ease, height 0.2s ease;
    opacity: 0;
  }

  .resize-handle-vertical:hover > div {
    background-color: var(--border-primary);
    height: 2px;
    opacity: 1;
  }

  @media (max-width: 768px) {
    .header-title {
      display: none;
    }
  }
</style>
