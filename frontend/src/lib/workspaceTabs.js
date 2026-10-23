export const APP_MODES = [
  { id: 'ssh', label: 'SSH 会话' },
  { id: 'database', label: '数据' },
  { id: 'docker', label: '容器' },
  { id: 'overview', label: '状态墙' }
];

export const SSH_TOOL_TABS = [
  { id: 'files', label: '文件' },
  { id: 'performance', label: '性能' },
  { id: 'logs', label: '会话日志' }
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
  logs: 'ssh',
  database: 'database',
  ssh: 'ssh'
};

export function resolveMode(mode) {
  if (modes.has(mode)) return mode;
  return legacyModeMap[mode] || 'ssh';
}

export function modeForAsset(asset) {
  if (asset?.type === 'database') return 'database';
  if (asset?.type === 'docker') return 'docker';
  return 'ssh';
}

export function isDatabaseSession(session) {
  return session?.type === 'database';
}

export function isDockerSession(session) {
  return session?.type === 'docker';
}

/** SSH / 数据库 / Docker 模式各自只展示对应类型的会话标签 */
export function sessionMatchesMode(session, mode) {
  const resolved = resolveMode(mode);
  if (resolved === 'database') {
    return isDatabaseSession(session);
  }
  if (resolved === 'docker') {
    return isDockerSession(session);
  }
  return !isDatabaseSession(session) && !isDockerSession(session);
}

export function resolveSshToolTab(tab) {
  return sshTools.has(tab) ? tab : 'files';
}

export function sshToolPanelHidden(activeTab, panelId) {
  return resolveSshToolTab(activeTab) !== panelId;
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
  if (mode === 'docker') {
    return {
      title: '容器',
      description: '连接 Docker 引擎后浏览与管理容器。',
      available: true
    };
  }
  return {
    title: 'SSH 会话',
    description: '终端为主舞台，文件、性能与会话日志绑定当前会话。',
    available: true
  };
}
