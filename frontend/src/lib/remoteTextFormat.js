import { isMap, isSeq, parseDocument } from 'yaml';
import { fileBaseName, fileExtension } from './remoteTextFile.js';

const GENERIC_EXTENSIONS = new Set(['', 'txt', 'log', 'conf', 'cfg', 'config', 'cnf', 'inc', 'local', 'bak', 'tmp', 'text']);

const EXTENSION_LANGUAGES = {
  json: { id: 'json', label: 'JSON' },
  jsonc: { id: 'jsonc', label: 'JSONC' },
  yml: { id: 'yaml', label: 'YAML' },
  yaml: { id: 'yaml', label: 'YAML' },
  xml: { id: 'xml', label: 'XML' },
  svg: { id: 'xml', label: 'XML' },
  html: { id: 'html', label: 'HTML' },
  htm: { id: 'html', label: 'HTML' },
  css: { id: 'css', label: 'CSS' },
  scss: { id: 'css', label: 'CSS' },
  less: { id: 'css', label: 'CSS' },
  js: { id: 'javascript', label: 'JavaScript' },
  mjs: { id: 'javascript', label: 'JavaScript' },
  cjs: { id: 'javascript', label: 'JavaScript' },
  jsx: { id: 'javascript', label: 'JavaScript' },
  ts: { id: 'javascript', label: 'TypeScript' },
  tsx: { id: 'javascript', label: 'TypeScript' },
  md: { id: 'markdown', label: 'Markdown' },
  markdown: { id: 'markdown', label: 'Markdown' },
  sql: { id: 'sql', label: 'SQL' },
  py: { id: 'python', label: 'Python' },
  sh: { id: 'shell', label: 'Shell' },
  bash: { id: 'shell', label: 'Shell' },
  zsh: { id: 'shell', label: 'Shell' },
  toml: { id: 'toml', label: 'TOML' },
  ini: { id: 'properties', label: 'INI' },
  properties: { id: 'properties', label: 'Properties' },
  env: { id: 'properties', label: 'ENV' },
  service: { id: 'properties', label: 'systemd' },
  socket: { id: 'properties', label: 'systemd' },
  timer: { id: 'properties', label: 'systemd' },
  nginx: { id: 'nginx', label: 'Nginx' },
  vue: { id: 'html', label: 'Vue' },
  svelte: { id: 'html', label: 'Svelte' },
};

const NAME_LANGUAGES = {
  dockerfile: { id: 'dockerfile', label: 'Dockerfile' },
  makefile: { id: 'text', label: 'Makefile' },
  'nginx.conf': { id: 'nginx', label: 'Nginx' },
  '.bashrc': { id: 'shell', label: 'Shell' },
  '.zshrc': { id: 'shell', label: 'Shell' },
  '.profile': { id: 'shell', label: 'Shell' },
  '.bash_profile': { id: 'shell', label: 'Shell' },
  '.gitignore': { id: 'text', label: '文本' },
  '.editorconfig': { id: 'properties', label: 'Properties' },
  '.npmrc': { id: 'properties', label: 'Properties' },
};

const SQL_KEYWORDS = [
  'left outer join',
  'right outer join',
  'full outer join',
  'inner join',
  'left join',
  'right join',
  'cross join',
  'group by',
  'order by',
  'insert into',
  'delete from',
  'union all',
  'select',
  'from',
  'where',
  'having',
  'limit',
  'offset',
  'values',
  'union',
  'join',
  'update',
  'set',
];

function toLf(source) {
  return String(source ?? '').replace(/\r\n/g, '\n').replace(/\r/g, '\n');
}

function namedLanguage(filename) {
  const base = fileBaseName(filename).toLowerCase();
  if (NAME_LANGUAGES[base]) return NAME_LANGUAGES[base];
  if (base === '.env' || base.startsWith('.env.')) return { id: 'properties', label: 'ENV' };
  return null;
}

export function sniffLanguage(content) {
  const text = toLf(content);
  const trimmed = text.trim();
  if (!trimmed) return null;
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      JSON.parse(trimmed);
      return { id: 'json', label: 'JSON' };
    } catch {
      // 不是 JSON 时继续按其他格式判断。
    }
  }
  if (trimmed.startsWith('<')) {
    if (/^<!doctype html/i.test(trimmed) || /<html[\s>]/i.test(trimmed)) {
      return { id: 'html', label: 'HTML' };
    }
    if (/^<\?xml/i.test(trimmed) || /^<\/?[A-Za-z!]/.test(trimmed)) {
      return { id: 'xml', label: 'XML' };
    }
  }
  if (looksLikeYaml(text)) return { id: 'yaml', label: 'YAML' };
  if (looksLikeProperties(text)) return { id: 'properties', label: 'Properties' };
  if (/^\s*(select|insert|update|delete|with)\b/i.test(trimmed)) return { id: 'sql', label: 'SQL' };
  return null;
}

function looksLikeYaml(text) {
  const trimmed = text.trim();
  if (!trimmed || trimmed.startsWith('{') || trimmed.startsWith('[') || trimmed.startsWith('<')) return false;
  const doc = parseDocument(trimmed);
  if (doc.errors?.length) return false;
  return isMap(doc.contents) || isSeq(doc.contents);
}

function looksLikeProperties(text) {
  const lines = toLf(text).split('\n').map((line) => line.trim()).filter((line) => line && !line.startsWith('#') && !line.startsWith(';'));
  if (!lines.length) return false;
  const matched = lines.filter((line) => /^\[[^\]\n]+\]$/.test(line) || /^[^=\s][^=]*=/.test(line)).length;
  return matched / lines.length >= 0.6;
}

