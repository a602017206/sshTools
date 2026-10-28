import {
  applyAppearance,
  readThemeState,
  THEME_PRESETS,
  BG_TOKENS,
  FG_TOKENS,
  ACCENT_TOKENS,
  THEME_IDS,
} from './theme.js';
import {
  CURSOR_STYLES,
  applyCursorStyle,
  cursorLabel,
  readStoredCursor,
} from './cursor.js';
import { applyI18n, readStoredLang, setStoredLang, MESSAGES, t } from './i18n.js';

function swatchStyle(color) {
  return `--dot-color:${color}`;
}

function syncCursorControls(root, cursorId) {
  root.querySelectorAll('[data-set-cursor]').forEach((btn) => {
    const active = btn.getAttribute('data-set-cursor') === cursorId;
    btn.classList.toggle('is-active', active);
    btn.setAttribute('aria-pressed', active ? 'true' : 'false');
  });
}

function syncThemeControls(root, appearance) {
  root.querySelectorAll('[data-theme-mode]').forEach((btn) => {
    const active = btn.getAttribute('data-theme-mode') === appearance.mode;
    btn.classList.toggle('is-active', active);
    btn.setAttribute('aria-pressed', active ? 'true' : 'false');
  });
  root.querySelectorAll('[data-set-theme]').forEach((btn) => {
    const active = btn.getAttribute('data-set-theme') === appearance.preset;
    btn.classList.toggle('is-active', active);
    btn.setAttribute('aria-pressed', active ? 'true' : 'false');
  });
  root.querySelectorAll('[data-set-bg]').forEach((btn) => {
    const active = btn.getAttribute('data-set-bg') === appearance.bg;
    btn.classList.toggle('is-active', active);
    btn.setAttribute('aria-pressed', active ? 'true' : 'false');
  });
  root.querySelectorAll('[data-set-fg]').forEach((btn) => {
    const active = btn.getAttribute('data-set-fg') === appearance.fg;
    btn.classList.toggle('is-active', active);
    btn.setAttribute('aria-pressed', active ? 'true' : 'false');
  });
  root.querySelectorAll('[data-set-accent]').forEach((btn) => {
    const active = btn.getAttribute('data-set-accent') === appearance.accent;
    btn.classList.toggle('is-active', active);
    btn.setAttribute('aria-pressed', active ? 'true' : 'false');
  });

  root.querySelectorAll('[data-theme-pane]').forEach((pane) => {
    pane.hidden = pane.getAttribute('data-theme-pane') !== appearance.mode;
  });
}

function cursorButtons(lang, group) {
  return CURSOR_STYLES
    .filter((style) => style.group === group)
    .map((style) => {
      const label = cursorLabel(style, lang);
      return `<button type="button" class="cursor-pick" data-set-cursor="${style.id}" aria-label="${label}" title="${label}" aria-pressed="false"><span>${style.glyph}</span></button>`;
    })
    .join('');
}

