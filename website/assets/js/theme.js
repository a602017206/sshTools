/** Theme modes: preset | duo (bg+fg) | trio (bg+fg+accent) */
export const THEME_MODE_IDS = ['preset', 'duo', 'trio'];
export const THEME_MODE_STORAGE_KEY = 'sshtools-site-theme-mode';
export const THEME_STORAGE_KEY = 'sshtools-site-theme';
export const THEME_BG_STORAGE_KEY = 'sshtools-site-bg';
export const THEME_FG_STORAGE_KEY = 'sshtools-site-fg';
export const THEME_ACCENT_STORAGE_KEY = 'sshtools-site-accent';

/** @deprecated kept as alias of preset ids for older callers */
export const THEME_IDS = [
  'graphite', 'amber', 'aurora', 'paper',
  'ember', 'ocean', 'forest', 'orchid', 'sand', 'noir',
];

export const BG_IDS = [
  'midnight', 'graphite', 'amberNight', 'oceanDeep', 'forestDeep',
  'ink', 'cream', 'paper', 'sand', 'lavenderMist',
];

export const FG_IDS = [
  'snow', 'mist', 'ice', 'ink', 'charcoal', 'tealTint', 'goldTint',
];

export const ACCENT_IDS = [
  'teal', 'amber', 'sky', 'lime', 'coral', 'violet', 'cyan', 'rose',
];

export const BG_TOKENS = {
  midnight: {
    bg: '#070b10', surface: '#101820', surfaceRaised: '#15202b',
    border: '#1c2a36', shadow: 'rgb(0 0 0 / 34%)', scheme: 'dark',
  },
  graphite: {
    bg: '#0b1214', surface: '#121a1c', surfaceRaised: '#172225',
    border: '#1e2a2e', shadow: 'rgb(0 0 0 / 30%)', scheme: 'dark',
  },
  amberNight: {
    bg: '#12100c', surface: '#1b1710', surfaceRaised: '#241e14',
    border: '#342b1d', shadow: 'rgb(0 0 0 / 32%)', scheme: 'dark',
  },
  oceanDeep: {
    bg: '#061018', surface: '#0d1b28', surfaceRaised: '#122536',
    border: '#1a3348', shadow: 'rgb(0 0 0 / 34%)', scheme: 'dark',
  },
  forestDeep: {
    bg: '#08110d', surface: '#101a15', surfaceRaised: '#16231c',
    border: '#21362b', shadow: 'rgb(0 0 0 / 32%)', scheme: 'dark',
  },
  ink: {
    bg: '#0a0a0c', surface: '#141418', surfaceRaised: '#1b1b21',
    border: '#2a2a33', shadow: 'rgb(0 0 0 / 36%)', scheme: 'dark',
  },
  cream: {
    bg: '#f6f1e8', surface: '#fffaf2', surfaceRaised: '#ffffff',
    border: '#e4d8c4', shadow: 'rgb(48 36 18 / 10%)', scheme: 'light',
  },
  paper: {
    bg: '#f4f7f6', surface: '#ffffff', surfaceRaised: '#ffffff',
    border: '#d5e0dd', shadow: 'rgb(18 32 30 / 10%)', scheme: 'light',
  },
  sand: {
    bg: '#f3eee6', surface: '#fbf7f0', surfaceRaised: '#ffffff',
    border: '#e2d6c5', shadow: 'rgb(60 42 20 / 10%)', scheme: 'light',
  },
  lavenderMist: {
    bg: '#f3f1f8', surface: '#fbfaff', surfaceRaised: '#ffffff',
    border: '#ddd6ea', shadow: 'rgb(40 28 60 / 10%)', scheme: 'light',
  },
};

export const FG_TOKENS = {
  snow: { fg: '#e8eef0', muted: '#9db0b5' },
  mist: { fg: '#d7e0e6', muted: '#8fa0a8' },
  ice: { fg: '#e6edf7', muted: '#9cacbf' },
  ink: { fg: '#12201e', muted: '#5b6f6c' },
  charcoal: { fg: '#1a1714', muted: '#6b5f52' },
  tealTint: { fg: '#d7f6ef', muted: '#7fb5aa' },
  goldTint: { fg: '#f6eccf', muted: '#b7a078' },
};

