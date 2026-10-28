import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { prefersReducedMotion, initFx } from '../assets/js/fx.js';

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

describe('initFx', () => {
  it('marks reveal elements visible when motion is reduced', () => {
    const el = { classList: { values: new Set(), add(v) { this.values.add(v); } } };
    const root = { querySelectorAll: (sel) => (sel === '.reveal' ? [el] : []) };
    initFx({ root, matchMedia: () => ({ matches: true }) });
    assert.ok(el.classList.values.has('is-visible'));
  });
});