function buildThemePanel(lang) {
  const panel = document.createElement('div');
  panel.className = 'theme-panel';
  panel.dataset.themePanel = 'true';

  const toggle = document.createElement('button');
  toggle.type = 'button';
  toggle.className = 'button theme-panel__toggle';
  toggle.dataset.themeToggle = 'true';
  toggle.setAttribute('aria-expanded', 'false');
  toggle.textContent = t(MESSAGES, lang, 'theme.open');

  const pop = document.createElement('div');
  pop.className = 'theme-panel__pop';
  pop.hidden = true;

  const modes = [
    ['preset', 'theme.mode.preset'],
    ['duo', 'theme.mode.duo'],
    ['trio', 'theme.mode.trio'],
  ].map(([id, key]) => (
    `<button type="button" class="theme-mode" data-theme-mode="${id}" aria-pressed="false">${t(MESSAGES, lang, key)}</button>`
  )).join('');

  const presets = Object.entries(THEME_PRESETS).map(([id, preset]) => {
    const accent = ACCENT_TOKENS[preset.accent].accent;
    return `<button type="button" class="theme-swatch" data-set-theme="${id}" aria-label="${preset.label}" style="${swatchStyle(accent)}" title="${preset.label}"></button>`;
  }).join('');

  const bgs = Object.entries(BG_TOKENS).map(([id, token]) => (
    `<button type="button" class="theme-swatch" data-set-bg="${id}" aria-label="${id}" style="${swatchStyle(token.bg)}" title="${id}"></button>`
  )).join('');

  const fgs = Object.entries(FG_TOKENS).map(([id, token]) => (
    `<button type="button" class="theme-swatch theme-swatch--fg" data-set-fg="${id}" aria-label="${id}" style="${swatchStyle(token.fg)}" title="${id}"></button>`
  )).join('');

  const accents = Object.entries(ACCENT_TOKENS).map(([id, token]) => (
    `<button type="button" class="theme-swatch" data-set-accent="${id}" aria-label="${id}" style="${swatchStyle(token.accent)}" title="${id}"></button>`
  )).join('');

  pop.innerHTML = `
    <div class="theme-panel__modes" role="group" aria-label="${t(MESSAGES, lang, 'theme.mode')}">${modes}</div>
    <div class="theme-panel__pane" data-theme-pane="preset">
      <div class="theme-panel__label" data-i18n="theme.presets">${t(MESSAGES, lang, 'theme.presets')}</div>
      <div class="theme-swatches">${presets}</div>
    </div>
    <div class="theme-panel__pane" data-theme-pane="duo" hidden>
      <div class="theme-panel__label" data-i18n="theme.bg">${t(MESSAGES, lang, 'theme.bg')}</div>
      <div class="theme-swatches">${bgs}</div>
      <div class="theme-panel__label" data-i18n="theme.fg">${t(MESSAGES, lang, 'theme.fg')}</div>
      <div class="theme-swatches">${fgs}</div>
    </div>
    <div class="theme-panel__pane" data-theme-pane="trio" hidden>
      <div class="theme-panel__label" data-i18n="theme.bg">${t(MESSAGES, lang, 'theme.bg')}</div>
      <div class="theme-swatches">${bgs}</div>
      <div class="theme-panel__label" data-i18n="theme.fg">${t(MESSAGES, lang, 'theme.fg')}</div>
      <div class="theme-swatches">${fgs}</div>
      <div class="theme-panel__label" data-i18n="theme.accent">${t(MESSAGES, lang, 'theme.accent')}</div>
      <div class="theme-swatches">${accents}</div>
    </div>
    <div class="theme-panel__cursor">
      <div class="theme-panel__label" data-i18n="theme.cursor">${t(MESSAGES, lang, 'theme.cursor')}</div>
      <div class="theme-panel__label theme-panel__label--sub" data-i18n="theme.cursor.basic">${t(MESSAGES, lang, 'theme.cursor.basic')}</div>
      <div class="cursor-picks">${cursorButtons(lang, 'basic')}</div>
      <div class="theme-panel__label theme-panel__label--sub" data-i18n="theme.cursor.pet">${t(MESSAGES, lang, 'theme.cursor.pet')}</div>
      <div class="cursor-picks">${cursorButtons(lang, 'pet')}</div>
      <div class="theme-panel__label theme-panel__label--sub" data-i18n="theme.cursor.zodiac">${t(MESSAGES, lang, 'theme.cursor.zodiac')}</div>
      <div class="cursor-picks">${cursorButtons(lang, 'zodiac')}</div>
    </div>
  `;

  panel.append(toggle, pop);
  return panel;
}

function refreshThemePanelLabels(lang) {
  document.querySelectorAll('[data-theme-toggle]').forEach((el) => {
    el.textContent = t(MESSAGES, lang, 'theme.open');
  });
  document.querySelectorAll('[data-theme-mode]').forEach((el) => {
    const mode = el.getAttribute('data-theme-mode');
    el.textContent = t(MESSAGES, lang, `theme.mode.${mode}`);
  });
  [
    'theme.presets',
    'theme.bg',
    'theme.fg',
    'theme.accent',
    'theme.cursor',
    'theme.cursor.basic',
    'theme.cursor.pet',
    'theme.cursor.zodiac',
  ].forEach((key) => {
    document.querySelectorAll(`[data-i18n="${key}"]`).forEach((el) => {
      el.textContent = t(MESSAGES, lang, key);
    });
  });
  document.querySelectorAll('[data-set-cursor]').forEach((btn) => {
    const style = CURSOR_STYLES.find((item) => item.id === btn.getAttribute('data-set-cursor'));
    if (!style) return;
    const label = cursorLabel(style, lang);
    btn.setAttribute('aria-label', label);
    btn.title = label;
  });
}

