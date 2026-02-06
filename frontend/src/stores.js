import { writable, derived } from 'svelte/store';

export const assetsStore = writable([]);

// 按分组分组的资产
export const groupedAssetsStore = derived(assetsStore, ($assets) => {
  return $assets.reduce((acc, asset) => {
    if (!acc[asset.group]) {
      acc[asset.group] = [];
    }
    acc[asset.group].push(asset);
    return acc;
  }, {});
});

// ==================== Connections Store ====================
// SSH 连接会话状态

export const connectionsStore = writable(new Map());
export const activeSessionIdStore = writable(null);

// ==================== Theme Store ====================

const THEME_STORAGE_KEY = 'ssh-tools-theme';

function loadSavedTheme() {
  if (typeof window === 'undefined') return null;
  try {
    return localStorage.getItem(THEME_STORAGE_KEY);
  } catch {
    return null;
  }
}

function saveTheme(theme) {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
  }
}

const savedTheme = loadSavedTheme();
const initialTheme = savedTheme || 'light';

export const themeStore = writable(initialTheme);

export function setTheme(theme) {
  themeStore.set(theme);
  saveTheme(theme);
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }
}

export function toggleTheme() {
  themeStore.update(current => {
    const newTheme = current === 'light' ? 'dark' : 'light';
    saveTheme(newTheme);
    if (typeof document !== 'undefined') {
      document.documentElement.classList.toggle('dark', newTheme === 'dark');
    }
    return newTheme;
  });
}

// ==================== UI Store ====================
// UI 状态

// 从 localStorage 加载配置
function loadUIConfig() {
  if (typeof window === 'undefined') return null;
  try {
    const saved = localStorage.getItem('ssh-tools-ui-config');
    if (!saved) return null;

    const config = JSON.parse(saved);

    // 确保数字类型正确（JSON.parse 会将数字转换为字符串）
    if (config.sidebarWidth) config.sidebarWidth = Number(config.sidebarWidth);
    if (config.rightPanelWidth) config.rightPanelWidth = Number(config.rightPanelWidth);
    if (config.fileManagerHeight) config.fileManagerHeight = Number(config.fileManagerHeight);

    return config;
  } catch {
    return null;
  }
}

const defaultUIConfig = {
  isDevToolsOpen: false,
  isFileManagerOpen: false,
  isMonitorOpen: false,
  sidebarWidth: 288,
  rightPanelWidth: 320,
  fileManagerHeight: 50,
};

const savedConfig = loadUIConfig();

export const uiStore = writable({
  ...defaultUIConfig,
  ...savedConfig,
});

// 保存配置到 localStorage
uiStore.subscribe($ui => {
  if (typeof window !== 'undefined') {
    try {
      localStorage.setItem('ssh-tools-ui-config', JSON.stringify($ui));
    } catch {
      // 忽略存储错误
    }
  }
});

export const isDevToolsOpenStore = derived(uiStore, ($ui) => $ui.isDevToolsOpen);
export const isFileManagerOpenStore = derived(uiStore, ($ui) => $ui.isFileManagerOpen);
export const isMonitorOpenStore = derived(uiStore, ($ui) => $ui.isMonitorOpen);
export const sidebarWidthStore = derived(uiStore, ($ui) => $ui.sidebarWidth);
export const rightPanelWidthStore = derived(uiStore, ($ui) => $ui.rightPanelWidth);
export const fileManagerHeightStore = derived(uiStore, ($ui) => $ui.fileManagerHeight);

// UI 动作
export function toggleDevTools() {
  uiStore.update($ui => ({ ...$ui, isDevToolsOpen: !$ui.isDevToolsOpen }));
}

export function toggleFileManager() {
  uiStore.update($ui => ({ ...$ui, isFileManagerOpen: !$ui.isFileManagerOpen }));
}

export function toggleMonitor() {
  uiStore.update($ui => ({ ...$ui, isMonitorOpen: !$ui.isMonitorOpen }));
}

export function setSidebarWidth(width) {
  uiStore.update($ui => ({ ...$ui, sidebarWidth: width }));
}

export function setRightPanelWidth(width) {
  uiStore.update($ui => ({ ...$ui, rightPanelWidth: width }));
}

export function setFileManagerHeight(height) {
  uiStore.update($ui => ({ ...$ui, fileManagerHeight: height }));
}

// ==================== File Manager Config Store ====================
// 文件管理器配置存储（通过后端持久化到 ~/.ahasshtools/config.json）

const FILEMANAGER_CONFIG_KEY = 'ssh-tools-filemanager-config-temp';

// 文件管理器配置（临时缓存，实际值从后端加载）
const defaultFileManagerConfig = {
  directoryTracking: false,
  historyEnabled:    true,
  historyLimit:     5,
  history:          [],
};

export const fileManagerConfigStore = writable({
  ...defaultFileManagerConfig,
});

// 辅助函数：获取当前会话的配置
export async function loadFileManagerConfig(connectionId) {
  if (typeof window === 'undefined' || !window.wailsBindings) return null;

  try {
    const { GetFileManagerSettings } = window.wailsBindings;
    if (typeof GetFileManagerSettings !== 'function') return null;

    const config = await GetFileManagerSettings(connectionId);
    // Map snake_case backend response to camelCase frontend state
    const mappedConfig = {
      directoryTracking: config?.directory_tracking ?? false,
      historyEnabled: config?.history_enabled ?? true,
      historyLimit: config?.history_limit ?? 5,
      history: config?.history ?? [],
    };
    fileManagerConfigStore.set(mappedConfig);
    return mappedConfig;
  } catch (error) {
    console.error('Failed to load file manager config:', error);
    return null;
  }
}

// 辅助函数：更新文件管理器配置
export async function updateFileManagerConfig(connectionId, settings) {
  if (typeof window === 'undefined' || !window.wailsBindings) return;
  
  try {
    const { UpdateFileManagerSettings } = window.wailsBindings;
    if (typeof UpdateFileManagerSettings !== 'function') return;
    
    await UpdateFileManagerSettings(connectionId, settings);
    
    fileManagerConfigStore.update($config => ({ ...$config, ...settings }));
  } catch (error) {
    console.error('Failed to update file manager config:', error);
  }
}

// 辅助函数：获取当前活动的连接 ID
export function getActiveConnectionId() {
  if (typeof window === 'undefined') return null;
  
  // 从 connectionsStore 获取当前会话的连接信息
  const activeSessionId = window.activeSessionIdStoreValue;
  if (!activeSessionId) return null;
  
  const connectionsStore = window.connectionsStoreValue;
  if (!connectionsStore || !connectionsStore.has) return null;
  
  const session = connectionsStore.get(activeSessionId);
  if (!session) return null;
  
  return session.connection?.id || null;
}
