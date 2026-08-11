export const APP_MODES = [
  { id: 'ssh', label: 'SSH 会话' },
  { id: 'database', label: '数据库' }
];

export const SSH_TOOL_TABS = [
  { id: 'files', label: '文件' },
  { id: 'performance', label: '性能' }
];

/** @deprecated 旧七标签仅作兼容映射，顶栏不再展示 */
export const WORKSPACE_TABS = APP_MODES;

const modes = new Set(APP_MODES.map((mode) => mode.id));
const sshTools = new Set(SSH_TOOL_TABS.map((tab) => tab.id));

const legacyModeMap = {
  dashboard: 'ssh',
  terminal: 'ssh',
  files: 'ssh',
  performance: 'ssh',
  docker: 'ssh',
  logs: 'ssh',
  database: 'database',
  ssh: 'ssh'
};

export function resolveMode(mode) {
  if (modes.has(mode)) return mode;
  return legacyModeMap[mode] || 'ssh';
}

export function modeForAsset(asset) {
  return asset?.type === 'database' ? 'database' : 'ssh';
}

export function isDatabaseSession(session) {
  return session?.type === 'database';
}

/** SSH 模式展示终端/本地会话；数据库模式只展示数据库会话 */
export function sessionMatchesMode(session, mode) {
  const resolved = resolveMode(mode);
  if (resolved === 'database') {
    return isDatabaseSession(session);
  }
  return !isDatabaseSession(session);
}

export function resolveSshToolTab(tab) {
  return sshTools.has(tab) ? tab : 'files';
}

/** 兼容旧调用：返回双模式 id */
export function resolveWorkspace(workspace) {
  return resolveMode(workspace);
}

export function getWorkspaceMeta(workspace) {
  const mode = resolveMode(workspace);
  if (mode === 'database') {
    return {
      title: '数据库',
      description: '连接实例后浏览对象、设计表并执行查询。',
      available: true
    };
  }
  return {
    title: 'SSH 会话',
    description: '终端为主舞台，文件与性能绑定当前会话。',
    available: true
  };
}