export function initChrome({ sceneRef, sceneApi } = {}) {
  const resolvedSceneRef = sceneRef || { api: sceneApi || null };
  let appearance = applyAppearance(readThemeState(), {
    onPalette: (palette) => resolvedSceneRef.api?.setPalette?.(palette),
  });
  let cursorId = applyCursorStyle(readStoredCursor()).id;
  let lang = readStoredLang();
  applyI18n(document, MESSAGES, lang);
  document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';

  document.querySelectorAll('.theme-dots').forEach((legacy) => {
    const panel = buildThemePanel(lang);
    legacy.replaceWith(panel);
  });

  const applyPartial = (partial) => {
    appearance = applyAppearance(partial, {
      onPalette: (palette) => resolvedSceneRef.api?.setPalette?.(palette),
    });
    syncThemeControls(document, appearance);
  };

  syncThemeControls(document, appearance);
  syncCursorControls(document, cursorId);

  document.querySelectorAll('[data-theme-toggle]').forEach((btn) => {
    btn.addEventListener('click', (event) => {
      event.stopPropagation();
      const panel = btn.closest('.theme-panel');
      const pop = panel?.querySelector('.theme-panel__pop');
      if (!pop) return;
      const open = pop.hidden;
      document.querySelectorAll('.theme-panel__pop').forEach((el) => { el.hidden = true; });
      document.querySelectorAll('[data-theme-toggle]').forEach((el) => el.setAttribute('aria-expanded', 'false'));
      if (open) {
        pop.hidden = false;
        btn.setAttribute('aria-expanded', 'true');
      }
    });
  });

  document.addEventListener('click', (event) => {
    if (event.target.closest?.('.theme-panel')) return;
    document.querySelectorAll('.theme-panel__pop').forEach((el) => { el.hidden = true; });
    document.querySelectorAll('[data-theme-toggle]').forEach((el) => el.setAttribute('aria-expanded', 'false'));
  });

  document.querySelectorAll('[data-theme-mode]').forEach((btn) => {
    btn.addEventListener('click', () => {
      applyPartial({ mode: btn.getAttribute('data-theme-mode') });
    });
  });

  document.querySelectorAll('[data-set-theme]').forEach((btn) => {
    btn.addEventListener('click', () => {
      applyPartial({ mode: 'preset', preset: btn.getAttribute('data-set-theme') });
    });
  });

  document.querySelectorAll('[data-set-bg]').forEach((btn) => {
    btn.addEventListener('click', () => {
      applyPartial({ bg: btn.getAttribute('data-set-bg') });
    });
  });

  document.querySelectorAll('[data-set-fg]').forEach((btn) => {
    btn.addEventListener('click', () => {
      applyPartial({ fg: btn.getAttribute('data-set-fg') });
    });
  });

  document.querySelectorAll('[data-set-accent]').forEach((btn) => {
    btn.addEventListener('click', () => {
      applyPartial({ mode: 'trio', accent: btn.getAttribute('data-set-accent') });
    });
  });

  document.querySelectorAll('[data-set-cursor]').forEach((btn) => {
    btn.addEventListener('click', () => {
      cursorId = applyCursorStyle(btn.getAttribute('data-set-cursor')).id;
      syncCursorControls(document, cursorId);
    });
  });

  document.querySelectorAll('[data-set-lang]').forEach((btn) => {
    btn.addEventListener('click', () => {
      lang = setStoredLang(btn.getAttribute('data-set-lang'));
      applyI18n(document, MESSAGES, lang);
      document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';
      refreshThemePanelLabels(lang);
    });
  });

  const page = document.body.dataset.page;
  document.querySelectorAll('[data-nav]').forEach((link) => {
    if (link.getAttribute('data-nav') === page) link.setAttribute('aria-current', 'page');
  });

  return {
    theme: appearance.preset,
    appearance,
    cursor: cursorId,
    lang,
    THEME_IDS,
    sceneRef: resolvedSceneRef,
  };
}
