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
  it('keeps only ring, comet, and cone shapes', () => {
    assert.deepEqual(CURSOR_STYLES.map((s) => s.id), ['ring', 'comet', 'cone']);
    for (const style of CURSOR_STYLES) {
      assert.ok(CURSOR_SVG_PATHS[style.id], `missing svg for ${style.id}`);
    }
  });
});

describe('normalizeCursorId / svg markup', () => {
  it('falls back to ring and renders stroked svg', () => {
    assert.equal(normalizeCursorId('cone'), 'cone');
    assert.equal(normalizeCursorId('cat'), 'ring');
    assert.equal(normalizeCursorId('nope'), 'ring');
    const svg = cursorSvgMarkup('comet');
    assert.match(svg, /<svg/);
    assert.match(svg, /stroke="currentColor"/);
    assert.equal(getCursorStyle('cone').rotate, true);
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
    const style = applyCursorStyle('cone', { documentElement: root, storage });
    assert.equal(style.id, 'cone');
    assert.equal(values[CURSOR_STORAGE_KEY], 'cone');
    assert.equal(root.attrs['data-cursor'], 'cone');
    assert.equal(cursorLabel(style, 'zh'), '锥形');
    assert.equal(cursorLabel(style, 'en'), 'Cone');
  });
});
