import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  CURSOR_STYLES,
  CURSOR_SVG_PATHS,
  CURSOR_STORAGE_KEY,
  normalizeCursorId,
  getCursorStyle,
  cursorSvgMarkup,
  readStoredCursor,
  applyCursorStyle,
  cursorLabel,
} from '../assets/js/cursor.js';

describe('cursor catalog', () => {
  it('includes neon SVG shapes for basic, pets, and twelve zodiac animals', () => {
    const groups = Object.fromEntries(['basic', 'pet', 'zodiac'].map((g) => [g, CURSOR_STYLES.filter((s) => s.group === g)]));
    assert.ok(groups.basic.length >= 4);
    assert.ok(groups.pet.some((s) => s.id === 'cat'));
    assert.equal(groups.zodiac.length, 12);
    for (const style of CURSOR_STYLES) {
      assert.ok(CURSOR_SVG_PATHS[style.id], `missing svg for ${style.id}`);
      assert.equal(Object.prototype.hasOwnProperty.call(style, 'glyph'), false);
    }
  });
});

describe('normalizeCursorId / svg markup', () => {
  it('normalizes unknown ids to cat and renders line-art svg', () => {
    assert.equal(normalizeCursorId('cat'), 'cat');
    assert.equal(normalizeCursorId('nope'), 'cat');
    const svg = cursorSvgMarkup('tiger');
    assert.match(svg, /<svg/);
    assert.match(svg, /stroke="currentColor"/);
    assert.doesNotMatch(svg, /🐱|🐯|emoji/i);
    assert.ok(getCursorStyle('tiger').trail === 'outline');
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
    assert.equal(readStoredCursor(storage), 'cat');
    const style = applyCursorStyle('dragon', { documentElement: root, storage });
    assert.equal(style.id, 'dragon');
    assert.equal(values[CURSOR_STORAGE_KEY], 'dragon');
    assert.equal(root.attrs['data-cursor'], 'dragon');
    assert.equal(cursorLabel(style, 'zh'), '龙');
    assert.equal(cursorLabel(style, 'en'), 'Dragon');
  });
});
