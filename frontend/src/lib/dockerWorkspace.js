/** Docker 工作区纯函数（便于单测）。 */

const TRANSPORT_LABELS = { local: '本机', ssh: 'SSH', tcp: 'TCP' };

export function dockerTransportOf(asset) {
  const metadata = asset?.metadata || {};
  const explicit = String(metadata.transport || '').toLowerCase();
  if (explicit) return explicit;
  if (metadata.ssh_connection_id) return 'ssh';
  if (asset?.host) return 'tcp';
  return 'local';
}

export function dockerTransportLabel(asset) {
  const transport = dockerTransportOf(asset);
  return TRANSPORT_LABELS[transport] || transport;
}

/** Docker 状态文案以 "Up" 开头视为运行中（API 与 CLI 回退一致）。 */
export function isContainerRunning(container) {
  return /^up\b/i.test(String(container?.Status || container?.status || '').trim());
}

export function containerField(container, key) {
  if (!container) return '';
  return container[key] ?? container[key.toLowerCase()] ?? '';
}

export function formatContainerCreated(seconds) {
  const value = Number(seconds);
  if (!value || value <= 0) return '-';
  const d = new Date(value * 1000);
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** 后端若返回 via: cli 的容器行，则提示 CLI 回退。 */
export function usedCliFallback(containers) {
  return (containers || []).some((item) => String(item?.via || item?.Via || '').toLowerCase() === 'cli');
}

export function visibleContainers(containers, onlyRunning) {
  const list = containers || [];
  return onlyRunning ? list.filter(isContainerRunning) : list;
}

export const DEFAULT_DOCKER_DETAIL_WIDTH = 420;
export const MIN_DOCKER_DETAIL_WIDTH = 320;
export const MAX_DOCKER_DETAIL_WIDTH = 720;

/** 将详情侧栏宽度夹到规格范围，并可选按主区宽度限制为约 55%。 */
export function clampDockerDetailWidth(width, mainWidth) {
  const raw = Number(width);
  const base = Number.isFinite(raw) ? raw : DEFAULT_DOCKER_DETAIL_WIDTH;
  let next = Math.min(MAX_DOCKER_DETAIL_WIDTH, Math.max(MIN_DOCKER_DETAIL_WIDTH, base));
  const main = Number(mainWidth);
  if (Number.isFinite(main) && main > 0) {
    const cap = Math.max(MIN_DOCKER_DETAIL_WIDTH, Math.floor(main * 0.55));
    next = Math.min(next, Math.min(MAX_DOCKER_DETAIL_WIDTH, cap));
    next = Math.max(MIN_DOCKER_DETAIL_WIDTH, next);
  }
  return next;
}
