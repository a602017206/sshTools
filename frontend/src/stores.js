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
// 主题状态

export const themeStore = writable('light'); // 'light' | 'dark'

export function setTheme(theme) {
  themeStore.set(theme);
  // 应用主题到 DOM
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }
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
