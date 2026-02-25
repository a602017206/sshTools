export const FONT_PRESETS = [
  { id: 'system-ui', label: 'System UI', value: '"Avenir Next", "SF Pro Text", "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif' },
  { id: 'inter', label: 'Inter', value: 'Inter, "Avenir Next", "Segoe UI", sans-serif' },
  { id: 'source-han', label: 'Source Han Sans', value: '"Source Han Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif' },
  { id: 'noto', label: 'Noto Sans', value: '"Noto Sans SC", "Segoe UI", sans-serif' }
];

export const TERMINAL_FONT_PRESETS = [
  { id: 'menlo', label: 'Menlo', value: 'Menlo, Monaco, "Courier New", monospace' },
  { id: 'jetbrains-mono', label: 'JetBrains Mono', value: '"JetBrains Mono", Menlo, Monaco, "Courier New", monospace' },
  { id: 'fira-code', label: 'Fira Code', value: '"Fira Code", Menlo, Monaco, "Courier New", monospace' }
];

export const ACCENT_PRESETS = {
  teal: {
    label: '青绿',
    light: {
      accentPrimary: '#0f766e',
      accentHover: '#115e59',
      accentSoft: '#ccfbf1',
      focusRing: '#14b8a6',
      glow1: 'rgba(20, 184, 166, 0.08)',
      glow2: 'rgba(56, 189, 248, 0.08)'
    },
    dark: {
      accentPrimary: '#14b8a6',
      accentHover: '#2dd4bf',
      accentSoft: '#123a3a',
      focusRing: '#2dd4bf',
      glow1: 'rgba(20, 184, 166, 0.12)',
      glow2: 'rgba(56, 189, 248, 0.12)'
    }
  },
  blue: {
    label: '科技蓝',
    light: {
      accentPrimary: '#2563eb',
      accentHover: '#1d4ed8',
      accentSoft: '#dbeafe',
      focusRing: '#3b82f6',
      glow1: 'rgba(37, 99, 235, 0.08)',
      glow2: 'rgba(14, 165, 233, 0.08)'
    },
    dark: {
      accentPrimary: '#60a5fa',
      accentHover: '#93c5fd',
      accentSoft: '#1e3a8a',
      focusRing: '#60a5fa',
      glow1: 'rgba(59, 130, 246, 0.12)',
      glow2: 'rgba(14, 165, 233, 0.12)'
    }
  },
  emerald: {
    label: '翡翠绿',
    light: {
      accentPrimary: '#059669',
      accentHover: '#047857',
      accentSoft: '#d1fae5',
      focusRing: '#10b981',
      glow1: 'rgba(16, 185, 129, 0.08)',
      glow2: 'rgba(110, 231, 183, 0.08)'
    },
    dark: {
      accentPrimary: '#34d399',
      accentHover: '#6ee7b7',
      accentSoft: '#064e3b',
      focusRing: '#34d399',
      glow1: 'rgba(16, 185, 129, 0.12)',
      glow2: 'rgba(110, 231, 183, 0.12)'
    }
  },
  amber: {
    label: '琥珀橙',
    light: {
      accentPrimary: '#d97706',
      accentHover: '#b45309',
      accentSoft: '#fef3c7',
      focusRing: '#f59e0b',
      glow1: 'rgba(245, 158, 11, 0.08)',
      glow2: 'rgba(251, 191, 36, 0.08)'
    },
    dark: {
      accentPrimary: '#f59e0b',
      accentHover: '#fbbf24',
      accentSoft: '#78350f',
      focusRing: '#fbbf24',
      glow1: 'rgba(245, 158, 11, 0.12)',
      glow2: 'rgba(251, 191, 36, 0.12)'
    }
  },
  purple: {
    label: '星夜紫',
    light: {
      accentPrimary: '#7c3aed',
      accentHover: '#6d28d9',
      accentSoft: '#ede9fe',
      focusRing: '#8b5cf6',
      glow1: 'rgba(124, 58, 237, 0.08)',
      glow2: 'rgba(139, 92, 246, 0.08)'
    },
    dark: {
      accentPrimary: '#a78bfa',
      accentHover: '#c4b5fd',
      accentSoft: '#3b2a67',
      focusRing: '#a78bfa',
      glow1: 'rgba(167, 139, 250, 0.14)',
      glow2: 'rgba(196, 181, 253, 0.12)'
    }
  }
};

export function resolveTheme(themeMode, explicitTheme) {
  if (themeMode === 'light' || themeMode === 'dark') {
    return themeMode;
  }

  if (themeMode === 'system') {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    if (explicitTheme === 'light' || explicitTheme === 'dark') {
      return explicitTheme;
    }
    return 'light';
  }

  if (explicitTheme === 'light' || explicitTheme === 'dark') {
    return explicitTheme;
  }

  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  return 'light';
}

export function applyAppearanceSettings(settings) {
  if (typeof document === 'undefined') {
    return;
  }

  const root = document.documentElement;
  const accent = ACCENT_PRESETS[settings.accent_color] || ACCENT_PRESETS.teal;
  const mode = settings.theme === 'dark' ? 'dark' : 'light';
  const accentValues = accent[mode];

  root.style.setProperty('--accent-primary', accentValues.accentPrimary);
  root.style.setProperty('--accent-hover', accentValues.accentHover);
  root.style.setProperty('--accent-soft', accentValues.accentSoft);
  root.style.setProperty('--focus-ring', accentValues.focusRing);
  root.style.setProperty('--bg-glow-1', accentValues.glow1);
  root.style.setProperty('--bg-glow-2', accentValues.glow2);
  const appFontSize = Number(settings.font_size) || 14;
  const terminalFontSize = Number(settings.terminal_font_size) || 14;

  root.style.setProperty('--app-font-family', settings.font_family || '"Avenir Next", "SF Pro Text", sans-serif');
  root.style.setProperty('--app-font-size', `${appFontSize}px`);
  root.style.setProperty('--terminal-font-family', settings.terminal_font_family || 'Menlo, Monaco, "Courier New", monospace');
  root.style.setProperty('--terminal-font-size', `${terminalFontSize}px`);
  root.style.fontSize = `${appFontSize}px`;
  root.setAttribute('data-compact', settings.compact_mode ? 'true' : 'false');
  root.setAttribute('data-reduced-motion', settings.reduced_motion ? 'true' : 'false');

  window.dispatchEvent(new CustomEvent('app:appearance-updated', { detail: settings }));
}

export function getDefaultAppSettings() {
  return {
    theme: 'light',
    theme_mode: 'system',
    use_system_theme: true,
    font_family: FONT_PRESETS[0].value,
    font_size: 14,
    terminal_theme: 'default',
    terminal_font_family: TERMINAL_FONT_PRESETS[0].value,
    terminal_font_size: 14,
    accent_color: 'teal',
    compact_mode: false,
    reduced_motion: false,
    sidebar_width: 300
  };
}
