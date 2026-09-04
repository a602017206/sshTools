export const TERMINAL_THEME_PRESETS = [
  { id: 'dark', label: '深色' },
  { id: 'light', label: '浅色' },
  { id: 'follow', label: '跟随界面' },
];

const DARK_XTERM_THEME = {
  background: '#050914',
  foreground: '#d8e1ef',
  cursor: '#39d5e8',
  black: '#000000',
  red: '#cd3131',
  green: '#0dbc79',
  yellow: '#e5e510',
  blue: '#2472c8',
  magenta: '#bc3fbc',
  cyan: '#11a8cd',
  white: '#e5e5e5',
  brightBlack: '#666666',
  brightRed: '#f14c4c',
  brightGreen: '#23d18b',
  brightYellow: '#f5f543',
  brightBlue: '#3b8eea',
  brightMagenta: '#d670d6',
  brightCyan: '#29b8db',
  brightWhite: '#e5e5e5',
  selectionBackground: 'rgba(57, 213, 232, 0.28)',
  selectionForeground: undefined,
  selectionInactiveBackground: 'rgba(57, 213, 232, 0.14)',
};

const LIGHT_XTERM_THEME = {
  background: '#f7f8fb',
  foreground: '#1e293b',
  cursor: '#0f766e',
  black: '#1e293b',
  red: '#b91c1c',
  green: '#15803d',
  yellow: '#a16207',
  blue: '#1d4ed8',
  magenta: '#7e22ce',
  cyan: '#0e7490',
  white: '#334155',
  brightBlack: '#64748b',
  brightRed: '#dc2626',
  brightGreen: '#16a34a',
  brightYellow: '#ca8a04',
  brightBlue: '#2563eb',
  brightMagenta: '#9333ea',
  brightCyan: '#0891b2',
  brightWhite: '#0f172a',
  selectionBackground: 'rgba(13, 148, 136, 0.22)',
  selectionForeground: undefined,
  selectionInactiveBackground: 'rgba(13, 148, 136, 0.12)',
};

const TERMINAL_SURFACE = {
  dark: {
    background: '#050914',
    border: 'rgba(51, 65, 85, 0.9)',
  },
  light: {
    background: '#f7f8fb',
    border: 'rgba(15, 23, 42, 0.12)',
  },
};

export function resolveTerminalTheme(terminalTheme, appTheme) {
  if (terminalTheme === 'light') {
    return 'light';
  }
  if (terminalTheme === 'follow') {
    return appTheme === 'light' ? 'light' : 'dark';
  }
  return 'dark';
}

export function getXtermTheme(resolvedTheme) {
  return resolvedTheme === 'light' ? LIGHT_XTERM_THEME : DARK_XTERM_THEME;
}

export function getTerminalSurface(resolvedTheme) {
  return resolvedTheme === 'light' ? TERMINAL_SURFACE.light : TERMINAL_SURFACE.dark;
}

export function resolveTerminalThemeFromSettings(settings = {}) {
  return resolveTerminalTheme(settings.terminal_theme, settings.theme);
}
