/** Cursor shape catalog: neon line-art SVG animals (no emoji). */

export const CURSOR_STORAGE_KEY = 'sshtools-site-cursor';

/**
 * @typedef {{ id: string, labelZh: string, labelEn: string, trail: 'dots'|'outline'|'comet'|'spark', burst: 'spark'|'outline'|'heart', group: 'basic'|'pet'|'zodiac' }} CursorStyle
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
  spark: `
    <path d="M16 3 L18.2 13.2 L28 16 L18.2 18.8 L16 29 L13.8 18.8 L4 16 L13.8 13.2 Z"/>
    <path d="M16 10 L16 22 M10 16 L22 16"/>
  `,
  heart: `
    <path d="M16 27 C16 27 5 19.5 5 12.5 C5 8.8 7.8 6.5 11 6.5 C13.2 6.5 14.9 7.7 16 9.2 C17.1 7.7 18.8 6.5 21 6.5 C24.2 6.5 27 8.8 27 12.5 C27 19.5 16 27 16 27 Z"/>
  `,
  ghost: `
    <path d="M8 14 C8 8.5 11.5 5 16 5 C20.5 5 24 8.5 24 14 L24 26 L21 23.5 L18 26 L16 23.5 L14 26 L11 23.5 L8 26 Z"/>
    <circle cx="12.5" cy="14" r="1.4"/>
    <circle cx="19.5" cy="14" r="1.4"/>
  `,
  rocket: `
    <path d="M16 4 L22 18 L16 15.5 L10 18 Z"/>
    <path d="M12.5 17.5 L10 26"/>
    <path d="M19.5 17.5 L22 26"/>
    <path d="M13 20 L16 22.5 L19 20"/>
    <circle cx="16" cy="11" r="1.6"/>
  `,
  cat: `
    <circle cx="16" cy="18" r="8"/>
    <path d="M9 14 L10.5 5.5 L14.5 12.5"/>
    <path d="M23 14 L21.5 5.5 L17.5 12.5"/>
    <path d="M11 18.5 L7 17 M11 20.5 L7 22"/>
    <path d="M21 18.5 L25 17 M21 20.5 L25 22"/>
    <circle cx="13" cy="17.5" r="1.1"/>
    <circle cx="19" cy="17.5" r="1.1"/>
    <path d="M15.2 21.2 C15.7 22 16.3 22 16.8 21.2"/>
  `,
  bunny: `
    <circle cx="16" cy="20" r="7.2"/>
    <path d="M11.5 15 C10 6 12.5 3 14.2 12.5"/>
    <path d="M20.5 15 C22 6 19.5 3 17.8 12.5"/>
    <circle cx="13.5" cy="19.5" r="1.1"/>
    <circle cx="18.5" cy="19.5" r="1.1"/>
    <path d="M15.2 22.2 C15.7 23 16.3 23 16.8 22.2"/>
  `,
  fox: `
    <path d="M6 24 L16 7 L26 24 Z"/>
    <path d="M11 24 C13 20.5 19 20.5 21 24"/>
    <path d="M10 12.5 L7 7.5 L13 11"/>
    <path d="M22 12.5 L25 7.5 L19 11"/>
    <circle cx="13.2" cy="16.5" r="1.1"/>
    <circle cx="18.8" cy="16.5" r="1.1"/>
    <path d="M16 18.2 L16 21"/>
  `,
  rat: `
    <ellipse cx="14" cy="17" rx="7" ry="6"/>
    <path d="M20 18 C24 16 27 20 28 25"/>
    <path d="M10 12 L8 7 L12.5 11"/>
    <path d="M16.5 11.5 L18 7 L19.5 12"/>
    <circle cx="12" cy="16.5" r="1"/>
    <circle cx="16.5" cy="16.5" r="1"/>
    <path d="M11 20.5 L9.5 22"/>
  `,
  ox: `
    <ellipse cx="16" cy="18" rx="8" ry="7"/>
    <path d="M8 12 C5 7 8 5 11 10"/>
    <path d="M24 12 C27 7 24 5 21 10"/>
    <circle cx="13" cy="17" r="1.1"/>
    <circle cx="19" cy="17" r="1.1"/>
    <path d="M13 22 H19"/>
    <path d="M16 20 V22.5"/>
  `,
  tiger: `
    <circle cx="16" cy="18" r="8"/>
    <path d="M9 14 L10.2 5.8 L14.2 12.8"/>
    <path d="M23 14 L21.8 5.8 L17.8 12.8"/>
    <path d="M12.5 15.5 L14 17.5 L12.5 19.5"/>
    <path d="M19.5 15.5 L18 17.5 L19.5 19.5"/>
    <circle cx="13" cy="17.2" r="1"/>
    <circle cx="19" cy="17.2" r="1"/>
    <path d="M15 21.2 H17"/>
  `,
  rabbit: `
    <circle cx="16" cy="20.5" r="7"/>
    <path d="M11 16 C9.5 7 12 3.5 14 13"/>
    <path d="M21 16 C22.5 7 20 3.5 18 13"/>
    <circle cx="13.5" cy="19.8" r="1"/>
    <circle cx="18.5" cy="19.8" r="1"/>
    <path d="M15.2 22.8 C15.7 23.5 16.3 23.5 16.8 22.8"/>
  `,
  dragon: `
    <path d="M5 22 C8 14 12 10 16 12 C20 14 22 10 27 8"/>
    <path d="M16 12 C15 17 17 22 22 25"/>
    <path d="M24.5 8.5 L26 5.5 M25.5 9.5 L28 8"/>
    <circle cx="25.2" cy="8.2" r="1"/>
    <path d="M10 18 L8 16 M12 20 L10 22"/>
  `,
  snake: `
    <path d="M6 22 C10 14 14 22 18 14 C21 9 24 12 27 10"/>
    <circle cx="27" cy="9.5" r="1.3"/>
    <path d="M26 10.5 L24.5 13"/>
  `,
  horse: `
    <path d="M8 24 L11 14 L16 11 L22 12 L25 9"/>
    <path d="M16 11 L15 24"/>
    <path d="M22 12 L24 18 L22 24"/>
    <path d="M11 14 L8 10 L12 12"/>
    <circle cx="20.5" cy="12.5" r="0.9"/>
  `,
  goat: `
    <ellipse cx="16" cy="19" rx="7.5" ry="6.5"/>
    <path d="M10 14 C8 8 11 6 13.5 12"/>
    <path d="M22 14 C24 8 21 6 18.5 12"/>
    <circle cx="13.5" cy="18" r="1"/>
    <circle cx="18.5" cy="18" r="1"/>
    <path d="M15 21.5 H17"/>
  `,
  monkey: `
    <circle cx="16" cy="17" r="8"/>
    <circle cx="7.5" cy="15" r="3"/>
    <circle cx="24.5" cy="15" r="3"/>
    <ellipse cx="16" cy="20" rx="4" ry="3"/>
    <circle cx="13" cy="15.5" r="1.1"/>
    <circle cx="19" cy="15.5" r="1.1"/>
  `,
  rooster: `
    <path d="M10 22 C10 15 14 11 18 12 C22 13 24 17 23 22"/>
    <path d="M18 12 L20 6 L22 11 L24 7 L23.5 12.5"/>
    <path d="M23 16 L28 15"/>
    <path d="M12 22 L10 27 M16 22.5 L16 28 M20 22 L22 27"/>
    <circle cx="20.5" cy="15.5" r="0.9"/>
  `,
  dog: `
    <circle cx="16" cy="18" r="7.5"/>
    <path d="M9.5 14 C7 10 6 15 10 17"/>
    <path d="M22.5 14 C25 10 26 15 22 17"/>
    <path d="M16 19.5 L16 22.5 L18 23"/>
    <circle cx="13.5" cy="17" r="1"/>
    <circle cx="18.5" cy="17" r="1"/>
  `,
  pig: `
    <circle cx="16" cy="17.5" r="8"/>
    <ellipse cx="16" cy="21" rx="3.4" ry="2.2"/>
    <path d="M10 12 L8.5 7.5 L12.5 11"/>
    <path d="M22 12 L23.5 7.5 L19.5 11"/>
    <circle cx="13" cy="16" r="1"/>
    <circle cx="19" cy="16" r="1"/>
    <circle cx="14.7" cy="21" r="0.6"/>
    <circle cx="17.3" cy="21" r="0.6"/>
  `,
};

/** @type {CursorStyle[]} */
export const CURSOR_STYLES = [
  { id: 'ring', labelZh: '光环', labelEn: 'Ring', trail: 'dots', burst: 'spark', group: 'basic' },
  { id: 'comet', labelZh: '彗星', labelEn: 'Comet', trail: 'comet', burst: 'spark', group: 'basic' },
  { id: 'spark', labelZh: '星芒', labelEn: 'Spark', trail: 'spark', burst: 'spark', group: 'basic' },
  { id: 'heart', labelZh: '爱心', labelEn: 'Heart', trail: 'outline', burst: 'heart', group: 'basic' },
  { id: 'ghost', labelZh: '幽灵', labelEn: 'Ghost', trail: 'outline', burst: 'outline', group: 'basic' },
  { id: 'rocket', labelZh: '火箭', labelEn: 'Rocket', trail: 'comet', burst: 'spark', group: 'basic' },
  { id: 'cat', labelZh: '小猫', labelEn: 'Cat', trail: 'outline', burst: 'outline', group: 'pet' },
  { id: 'bunny', labelZh: '兔子', labelEn: 'Bunny', trail: 'outline', burst: 'outline', group: 'pet' },
  { id: 'fox', labelZh: '狐狸', labelEn: 'Fox', trail: 'outline', burst: 'outline', group: 'pet' },
  { id: 'rat', labelZh: '鼠', labelEn: 'Rat', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'ox', labelZh: '牛', labelEn: 'Ox', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'tiger', labelZh: '虎', labelEn: 'Tiger', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'rabbit', labelZh: '兔', labelEn: 'Rabbit', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'dragon', labelZh: '龙', labelEn: 'Dragon', trail: 'comet', burst: 'outline', group: 'zodiac' },
  { id: 'snake', labelZh: '蛇', labelEn: 'Snake', trail: 'comet', burst: 'outline', group: 'zodiac' },
  { id: 'horse', labelZh: '马', labelEn: 'Horse', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'goat', labelZh: '羊', labelEn: 'Goat', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'monkey', labelZh: '猴', labelEn: 'Monkey', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'rooster', labelZh: '鸡', labelEn: 'Rooster', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'dog', labelZh: '狗', labelEn: 'Dog', trail: 'outline', burst: 'outline', group: 'zodiac' },
  { id: 'pig', labelZh: '猪', labelEn: 'Pig', trail: 'outline', burst: 'outline', group: 'zodiac' },
];

export const CURSOR_STYLE_MAP = Object.fromEntries(CURSOR_STYLES.map((s) => [s.id, s]));

export function normalizeCursorId(id) {
  return Object.prototype.hasOwnProperty.call(CURSOR_STYLE_MAP, id) ? id : 'cat';
}

export function getCursorStyle(id) {
  return CURSOR_STYLE_MAP[normalizeCursorId(id)];
}

export function cursorSvgMarkup(id, { className = 'cursor-svg' } = {}) {
  const key = normalizeCursorId(id);
  const inner = CURSOR_SVG_PATHS[key] || CURSOR_SVG_PATHS.cat;
  return `<svg class="${className}" viewBox="0 0 32 32" aria-hidden="true" focusable="false"><g fill="none" stroke="currentColor" stroke-width="1.55" stroke-linecap="round" stroke-linejoin="round">${inner}</g></svg>`;
}

export function readStoredCursor(storage = globalThis.localStorage) {
  try {
    return normalizeCursorId(storage?.getItem?.(CURSOR_STORAGE_KEY));
  } catch {
    return 'cat';
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