export const ACCENT_TOKENS = {
  teal: { accent: '#2dd4bf', particle: '#5eead4', line: '#2dd4bf66', glow: 'rgb(45 212 191 / 35%)' },
  amber: { accent: '#f59e0b', particle: '#fbbf24', line: '#f59e0b66', glow: 'rgb(245 158 11 / 32%)' },
  sky: { accent: '#38bdf8', particle: '#67e8f9', line: '#38bdf866', glow: 'rgb(56 189 248 / 34%)' },
  lime: { accent: '#84cc16', particle: '#bef264', line: '#84cc1666', glow: 'rgb(132 204 22 / 30%)' },
  coral: { accent: '#fb7185', particle: '#fda4af', line: '#fb718566', glow: 'rgb(251 113 133 / 30%)' },
  violet: { accent: '#a78bfa', particle: '#c4b5fd', line: '#a78bfa66', glow: 'rgb(167 139 250 / 30%)' },
  cyan: { accent: '#22d3ee', particle: '#67e8f9', line: '#22d3ee66', glow: 'rgb(34 211 238 / 32%)' },
  rose: { accent: '#f472b6', particle: '#f9a8d4', line: '#f472b666', glow: 'rgb(244 114 182 / 30%)' },
};

/** Named full combinations (preset model) */
export const THEME_PRESETS = {
  graphite: { bg: 'graphite', fg: 'snow', accent: 'teal', label: 'Graphite' },
  amber: { bg: 'amberNight', fg: 'goldTint', accent: 'amber', label: 'Amber' },
  aurora: { bg: 'oceanDeep', fg: 'ice', accent: 'sky', label: 'Aurora' },
  paper: { bg: 'paper', fg: 'ink', accent: 'teal', label: 'Paper' },
  ember: { bg: 'amberNight', fg: 'snow', accent: 'coral', label: 'Ember' },
  ocean: { bg: 'oceanDeep', fg: 'ice', accent: 'cyan', label: 'Ocean' },
  forest: { bg: 'forestDeep', fg: 'tealTint', accent: 'lime', label: 'Forest' },
  orchid: { bg: 'ink', fg: 'mist', accent: 'violet', label: 'Orchid' },
  sand: { bg: 'sand', fg: 'charcoal', accent: 'amber', label: 'Sand' },
  noir: { bg: 'midnight', fg: 'snow', accent: 'rose', label: 'Noir' },
};

export function normalizeThemeMode(mode) {
  return THEME_MODE_IDS.includes(mode) ? mode : 'preset';
}

export function normalizeThemeId(id) {
  return Object.prototype.hasOwnProperty.call(THEME_PRESETS, id) ? id : 'graphite';
}

export function normalizeBgId(id) {
  return BG_IDS.includes(id) ? id : 'graphite';
}

export function normalizeFgId(id) {
  return FG_IDS.includes(id) ? id : 'snow';
}

export function normalizeAccentId(id) {
  return ACCENT_IDS.includes(id) ? id : 'teal';
}

export function resolveTokens(bgId, fgId, accentId) {
  const bg = BG_TOKENS[normalizeBgId(bgId)];
  const fg = FG_TOKENS[normalizeFgId(fgId)];
  const accent = ACCENT_TOKENS[normalizeAccentId(accentId)];
  return {
    bg: bg.bg,
    surface: bg.surface,
    surfaceRaised: bg.surfaceRaised,
    border: bg.border,
    shadow: bg.shadow,
    scheme: bg.scheme,
    fg: fg.fg,
    muted: fg.muted,
    accent: accent.accent,
    particle: accent.particle,
    line: accent.line,
    glow: accent.glow,
    accentSoft: `${accent.accent}29`,
  };
}

/** @deprecated scene/nav callers — built from presets */
export const THEME_PALETTES = Object.fromEntries(
  Object.entries(THEME_PRESETS).map(([id, preset]) => [id, resolveTokens(preset.bg, preset.fg, preset.accent)]),
);

