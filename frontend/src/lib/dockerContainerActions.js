/** Docker 容器行操作纯函数（便于单测）。 */

function statusOf(container) {
  return String(container?.Status ?? container?.status ?? '').trim();
}

/** 状态以 Up 开头视为运行中；Restarting/Paused 视为非「已停止」，只允许停止/重启。 */
function isUp(container) {
  return /^up\b/i.test(statusOf(container));
}

function isTransient(container) {
  return /^restarting\b/i.test(statusOf(container));
}

/** 运行中（含 Paused、Restarting）的容器不可再启动。 */
export function canStart(container) {
  if (!container) return false;
  return !isUp(container) && !isTransient(container);
}

/** 仅运行中或重启中的容器可停止。 */
export function canStop(container) {
  if (!container) return false;
  return isUp(container) || isTransient(container);
}

/** 重启对任意存在的容器均可用。 */
export function canRestart(container) {
  return Boolean(container);
}

/** 删除始终可选，由确认框把关。 */
export function canRemove(container) {
  return Boolean(container);
}

/** 行菜单项：{ id, label, disabled, danger? }。 */
export function containerActionMenu(container) {
  return [
    { id: 'detail', label: '详情', disabled: !container },
    { id: 'start', label: '启动', disabled: !canStart(container) },
    { id: 'stop', label: '停止', disabled: !canStop(container) },
    { id: 'restart', label: '重启', disabled: !canRestart(container) },
    { id: 'remove', label: '删除', disabled: !canRemove(container), danger: true },
  ];
}

/** 删除确认框文案；force 为 true 时强调强制删除。 */
export function removeConfirmCopy(name, force = false) {
  const label = String(name || '').trim() || '该容器';
  if (force) {
    return {
      title: '强制删除容器',
      message: `将强制删除容器「${label}」（运行中的容器会被直接终止），此操作不可恢复。`,
      confirmText: '强制删除',
    };
  }
  return {
    title: '删除容器',
    message: `确定删除容器「${label}」吗？运行中的容器需先停止或勾选强制删除。`,
    confirmText: '删除',
  };
}

/** 删除失败且提示容器仍在运行时，建议二次确认强制删除。 */
export function isStillRunningError(error) {
  const text = String(error?.message || error || '').toLowerCase();
  return /running|仍在运行|cannot remove a running|is running/.test(text);
}

/** 操作成功提示文案。 */
export function actionSuccessLabel(actionId) {
  return { start: '已启动', stop: '已停止', restart: '已重启', remove: '已删除' }[actionId] || '操作完成';
}
