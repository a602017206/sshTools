import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { parseCssColor, shouldStartScene } from '../assets/js/scene.js';

describe('parseCssColor', () => {
  it('separates the RGB and alpha portions of 8-digit hex colors', () => {
    assert.deepEqual(parseCssColor('#2dd4bf66'), {
      color: '#2dd4bf',
      opacity: 0.4,
    });
  });

  it('keeps colors without an alpha channel fully opaque', () => {
    assert.deepEqual(parseCssColor('#2dd4bf'), {
      color: '#2dd4bf',
      opacity: 1,
    });
  });
});

describe('shouldStartScene', () => {
  it('returns false when reduced motion is preferred', () => {
    assert.equal(
      shouldStartScene({
        matchMedia: (query) => ({ matches: query.includes('prefers-reduced-motion') }),
        webglAvailable: true,
      }),
      false,
    );
  });

  it('returns false when WebGL is unavailable', () => {
    assert.equal(
      shouldStartScene({
        matchMedia: () => ({ matches: false }),
        webglAvailable: false,
      }),
      false,
    );
  });

  it('returns true when motion is allowed and WebGL is available', () => {
    assert.equal(
      shouldStartScene({
        matchMedia: () => ({ matches: false }),
        webglAvailable: true,
      }),
      true,
    );
  });
});
