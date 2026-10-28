/** Docker 镜像/网络/卷资源标签页与删除文案纯函数（便于单测）。 */

export const DOCKER_RESOURCE_TABS = [
  { id: 'containers', label: '容器' },
  { id: 'images', label: '镜像' },
  { id: 'networks', label: '网络' },
  { id: 'volumes', label: '卷' },
];

const KIND_LABELS = {
  containers: '容器',
  images: '镜像',
  networks: '网络',
  volumes: '卷',
};

/** 字节数格式化为 B/KB/MB/GB；无效值返回 '-'。 */
export function formatDockerSize(bytes) {
  if (bytes === null || bytes === undefined || bytes === '') return '-';
  const n = Number(bytes);
  if (!Number.isFinite(n) || n < 0) return '-';
  if (n < 1024) return `${Math.round(n)} B`;
  const units = ['KB', 'MB', 'GB'];
  let value = n / 1024;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024;
    i += 1;
  }
  return `${value.toFixed(1)} ${units[i]}`;
}

function realTags(image) {
  const raw = image?.Tags ?? image?.RepoTags ?? image?.tags ?? [];
  if (!Array.isArray(raw)) return [];
  return raw.map((t) => String(t ?? '').trim()).filter((t) => t && t !== '<none>:<none>' && t !== '<none>');
}

/** 镜像标签展示串；无有效标签时为 '<none>'。 */
export function imageDisplayTags(image) {
  const tags = realTags(image);
  return tags.length ? tags.join(', ') : '<none>';
}

/** 列表行 key：images ID|首个标签；networks ID|Name；volumes Name；其余用 ID。 */
export function resourceRowKey(kind, row) {
  if (!row) return '';
  switch (kind) {
    case 'images':
      return `${row.ID ?? ''}|${realTags(row)[0] ?? ''}`;
    case 'networks':
      return `${row.ID ?? ''}|${row.Name ?? ''}`;
    case 'volumes':
      return String(row.Name ?? '');
    default:
      return String(row.ID ?? row.Id ?? '');
  }
}

/** 删除确认框中文文案。 */
export function removeResourceConfirmCopy(kind, label, force) {
  const noun = KIND_LABELS[kind] || '资源';
  const name = String(label ?? '').trim() || '(未命名)';
  const verb = force ? '强制删除' : '删除';
  let message = `确定要${verb}${noun}「${name}」吗？此操作不可撤销。`;
  if (force) {
    message = `确定要强制删除${noun}「${name}」吗？强制删除可能影响正在使用它的资源，且不可撤销。`;
  } else if (kind === 'volumes') {
    message = `确定要删除卷「${name}」吗？卷中的数据将永久丢失，且不可撤销。`;
  }
  return {
    title: `${verb}${noun}`,
    message,
    confirmLabel: verb,
    cancelLabel: '取消',
  };
}

function yesNo(v) {
  return v ? '是' : '否';
}

function pushField(fields, label, value) {
  if (value === null || value === undefined) return;
  const text = String(value).trim();
  if (!text) return;
  fields.push({ label, value: text });
}

function formatCreated(ts) {
  const n = Number(ts);
  if (!Number.isFinite(n) || n <= 0) return '';
  return new Date(n * 1000).toLocaleString('zh-CN');
}

/** 侧栏详情字段；空值跳过，未知类型返回空数组。 */
export function resourceDetailFields(kind, detail) {
  if (!detail) return [];
  const fields = [];
  switch (kind) {
    case 'images':
      pushField(fields, 'ID', detail.ID);
      pushField(fields, '标签', imageDisplayTags(detail));
      pushField(fields, '大小', formatDockerSize(detail.Size) === '-' ? '' : formatDockerSize(detail.Size));
      pushField(fields, '创建时间', formatCreated(detail.Created));
      pushField(fields, '架构', detail.Architecture);
      pushField(fields, '系统', detail.Os);
      pushField(fields, '作者', detail.Author);
      pushField(fields, '备注', detail.Comment);
      break;
    case 'networks':
      pushField(fields, 'ID', detail.ID);
      pushField(fields, '名称', detail.Name);
      pushField(fields, '驱动', detail.Driver);
      pushField(fields, '范围', detail.Scope);
      if ('Internal' in detail) pushField(fields, '内部网络', yesNo(detail.Internal));
      if ('Attachable' in detail) pushField(fields, '可挂载', yesNo(detail.Attachable));
      if ('Ingress' in detail) pushField(fields, 'Ingress', yesNo(detail.Ingress));
      pushField(fields, 'IPAM', detail.IPAM);
      break;
    case 'volumes': {
      pushField(fields, '名称', detail.Name);
      pushField(fields, '驱动', detail.Driver);
      pushField(fields, '挂载点', detail.Mountpoint);
      pushField(fields, '范围', detail.Scope);
      pushField(fields, '创建时间', formatCreated(detail.Created));
      const labels = detail.Labels && typeof detail.Labels === 'object' ? detail.Labels : {};
      pushField(
        fields,
        '标签',
        Object.keys(labels)
          .sort()
          .map((k) => `${k}=${labels[k]}`)
          .join(', '),
      );
      break;
    }
    default:
      return [];
  }
  return fields;
}
