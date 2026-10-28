import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  CURSOR_STYLES,
  CURSOR_STORAGE_KEY,
  normalizeCursorId,
  getCursorStyle,
  readStoredCursor,
  applyCursorStyle,
  cursorLabel,
} from '../assets/js/cursor.js';

describe('cursor catalog', () => {
  it('includes basic FX, pets, and all twelve zodiac animals', () => {
    const groups = Object.fromEntries(['basic', 'pet', 'zodiac'].map((g) => [g, CURSOR_STYLES.filter((s) => s.group === g)]));
    assert.ok(groups.basic.length >= 4);
    assert.ok(groups.pet.some((s) => s.id === 'cat'));
    assert.equal(groups.zodiac.length, 12);
    assert.ok(groups.zodiac.some((s) => s.id === 'dragon'));
    assert.ok(groups.zodiac.some((s) => s.id === 'dog'));
  });
});

describe('normalizeCursorId / getCursorStyle', () => {
  it('normalizes unknown ids to ring', () => {
    assert.equal(normalizeCursorId('cat'), 'cat');
    assert.equal(normalizeCursorId('nope'), 'ring');
    assert.equal(getCursorStyle('tiger').glyph, '🐯');
  });
});

describe('storage + apply', () => {
  it('persists and applies data-cursor', () => {
    const values = {};
    const storage = {
      getItem: (k) => values[k] ?? null,
      setItem: (k, v) => { values[k] = String(v); },
    };
    const root = { attrs: {}, setAttribute(k, v) { this.attrs[k] = v; } };
    assert.equal(readStoredCursor(storage), 'ring');
    const style = applyCursorStyle('dragon', { documentElement: root, storage });
    assert.equal(style.id, 'dragon');
    assert.equal(values[CURSOR_STORAGE_KEY], 'dragon');
    assert.equal(root.attrs['data-cursor'], 'dragon');
    assert.equal(cursorLabel(style, 'zh'), '龙');
    assert.equal(cursorLabel(style, 'en'), 'Dragon');
  });
});
