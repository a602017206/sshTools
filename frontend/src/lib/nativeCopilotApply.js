import { parseNativeMutationArtifact, parseNativeResourceContent } from './nativeDatabaseOperations.js';

/** 把 AI artifact 填入 ES Discover / Dev Tools / 文档编辑器。 */
export function applyNativeArtifactToElasticsearch(artifact, state = {}) {
  const type = String(artifact?.type || '').toLowerCase();
  const content = String(artifact?.content || '').trim();
  if (!content) {
    return { ...state, errorMessage: '无可填入内容' };
  }

  if (type === 'native_mutation') {
    const mutation = parseNativeMutationArtifact(content);
    const next = {
      ...state,
      selectedResource: mutation.name || state.selectedResource,
      activeTab: 'discover',
      actionMessage: artifact.summary || '已填入变更提案，请确认后执行'
    };
    if (mutation.operation === 'index_document' || mutation.operation === 'update_document') {
      const payload = parseNativeResourceContent(mutation.payload);
      next.esDocumentId = payload.id || '';
      next.esDocumentBody = JSON.stringify(payload.document ?? payload, null, 2);
    } else if (mutation.operation === 'create_index') {
      next.showCreateIndex = true;
      next.newIndexName = mutation.name || '';
      next.newIndexBody = mutation.payload || '{}';
    }
    return next;
  }

  const parsed = parseNativeResourceContent(content);
  if (parsed.method && parsed.path) {
    return {
      ...state,
      activeTab: 'devtools',
      devMethod: String(parsed.method || 'GET').toUpperCase(),
      devPath: String(parsed.path || ''),
      devBody: parsed.body != null ? JSON.stringify(parsed.body, null, 2) : '',
      actionMessage: artifact.summary || '已填入 Dev Tools'
    };
  }

  let queryText = content;
  try {
    queryText = JSON.stringify(JSON.parse(content), null, 2);
  } catch {
    // keep raw
  }
  return {
    ...state,
    activeTab: 'discover',
    queryText,
    actionMessage: artifact.summary || '已填入查询 DSL'
  };
}

export function applyNativeArtifactToRedis(artifact, state = {}) {
  const type = String(artifact?.type || '').toLowerCase();
  const content = String(artifact?.content || '').trim();
  if (!content) {
    return { ...state, errorMessage: '无可填入内容' };
  }

  if (type === 'native_mutation') {
    const mutation = parseNativeMutationArtifact(content);
    return {
      ...state,
      selectedRedisDb: mutation.parent || state.selectedRedisDb,
      selectedResource: mutation.name || state.selectedResource,
      activeTab: 'key',
      applyTarget: 'key',
      actionMessage: artifact.summary || '已填入变更提案，请确认后执行'
    };
  }

  const target = resolveRedisNativeApplyTarget(content, artifact);
  if (target.kind === 'match') {
    return {
      ...state,
      activeTab: 'key',
      applyTarget: 'match',
      keyPattern: target.keyPattern,
      shouldReloadKeys: true,
      cliCommand: target.cliCommand || state.cliCommand,
      actionMessage: artifact.summary || `已填入 MATCH：${target.keyPattern}`
    };
  }

  return {
    ...state,
    activeTab: 'cli',
    applyTarget: 'cli',
    shouldReloadKeys: false,
    cliCommand: target.cliCommand || normalizeRedisCLICommand(content),
    actionMessage: artifact.summary || '已填入 CLI 命令'
  };
}

/**
 * 根据 artifact 内容判断填入 MATCH 工具栏还是 CLI。
 * - SCAN + MATCH / 纯通配模式 → MATCH（键浏览）
 * - GET/SET 等其它命令 → CLI
 * - 可用 content.target 或 artifact.target 强制指定
 */
export function resolveRedisNativeApplyTarget(content, artifact = {}) {
  const text = String(content || '').trim();
  const explicit = String(artifact?.target || '').trim().toLowerCase();
  let parsed = {};
  if (text.startsWith('{')) {
    parsed = parseNativeResourceContent(text);
  }
  const explicitFromContent = String(parsed.target || '').trim().toLowerCase();
  const forced = explicit || explicitFromContent;

  const matchPattern = extractRedisMatchPattern(text, parsed);
  const cliCommand = normalizeRedisCLICommand(text);
  const primary = primaryRedisCommand(text, parsed);

  if (forced === 'cli') {
    return { kind: 'cli', cliCommand };
  }
  if (forced === 'match') {
    return {
      kind: 'match',
      keyPattern: matchPattern || (looksLikeRedisKeyPattern(text) ? text : '*'),
      cliCommand
    };
  }

  if (matchPattern != null && (!primary || primary === 'SCAN')) {
    return { kind: 'match', keyPattern: matchPattern, cliCommand };
  }
  if (looksLikeRedisKeyPattern(text)) {
    return { kind: 'match', keyPattern: text, cliCommand: '' };
  }
  return { kind: 'cli', cliCommand };
}

