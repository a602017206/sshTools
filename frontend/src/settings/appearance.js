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
    label: '海盐青',
    light: {
      accentPrimary: '#0f9f9a',
      accentSecondary: '#14b8a6',
      accentHover: '#0f766e',
      accentSoft: '#ccfbf1',
      focusRing: '#2dd4bf',
      glow1: 'rgba(148, 163, 184, 0.1)',
      glow2: 'rgba(148, 163, 184, 0.06)'
    },
    dark: {
      accentPrimary: '#5eead4',
      accentSecondary: '#2dd4bf',
      accentHover: '#2dd4bf',
      accentSoft: '#123a3a',
      focusRing: '#5eead4',
      glow1: 'rgba(148, 163, 184, 0.08)',
      glow2: 'rgba(71, 85, 105, 0.14)'
    }
  },
  blue: {
    label: '晴空蓝',
    light: {
      accentPrimary: '#2f6df6',
      accentSecondary: '#60a5fa',
      accentHover: '#2457d6',
      accentSoft: '#dbeafe',
      focusRing: '#60a5fa',
      glow1: 'rgba(148, 163, 184, 0.1)',
      glow2: 'rgba(148, 163, 184, 0.06)'
    },
    dark: {
      accentPrimary: '#8ab7ff',
      accentSecondary: '#93c5fd',
      accentHover: '#60a5fa',
      accentSoft: '#1e3a8a',
      focusRing: '#93c5fd',
      glow1: 'rgba(148, 163, 184, 0.08)',
      glow2: 'rgba(71, 85, 105, 0.14)'
    }
  },
  emerald: {
    label: '薄荷绿',
    light: {
      accentPrimary: '#10b981',
      accentSecondary: '#84cc16',
      accentHover: '#059669',
      accentSoft: '#d1fae5',
      focusRing: '#34d399',
      glow1: 'rgba(16, 185, 129, 0.1)',
      glow2: 'rgba(132, 204, 22, 0.08)'
    },
    dark: {
      accentPrimary: '#6ee7b7',
      accentSecondary: '#bef264',
      accentHover: '#34d399',
      accentSoft: '#064e3b',
      focusRing: '#6ee7b7',
      glow1: 'rgba(110, 231, 183, 0.12)',
      glow2: 'rgba(190, 242, 100, 0.08)'
    }
  },
  amber: {
    label: '柚子橙',
    light: {
      accentPrimary: '#f59e0b',
      accentSecondary: '#fb7185',
      accentHover: '#d97706',
      accentSoft: '#fef3c7',
      focusRing: '#fbbf24',
      glow1: 'rgba(245, 158, 11, 0.1)',
      glow2: 'rgba(251, 113, 133, 0.07)'
    },
    dark: {
      accentPrimary: '#fbbf24',
      accentSecondary: '#fda4af',
      accentHover: '#f59e0b',
      accentSoft: '#78350f',
      focusRing: '#fbbf24',
      glow1: 'rgba(251, 191, 36, 0.12)',
      glow2: 'rgba(253, 164, 175, 0.08)'
    }
  },
  purple: {
    label: '莓果紫',
    light: {
      accentPrimary: '#8b5cf6',
      accentSecondary: '#ec4899',
      accentHover: '#7c3aed',
      accentSoft: '#ede9fe',
      focusRing: '#a78bfa',
      glow1: 'rgba(139, 92, 246, 0.1)',
      glow2: 'rgba(236, 72, 153, 0.07)'
    },
    dark: {
      accentPrimary: '#c4b5fd',
      accentSecondary: '#f0abfc',
      accentHover: '#a78bfa',
      accentSoft: '#3b2a67',
      focusRing: '#c4b5fd',
      glow1: 'rgba(196, 181, 253, 0.12)',
      glow2: 'rgba(240, 171, 252, 0.08)'
    }
  },
  sky: {
    label: '湖水蓝',
    light: {
      accentPrimary: '#0ea5e9',
      accentSecondary: '#22d3ee',
      accentHover: '#0284c7',
      accentSoft: '#e0f2fe',
      focusRing: '#38bdf8',
      glow1: 'rgba(14, 165, 233, 0.1)',
      glow2: 'rgba(34, 211, 238, 0.08)'
    },
    dark: {
      accentPrimary: '#7dd3fc',
      accentSecondary: '#67e8f9',
      accentHover: '#38bdf8',
      accentSoft: '#0c4a6e',
      focusRing: '#7dd3fc',
      glow1: 'rgba(125, 211, 252, 0.12)',
      glow2: 'rgba(103, 232, 249, 0.08)'
    }
  },
  rose: {
    label: '樱桃粉',
    light: {
      accentPrimary: '#fb7185',
      accentSecondary: '#f472b6',
      accentHover: '#e11d48',
      accentSoft: '#ffe4e6',
      focusRing: '#fda4af',
      glow1: 'rgba(251, 113, 133, 0.1)',
      glow2: 'rgba(244, 114, 182, 0.08)'
    },
    dark: {
      accentPrimary: '#fda4af',
      accentSecondary: '#f9a8d4',
      accentHover: '#fb7185',
      accentSoft: '#4c1d24',
      focusRing: '#fda4af',
      glow1: 'rgba(253, 164, 175, 0.1)',
      glow2: 'rgba(249, 168, 212, 0.08)'
    }
  },
  coral: {
    label: '珊瑚红',
    light: {
      accentPrimary: '#f97316',
      accentSecondary: '#fb7185',
      accentHover: '#ea580c',
      accentSoft: '#ffedd5',
      focusRing: '#fdba74',
      glow1: 'rgba(249, 115, 22, 0.1)',
      glow2: 'rgba(251, 113, 133, 0.08)'
    },
    dark: {
      accentPrimary: '#fdba74',
      accentSecondary: '#fda4af',
      accentHover: '#fb923c',
      accentSoft: '#431407',
      focusRing: '#fdba74',
      glow1: 'rgba(253, 186, 116, 0.1)',
      glow2: 'rgba(253, 164, 175, 0.08)'
    }
  },
  lime: {
    label: '青柠绿',
    light: {
      accentPrimary: '#84cc16',
      accentSecondary: '#22c55e',
      accentHover: '#65a30d',
      accentSoft: '#ecfccb',
      focusRing: '#a3e635',
      glow1: 'rgba(132, 204, 22, 0.1)',
      glow2: 'rgba(34, 197, 94, 0.08)'
    },
    dark: {
      accentPrimary: '#bef264',
      accentSecondary: '#86efac',
      accentHover: '#a3e635',
      accentSoft: '#365314',
      focusRing: '#bef264',
      glow1: 'rgba(190, 242, 100, 0.1)',
      glow2: 'rgba(134, 239, 172, 0.08)'
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
    return 'dark';
  }

  if (explicitTheme === 'light' || explicitTheme === 'dark') {
    return explicitTheme;
  }

  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  return 'dark';
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
  root.style.setProperty('--accent-secondary', accentValues.accentSecondary || accentValues.accentHover);
  root.style.setProperty('--accent-hover', accentValues.accentHover);
  root.style.setProperty('--accent-soft', accentValues.accentSoft);
  root.style.setProperty('--accent-glow', accentValues.glow1 || (mode === 'dark' ? 'rgba(45, 212, 191, 0.35)' : 'rgba(13, 148, 136, 0.22)'));
  root.style.setProperty('--focus-ring', accentValues.focusRing);
  root.style.setProperty('--ops-signal', accentValues.accentPrimary);
  root.style.setProperty('--border-active', accentValues.accentPrimary);
  // 氛围光固定中性灰，不跟强调色染色（避免蓝/紫染屏）
  if (mode === 'dark') {
    root.style.setProperty('--bg-glow-1', 'rgba(148, 163, 184, 0.08)');
    root.style.setProperty('--bg-glow-2', 'rgba(71, 85, 105, 0.14)');
  } else {
    root.style.setProperty('--bg-glow-1', 'rgba(148, 163, 184, 0.1)');
    root.style.setProperty('--bg-glow-2', 'rgba(148, 163, 184, 0.06)');
  }
  const appFontSize = Number(settings.font_size) || 14;
  const terminalFontSize = Number(settings.terminal_font_size) || 14;

  root.style.setProperty('--app-font-family', settings.font_family || '"Avenir Next", "SF Pro Text", sans-serif');
  root.style.setProperty('--app-font-size', `${appFontSize}px`);
  root.style.setProperty('--terminal-font-family', settings.terminal_font_family || 'Menlo, Monaco, "Courier New", monospace');
  root.style.setProperty('--terminal-font-size', `${terminalFontSize}px`);
  root.style.fontSize = `${appFontSize}px`;
  root.setAttribute('data-compact', settings.compact_mode ? 'true' : 'false');
  root.setAttribute('data-reduced-motion', settings.reduced_motion ? 'true' : 'false');

  applyBackgroundImage(settings);

  window.dispatchEvent(new CustomEvent('app:appearance-updated', { detail: settings }));
}

