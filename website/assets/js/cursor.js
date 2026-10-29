/** Cursor shape catalog: neon geometric line-art (ring / comet / cone). */

export const CURSOR_STORAGE_KEY = 'sshtools-site-cursor';

/**
 * @typedef {{ id: string, labelZh: string, labelEn: string, trail: 'dots'|'comet', burst: 'spark', rotate: boolean }} CursorStyle
 */

/** Inner SVG path markup in a 32×32 viewBox */
export const CURSOR_SVG_PATHS = {
  ring: `
    <circle cx="16" cy="16" r="10"/>
    <circle cx="16" cy="16" r="3"/>
  `,
  comet: `
    <circle cx="22" cy="10" r="3.2"/>
    <path d="M19.2 12.2 C12 16 7 20 4 27"/>
    <path d="M18 13.5 C12.5 17 9 21.5 7 26"/>
  `,
  cone: `
    <path d="M16 4 L24 26 L16 21.5 L8 26 Z"/>
    <path d="M16 10 L16 20"/>
    <circle cx="16" cy="12.5" r="1.5"/>
  `,
};

/** @type {CursorStyle[]} */
export const CURSOR_STYLES = [
  { id: 'ring', labelZh: '圆圈', labelEn: 'Ring', trail: 'dots', burst: 'spark', rotate: false },
  { id: 'comet', labelZh: '彗星', labelEn: 'Comet', trail: 'comet', burst: 'spark', rotate: true },
  { id: 'cone', labelZh: '锥形', labelEn: 'Cone', trail: 'comet', burst: 'spark', rotate: true },
];

export const CURSOR_STYLE_MAP = Object.fromEntries(CURSOR_STYLES.map((s) => [s.id, s]));

export function normalizeCursorId(id) {
  return Object.prototype.hasOwnProperty.call(CURSOR_STYLE_MAP, id) ? id : 'ring';
}

export function getCursorStyle(id) {
  return CURSOR_STYLE_MAP[normalizeCursorId(id)];
}

export function cursorSvgMarkup(id, { className = 'cursor-svg' } = {}) {
  const key = normalizeCursorId(id);
  const inner = CURSOR_SVG_PATHS[key] || CURSOR_SVG_PATHS.ring;
  return `<svg class="${className}" viewBox="0 0 32 32" aria-hidden="true" focusable="false"><g fill="none" stroke="currentColor" stroke-width="1.55" stroke-linecap="round" stroke-linejoin="round">${inner}</g></svg>`;
}

export function readStoredCursor(storage = globalThis.localStorage) {
  try {
    return normalizeCursorId(storage?.getItem?.(CURSOR_STORAGE_KEY));
  } catch {
    return 'ring';
  }
}

export function writeStoredCursor(id, storage = globalThis.localStorage) {
  const next = normalizeCursorId(id);
  try {
    storage?.setItem?.(CURSOR_STORAGE_KEY, next);
  } catch { /* ignore */ }
  return next;
}

export function applyCursorStyle(id, {
  documentElement,
  storage,
} = {}) {
  const style = getCursorStyle(id);
  const next = writeStoredCursor(style.id, storage || globalThis.localStorage);
  const root = documentElement || globalThis.document?.documentElement;
  root?.setAttribute?.('data-cursor', next);
  try {
    globalThis.dispatchEvent?.(new CustomEvent('sshtools:cursor', { detail: { id: next, style } }));
  } catch { /* ignore */ }
  return style;
}

export function cursorLabel(style, lang = 'zh') {
  return lang === 'en' ? style.labelEn : style.labelZh;
}