export function resolveTextLanguage(filename, content = '') {
  const named = namedLanguage(filename);
  if (named) return named;
  const extension = fileExtension(filename);
  const byExtension = EXTENSION_LANGUAGES[extension] || null;
  if (byExtension && !GENERIC_EXTENSIONS.has(extension)) return byExtension;
  const sniffed = sniffLanguage(content);
  if (sniffed) return sniffed;
  if (byExtension) return byExtension;
  return { id: 'text', label: '文本' };
}

export function formatJson(source) {
  const text = toLf(source).trim().replace(/^\uFEFF/, '');
  try {
    return `${JSON.stringify(JSON.parse(text), null, 2)}\n`;
  } catch {
    throw new Error('JSON 无法解析');
  }
}

export function formatYaml(source) {
  const doc = parseDocument(toLf(source));
  if (doc.errors?.length) throw new Error('YAML 无法解析');
  if (!isMap(doc.contents) && !isSeq(doc.contents)) throw new Error('YAML 内容不是映射或列表');
  const out = doc.toString({ lineWidth: 0 });
  return out.endsWith('\n') || out === '' ? out : `${out}\n`;
}

export function formatXml(source) {
  const text = toLf(source).trim();
  if (!text.startsWith('<')) throw new Error('内容不是标记文本');
  const parts = text.split(/(<[^>]*>)/).filter((part) => part !== '');
  let pad = 0;
  const lines = [];
  parts.forEach((part) => {
    const piece = part.trim();
    if (!piece) return;
    if (piece.startsWith('</')) pad = Math.max(0, pad - 1);
    lines.push(`${'  '.repeat(pad)}${piece}`);
    const isTag = piece.startsWith('<');
    const selfClosing = /\/>$/.test(piece);
    const special = piece.startsWith('<?') || piece.startsWith('<!');
    if (isTag && !piece.startsWith('</') && !selfClosing && !special) pad += 1;
  });
  return `${lines.join('\n')}\n`;
}

function isWordChar(char) {
  return /[A-Za-z0-9_]/.test(char || '');
}

function matchSqlKeyword(text, index) {
  if (index > 0 && isWordChar(text[index - 1])) return '';
  const rest = text.slice(index);
  for (const keyword of SQL_KEYWORDS) {
    if (rest.length < keyword.length) continue;
    if (rest.slice(0, keyword.length).toLowerCase() !== keyword) continue;
    if (isWordChar(rest[keyword.length])) continue;
    return rest.slice(0, keyword.length);
  }
  return '';
}

function readQuoted(text, start, quote) {
  let index = start + 1;
  while (index < text.length) {
    if (text[index] === '\\' && quote !== '`') {
      index += 2;
      continue;
    }
    if (text[index] === quote) {
      if (text[index + 1] === quote) {
        index += 2;
        continue;
      }
      return index + 1;
    }
    index += 1;
  }
  return text.length;
}

export function formatSql(source) {
  const text = toLf(source).trim();
  if (!text) return '';
  let out = '';
  let index = 0;
  while (index < text.length) {
    const char = text[index];
    if (char === "'" || char === '"' || char === '`') {
      const end = readQuoted(text, index, char);
      out += text.slice(index, end);
      index = end;
      continue;
    }
    if (char === '-' && text[index + 1] === '-') {
      const end = text.indexOf('\n', index);
      const stop = end === -1 ? text.length : end;
      out += text.slice(index, stop);
      index = stop;
      continue;
    }
    if (char === '/' && text[index + 1] === '*') {
      const end = text.indexOf('*/', index + 2);
      const stop = end === -1 ? text.length : end + 2;
      out += text.slice(index, stop);
      index = stop;
      continue;
    }
    const keyword = matchSqlKeyword(text, index);
    if (keyword) {
      out = out.replace(/[ \t]+$/g, '');
      if (out && !out.endsWith('\n')) out += '\n';
      out += keyword;
      index += keyword.length;
      continue;
    }
    if (char === '\n' || char === '\t' || char === ' ') {
      if (out && !out.endsWith('\n') && !out.endsWith(' ')) out += ' ';
      index += 1;
      continue;
    }
    out += char;
    index += 1;
  }
  return `${out.replace(/[ \t]+\n/g, '\n').trim()}\n`;
}

export function formatPlain(source) {
  const lines = toLf(source).split('\n');
  while (lines.length > 1 && lines[lines.length - 1] === '') lines.pop();
  const body = lines.map((line) => line.replace(/[ \t]+$/g, '')).join('\n');
  if (!body) return '';
  return body.endsWith('\n') ? body : `${body}\n`;
}

function formatByLanguage(id, content) {
  const text = toLf(content);
  if (!text.trim()) return '';
  switch (id) {
    case 'json':
      return formatJson(text);
    case 'yaml':
      return formatYaml(text);
    case 'xml':
    case 'html':
      return formatXml(text);
    case 'sql':
      return formatSql(text);
    default:
      return formatPlain(text);
  }
}

export function formatRemoteText(filename, content) {
  const language = resolveTextLanguage(filename, content);
  try {
    const text = formatByLanguage(language.id, content);
    return {
      ok: true,
      text,
      language: language.id,
      label: language.label,
      changed: text !== toLf(content),
    };
  } catch (err) {
    return {
      ok: false,
      error: err?.message || '无法格式化',
      language: language.id,
      label: language.label,
      changed: false,
    };
  }
}
