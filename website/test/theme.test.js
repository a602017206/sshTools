import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  normalizeThemeId,
  normalizeThemeMode,
  normalizeBgId,
  normalizeFgId,
  normalizeAccentId,
  readStoredTheme,
  readThemeState,
  resolveAppearance,
  resolveTokens,
  applyTheme,
  applyAppearance,
  THEME_STORAGE_KEY,
  THEME_MODE_STORAGE_KEY,
  THEME_BG_STORAGE_KEY,
  THEME_FG_STORAGE_KEY,
  THEME_ACCENT_STORAGE_KEY,
  THEME_PALETTES,
  THEME_PRESETS,
  THEME_IDS,
} from '../assets/js/theme.js';

function mockStorage(initial = {}) {
  const values = { ...initial };
  return {
    values,
    getItem: (key) => (Object.prototype.hasOwnProperty.call(values, key) ? values[key] : null),
    setItem: (key, value) => { values[key] = String(value); },
  };
}

function mockRoot() {
  const props = {};
  const style = { setProperty: (k, v) => { props[k] = v; } };
  return {
    props,
    style,
    setAttribute: (k, v) => { props[k] = v; },
  };
}

describe('normalize helpers', () => {
  it('normalizes preset / mode / token ids', () => {
    assert.equal(normalizeThemeId('amber'), 'amber');
    assert.equal(normalizeThemeId('nope'), 'graphite');
    assert.equal(normalizeThemeMode('duo'), 'duo');
    assert.equal(normalizeThemeMode('x'), 'preset');
    assert.equal(normalizeBgId('oceanDeep'), 'oceanDeep');
    assert.equal(normalizeBgId('x'), 'graphite');
    assert.equal(normalizeFgId('ink'), 'ink');
    assert.equal(normalizeFgId('x'), 'snow');
    assert.equal(normalizeAccentId('violet'), 'violet');
    assert.equal(normalizeAccentId('x'), 'teal');
  });
});

describe('presets', () => {
  it('exposes at least 10 named presets and matching palettes', () => {
    assert.ok(THEME_IDS.length >= 10);
    assert.equal(Object.keys(THEME_PRESETS).length, THEME_IDS.length);
    for (const id of THEME_IDS) {
      assert.ok(THEME_PALETTES[id]?.bg);
      assert.ok(THEME_PALETTES[id]?.fg);
      assert.ok(THEME_PALETTES[id]?.accent);
    }
  });
});

describe('resolveAppearance', () => {
  it('resolves preset combinations', () => {
    const appearance = resolveAppearance({ mode: 'preset', preset: 'forest' });
    assert.equal(appearance.mode, 'preset');
    assert.equal(appearance.bg, 'forestDeep');
    assert.equal(appearance.fg, 'tealTint');
    assert.equal(appearance.accent, 'lime');
    assert.deepEqual(appearance.tokens, resolveTokens('forestDeep', 'tealTint', 'lime'));
  });

  it('resolves duo bg+fg with default accent', () => {
    const appearance = resolveAppearance({ mode: 'duo', bg: 'cream', fg: 'charcoal' });
    assert.equal(appearance.mode, 'duo');
    assert.equal(appearance.bg, 'cream');
    assert.equal(appearance.fg, 'charcoal');
    assert.equal(appearance.accent, 'teal');
    assert.equal(appearance.tokens.scheme, 'light');
  });

  it('resolves trio bg+fg+accent', () => {
    const appearance = resolveAppearance({
      mode: 'trio', bg: 'midnight', fg: 'ice', accent: 'rose',
    });
    assert.equal(appearance.mode, 'trio');
    assert.equal(appearance.accent, 'rose');
    assert.equal(appearance.tokens.accent, THEME_PALETTES.noir.accent);
  });
});

describe('readThemeState / readStoredTheme', () => {
  it('reads storage with fallbacks', () => {
    const storage = mockStorage({
      [THEME_MODE_STORAGE_KEY]: 'trio',
      [THEME_STORAGE_KEY]: 'orchid',
      [THEME_BG_STORAGE_KEY]: 'ink',
      [THEME_FG_STORAGE_KEY]: 'mist',
      [THEME_ACCENT_STORAGE_KEY]: 'violet',
    });
    assert.deepEqual(readThemeState(storage), {
      mode: 'trio',
      preset: 'orchid',
      bg: 'ink',
      fg: 'mist',
      accent: 'violet',
    });
    assert.equal(readStoredTheme(storage), 'orchid');
  });

  it('falls back when missing', () => {
    assert.equal(readStoredTheme(mockStorage()), 'graphite');
    assert.equal(readThemeState(mockStorage()).mode, 'preset');
  });
});

describe('applyTheme / applyAppearance', () => {
  it('applyTheme keeps preset back-compat', () => {
    const el = mockRoot();
    const store = mockStorage();
    let palette;
    applyTheme('aurora', {
      documentElement: el,
      storage: store,
      onPalette: (p) => { palette = p; },
    });
    assert.equal(el.props['data-theme'], 'aurora');
    assert.equal(el.props['data-theme-mode'], 'preset');
    assert.equal(store.values[THEME_STORAGE_KEY], 'aurora');
    assert.deepEqual(palette, THEME_PALETTES.aurora);
    assert.equal(el.props['--accent'], THEME_PALETTES.aurora.accent);
  });

  it('applyAppearance writes duo state and css tokens', () => {
    const el = mockRoot();
    const store = mockStorage();
    const appearance = applyAppearance(
      { mode: 'duo', bg: 'sand', fg: 'charcoal', accent: 'amber' },
      { documentElement: el, storage: store },
    );
    assert.equal(appearance.mode, 'duo');
    assert.equal(store.values[THEME_MODE_STORAGE_KEY], 'duo');
    assert.equal(store.values[THEME_BG_STORAGE_KEY], 'sand');
    assert.equal(store.values[THEME_FG_STORAGE_KEY], 'charcoal');
    assert.equal(el.props['data-scheme'], 'light');
    assert.equal(el.props['--bg'], appearance.tokens.bg);
    assert.equal(el.props['--fg'], appearance.tokens.fg);
  });
});
