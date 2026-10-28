// website/assets/js/theme.js
export const THEME_IDS = ['graphite', 'amber', 'aurora', 'paper'];
export const THEME_STORAGE_KEY = 'sshtools-site-theme';

export const THEME_PALETTES = {
  graphite: { bg: '#0b1214', accent: '#2dd4bf', particle: '#5eead4', line: '#2dd4bf66' },
  amber: { bg: '#12100c', accent: '#f59e0b', particle: '#fbbf24', line: '#f59e0b66' },
  aurora: { bg: '#0a0f18', accent: '#38bdf8', particle: '#67e8f9', line: '#38bdf866' },
  paper: { bg: '#f4f7f6', accent: '#0f766e', particle: '#0d9488', line: '#0f766e44' },
};

export function normalizeThemeId(id) {
  return THEME_IDS.includes(id) ? id : 'graphite';
}

export function readStoredTheme(storage = globalThis.localStorage) {
  try {
    return normalizeThemeId(storage?.getItem?.(THEME_STORAGE_KEY));
  } catch {
    return 'graphite';
  }
}

export function applyTheme(id, { documentElement, storage, onPalette } = {}) {
  const theme = normalizeThemeId(id);
  const root = documentElement || globalThis.document?.documentElement;
  root?.setAttribute?.('data-theme', theme);
  try {
    (storage || globalThis.localStorage)?.setItem?.(THEME_STORAGE_KEY, theme);
  } catch { /* ignore */ }
  onPalette?.(THEME_PALETTES[theme]);
  return theme;
}
