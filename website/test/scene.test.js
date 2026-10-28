import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { shouldStartScene } from '../assets/js/scene.js';

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
