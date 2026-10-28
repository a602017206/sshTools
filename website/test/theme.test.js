// website/test/theme.test.js
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  normalizeThemeId,
  readStoredTheme,
  applyTheme,
  THEME_STORAGE_KEY,
  THEME_PALETTES,
} from '../assets/js/theme.js';

describe('normalizeThemeId', () => {
  it('passes through known ids', () => {
    assert.equal(normalizeThemeId('amber'), 'amber');
  });
  it('falls back to graphite', () => {
    assert.equal(normalizeThemeId('nope'), 'graphite');
    assert.equal(normalizeThemeId(''), 'graphite');
  });
});

describe('readStoredTheme', () => {
  it('reads storage', () => {
    const storage = { getItem: () => 'paper' };
    assert.equal(readStoredTheme(storage), 'paper');
  });
  it('falls back when missing', () => {
    const storage = { getItem: () => null };
    assert.equal(readStoredTheme(storage), 'graphite');
  });
});

describe('applyTheme', () => {
  it('sets data-theme, storage, and palette callback', () => {
    const el = { setAttribute: (k, v) => { el[k] = v; } };
    const store = { setItem() {}, values: {} };
    store.setItem = (k, v) => { store.values[k] = v; };
    let palette;
    applyTheme('aurora', {
      documentElement: el,
      storage: store,
      onPalette: (p) => { palette = p; },
    });
    assert.equal(el['data-theme'], 'aurora');
    assert.equal(store.values[THEME_STORAGE_KEY], 'aurora');
    assert.deepEqual(palette, THEME_PALETTES.aurora);
  });
});
