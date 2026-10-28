import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  prefersReducedMotion,
  supportsCustomCursor,
  lerp,
  spawnBurst,
  initFx,
} from '../assets/js/fx.js';

describe('prefersReducedMotion', () => {
  it('returns true when reduced motion is preferred', () => {
    assert.equal(
      prefersReducedMotion((query) => ({ matches: query.includes('prefers-reduced-motion') })),
      true,
    );
  });

  it('returns false otherwise', () => {
    assert.equal(prefersReducedMotion(() => ({ matches: false })), false);
  });
});

describe('supportsCustomCursor', () => {
  it('is false when reduced motion is on', () => {
    assert.equal(
      supportsCustomCursor((query) => ({ matches: query.includes('prefers-reduced-motion') })),
      false,
    );
  });

  it('is true for fine pointers without reduced motion', () => {
    assert.equal(
      supportsCustomCursor((query) => ({ matches: query.includes('pointer: fine') })),
      true,
    );
  });
});

describe('lerp', () => {
  it('interpolates between values', () => {
    assert.equal(lerp(0, 10, 0.5), 5);
    assert.equal(lerp(2, 4, 0), 2);
    assert.equal(lerp(2, 4, 1), 4);
  });
});

describe('spawnBurst', () => {
  it('appends temporary spark nodes', () => {
    const children = [];
    const root = {
      ownerDocument: {
        createElement: () => {
          const el = {
            className: '',
            style: { setProperty() {} },
            setAttribute() {},
            remove() { el.removed = true; },
          };
          return el;
        },
      },
      appendChild(node) { children.push(node); },
    };
    const nodes = spawnBurst(root, 10, 20, { count: 4, ttlMs: 1 });
    assert.equal(nodes.length, 4);
    assert.equal(children.length, 4);
    assert.equal(nodes[0].className, 'cursor-spark');
  });
});

describe('initFx', () => {
  it('marks reveal elements visible when motion is reduced', () => {
    const el = { classList: { values: new Set(), add(v) { this.values.add(v); } } };
    const root = { querySelectorAll: (sel) => (sel === '.reveal' ? [el] : []) };
    initFx({ root, matchMedia: () => ({ matches: true }) });
    assert.ok(el.classList.values.has('is-visible'));
  });
});