export function resolveAppearance(state) {
  const mode = normalizeThemeMode(state?.mode);
  if (mode === 'preset') {
    const preset = THEME_PRESETS[normalizeThemeId(state?.preset)];
    return {
      mode,
      preset: normalizeThemeId(state?.preset),
      bg: preset.bg,
      fg: preset.fg,
      accent: preset.accent,
      tokens: resolveTokens(preset.bg, preset.fg, preset.accent),
    };
  }
  if (mode === 'duo') {
    const bg = normalizeBgId(state?.bg);
    const fg = normalizeFgId(state?.fg);
    const accent = normalizeAccentId(state?.accent || (BG_TOKENS[bg].scheme === 'light' ? 'teal' : 'cyan'));
    return {
      mode,
      preset: normalizeThemeId(state?.preset),
      bg,
      fg,
      accent,
      tokens: resolveTokens(bg, fg, accent),
    };
  }
  // trio
  const bg = normalizeBgId(state?.bg);
  const fg = normalizeFgId(state?.fg);
  const accent = normalizeAccentId(state?.accent);
  return {
    mode,
    preset: normalizeThemeId(state?.preset),
    bg,
    fg,
    accent,
    tokens: resolveTokens(bg, fg, accent),
  };
}

function safeGet(storage, key) {
  try {
    return storage?.getItem?.(key) ?? null;
  } catch {
    return null;
  }
}

function safeSet(storage, key, value) {
  try {
    storage?.setItem?.(key, value);
  } catch { /* ignore */ }
}

export function readThemeState(storage = globalThis.localStorage) {
  return {
    mode: normalizeThemeMode(safeGet(storage, THEME_MODE_STORAGE_KEY)),
    preset: normalizeThemeId(safeGet(storage, THEME_STORAGE_KEY)),
    bg: normalizeBgId(safeGet(storage, THEME_BG_STORAGE_KEY)),
    fg: normalizeFgId(safeGet(storage, THEME_FG_STORAGE_KEY)),
    accent: normalizeAccentId(safeGet(storage, THEME_ACCENT_STORAGE_KEY)),
  };
}

export function readStoredTheme(storage = globalThis.localStorage) {
  return readThemeState(storage).preset;
}

export function writeThemeState(state, storage = globalThis.localStorage) {
  const next = resolveAppearance(state);
  safeSet(storage, THEME_MODE_STORAGE_KEY, next.mode);
  safeSet(storage, THEME_STORAGE_KEY, next.preset);
  safeSet(storage, THEME_BG_STORAGE_KEY, next.bg);
  safeSet(storage, THEME_FG_STORAGE_KEY, next.fg);
  safeSet(storage, THEME_ACCENT_STORAGE_KEY, next.accent);
  return next;
}

export function applyCssTokens(tokens, documentElement) {
  const root = documentElement || globalThis.document?.documentElement;
  if (!root?.style?.setProperty) return;
  root.style.setProperty('--bg', tokens.bg);
  root.style.setProperty('--fg', tokens.fg);
  root.style.setProperty('--muted', tokens.muted);
  root.style.setProperty('--accent', tokens.accent);
  root.style.setProperty('--accent-soft', tokens.accentSoft);
  root.style.setProperty('--surface', tokens.surface);
  root.style.setProperty('--surface-raised', tokens.surfaceRaised);
  root.style.setProperty('--border', tokens.border);
  root.style.setProperty('--shadow', tokens.shadow);
  root.style.setProperty('--glow', tokens.glow);
  root.style.colorScheme = tokens.scheme;
  root.setAttribute('data-scheme', tokens.scheme);
}

export function applyAppearance(partialState = {}, {
  documentElement,
  storage,
  onPalette,
} = {}) {
  const current = readThemeState(storage || globalThis.localStorage);
  const merged = { ...current, ...partialState };
  const appearance = writeThemeState(merged, storage || globalThis.localStorage);
  const root = documentElement || globalThis.document?.documentElement;
  root?.setAttribute?.('data-theme', appearance.preset);
  root?.setAttribute?.('data-theme-mode', appearance.mode);
  applyCssTokens(appearance.tokens, root);
  onPalette?.(appearance.tokens);
  return appearance;
}

/** Back-compat: apply a named preset */
export function applyTheme(id, options = {}) {
  return applyAppearance({ mode: 'preset', preset: id }, options);
}