export function applyBackgroundImage(settings = {}) {
  if (typeof document === 'undefined') {
    return;
  }
  const root = document.documentElement;
  const enabled = Boolean(settings.background_image_enabled && settings.background_image_data_url);
  const fit = settings.background_image_fit === 'contain' ? 'contain' : 'cover';
  const opacity = Math.max(0, Math.min(100, Number(settings.background_image_opacity) || 0));

  if (enabled) {
    root.setAttribute('data-bg-image', 'true');
    root.style.setProperty('--ops-bg-image', `url("${settings.background_image_data_url}")`);
    root.style.setProperty('--ops-bg-image-fit', fit);
    root.style.setProperty('--ops-bg-image-opacity', String(opacity / 100));
  } else {
    root.removeAttribute('data-bg-image');
    root.style.removeProperty('--ops-bg-image');
    root.style.removeProperty('--ops-bg-image-fit');
    root.style.removeProperty('--ops-bg-image-opacity');
  }
}

export function getDefaultAppSettings() {
  return {
    theme: 'dark',
    theme_mode: 'dark',
    use_system_theme: false,
    font_family: FONT_PRESETS[0].value,
    font_size: 14,
    terminal_theme: 'default',
    terminal_font_family: TERMINAL_FONT_PRESETS[1].value,
    terminal_font_size: 14,
    accent_color: 'teal',
    compact_mode: false,
    reduced_motion: false,
    sidebar_width: 260,
    background_image_enabled: false,
    background_image_path: '',
    background_image_fit: 'cover',
    background_image_opacity: 35,
    background_image_data_url: ''
  };
}
