/** Cursor shape catalog + persistence for the marketing site. */

export const CURSOR_STORAGE_KEY = 'sshtools-site-cursor';

/**
 * @typedef {{ id: string, glyph: string, labelZh: string, labelEn: string, trail: 'dots'|'glyphs'|'comet'|'spark', burst: 'spark'|'emoji'|'heart', group: 'basic'|'pet'|'zodiac' }} CursorStyle
 */

/** @type {CursorStyle[]} */
export const CURSOR_STYLES = [
  { id: 'ring', glyph: '◎', labelZh: '光环', labelEn: 'Ring', trail: 'dots', burst: 'spark', group: 'basic' },
  { id: 'comet', glyph: '☄', labelZh: '彗星', labelEn: 'Comet', trail: 'comet', burst: 'spark', group: 'basic' },
  { id: 'spark', glyph: '✦', labelZh: '星芒', labelEn: 'Spark', trail: 'spark', burst: 'spark', group: 'basic' },
  { id: 'heart', glyph: '💖', labelZh: '爱心', labelEn: 'Heart', trail: 'glyphs', burst: 'heart', group: 'basic' },
  { id: 'ghost', glyph: '👻', labelZh: '幽灵', labelEn: 'Ghost', trail: 'glyphs', burst: 'emoji', group: 'basic' },
  { id: 'rocket', glyph: '🚀', labelZh: '火箭', labelEn: 'Rocket', trail: 'comet', burst: 'spark', group: 'basic' },
  { id: 'cat', glyph: '🐱', labelZh: '小猫', labelEn: 'Cat', trail: 'glyphs', burst: 'emoji', group: 'pet' },
  { id: 'bunny', glyph: '🐰', labelZh: '兔子', labelEn: 'Bunny', trail: 'glyphs', burst: 'emoji', group: 'pet' },
  { id: 'fox', glyph: '🦊', labelZh: '狐狸', labelEn: 'Fox', trail: 'glyphs', burst: 'emoji', group: 'pet' },
  { id: 'rat', glyph: '🐀', labelZh: '鼠', labelEn: 'Rat', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'ox', glyph: '🐂', labelZh: '牛', labelEn: 'Ox', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'tiger', glyph: '🐯', labelZh: '虎', labelEn: 'Tiger', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'rabbit', glyph: '🐇', labelZh: '兔', labelEn: 'Rabbit', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'dragon', glyph: '🐲', labelZh: '龙', labelEn: 'Dragon', trail: 'comet', burst: 'emoji', group: 'zodiac' },
  { id: 'snake', glyph: '🐍', labelZh: '蛇', labelEn: 'Snake', trail: 'comet', burst: 'emoji', group: 'zodiac' },
  { id: 'horse', glyph: '🐴', labelZh: '马', labelEn: 'Horse', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'goat', glyph: '🐐', labelZh: '羊', labelEn: 'Goat', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'monkey', glyph: '🐵', labelZh: '猴', labelEn: 'Monkey', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'rooster', glyph: '🐔', labelZh: '鸡', labelEn: 'Rooster', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'dog', glyph: '🐶', labelZh: '狗', labelEn: 'Dog', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
  { id: 'pig', glyph: '🐷', labelZh: '猪', labelEn: 'Pig', trail: 'glyphs', burst: 'emoji', group: 'zodiac' },
];

export const CURSOR_STYLE_MAP = Object.fromEntries(CURSOR_STYLES.map((s) => [s.id, s]));

export function normalizeCursorId(id) {
  return Object.prototype.hasOwnProperty.call(CURSOR_STYLE_MAP, id) ? id : 'cat';
}

export function getCursorStyle(id) {
  return CURSOR_STYLE_MAP[normalizeCursorId(id)];
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