/** 从 SCAN 结构化内容提取 MATCH 模式，保持原始大小写。 */
export function extractRedisMatchPattern(content, parsedInput) {
  const text = String(content || '').trim();
  let parsed = parsedInput;
  if (!parsed) {
    if (text.startsWith('{')) parsed = parseNativeResourceContent(text);
    else if (text.startsWith('[')) {
      try {
        const arr = JSON.parse(text);
        if (Array.isArray(arr)) return extractMatchFromTokens(arr);
      } catch {
        // fall through
      }
      parsed = {};
    } else {
      parsed = parseNativeResourceContent(text);
    }
  }

  if (parsed && parsed.match != null && String(parsed.match).trim() !== '') {
    return String(parsed.match);
  }

  if (parsed && Array.isArray(parsed.command)) {
    return extractMatchFromTokens(parsed.command);
  }

  const fromCli = extractMatchFromTokens(normalizeRedisCLICommand(text).split(/\s+/));
  if (fromCli != null) return fromCli;

  return null;
}

function extractMatchFromTokens(tokens) {
  const list = (tokens || []).map((item) => String(item));
  for (let i = 0; i < list.length; i += 1) {
    if (list[i].toUpperCase() === 'MATCH' && i + 1 < list.length) {
      return list[i + 1];
    }
  }
  return null;
}

function primaryRedisCommand(content, parsed) {
  if (parsed && Array.isArray(parsed.command) && parsed.command.length) {
    return String(parsed.command[0] || '').trim().toUpperCase();
  }
  if (parsed && parsed.command != null && !Array.isArray(parsed.command)) {
    const cmd = String(parsed.command || '').trim();
    if (cmd) return cmd.split(/\s+/)[0].toUpperCase();
  }
  const normalized = normalizeRedisCLICommand(content);
  const head = normalized.split(/\s+/)[0] || '';
  return head.toUpperCase();
}

function looksLikeRedisKeyPattern(text) {
  const value = String(text || '').trim();
  if (!value || /\s/.test(value)) return false;
  if (value.startsWith('{') || value.startsWith('[')) return false;
  const upper = value.toUpperCase();
  if (/^[A-Z][A-Z0-9]*$/.test(upper) && upper.length <= 12) {
    // PING / GET / SCAN 等单词命令不算 pattern
    return false;
  }
  return true;
}

/** 把 AI 产出的 JSON（如 SCAN+match）转成可直接执行的 Redis 命令行。 */
export function normalizeRedisCLICommand(content) {
  const text = String(content || '').trim();
  if (!text) return '';

  if (text.startsWith('[')) {
    try {
      const arr = JSON.parse(text);
      if (Array.isArray(arr) && arr.length) {
        return arr.map((item) => String(item)).join(' ');
      }
    } catch {
      // fall through
    }
  }

  if (!text.startsWith('{')) {
    const parsedLoose = parseNativeResourceContent(text);
    if (parsedLoose.command != null) {
      return formatRedisStructuredCommand(parsedLoose);
    }
    return normalizeCommaJoinedRedisTokens(text);
  }

  const parsed = parseNativeResourceContent(text);
  if (parsed.command != null || parsed.mode === 'cli' || parsed.args) {
    return formatRedisStructuredCommand(parsed);
  }
  return text;
}

function formatRedisStructuredCommand(parsed) {
  // 模型常产出 {"command":["SCAN","0","MATCH","mini*","COUNT","100"]}
  if (Array.isArray(parsed.command)) {
    return parsed.command.map((item) => String(item)).join(' ');
  }

  const command = String(parsed.command || '').trim();
  if (!command) {
    if (Array.isArray(parsed.args) && parsed.args.length) {
      return parsed.args.map((item) => String(item)).join(' ');
    }
    return '';
  }

  // 误把数组 String() 后变成 "SCAN,0,MATCH,..."
  if (command.includes(',')) {
    const repaired = normalizeCommaJoinedRedisTokens(command);
    if (repaired !== command) return repaired;
  }

  const upper = command.toUpperCase();
  if (upper === 'SCAN') {
    const cursor = parsed.cursor != null ? String(parsed.cursor) : '0';
    const parts = ['SCAN', cursor];
    if (parsed.match != null && String(parsed.match).trim() !== '') {
      parts.push('MATCH', String(parsed.match).trim());
    }
    if (parsed.count != null && String(parsed.count).trim() !== '') {
      parts.push('COUNT', String(parsed.count).trim());
    }
    return parts.join(' ');
  }
  if (parsed.args && Array.isArray(parsed.args)) {
    return [command, ...parsed.args.map((item) => String(item))].join(' ');
  }
  return command;
}

/** Array.toString / join(', ') 残留：SCAN,0,MATCH,mini* → SCAN 0 MATCH mini* */
function normalizeCommaJoinedRedisTokens(text) {
  const raw = String(text || '').trim();
  if (!raw.includes(',')) return raw;
  const tokens = raw.split(',').map((part) => part.trim()).filter(Boolean);
  if (tokens.length < 2) return raw;
  const head = tokens[0].toUpperCase();
  if (!/^[A-Z][A-Z0-9]*$/.test(head)) return raw;
  return tokens.join(' ');
}
