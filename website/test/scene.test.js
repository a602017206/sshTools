import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { mapPointerToWorld, parseCssColor, shouldStartScene } from '../assets/js/scene.js';

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

describe('mapPointerToWorld', () => {
  const viewport = { left: 100, top: 50, width: 800, height: 400 };

  it('maps the center of the canvas to the origin', () => {
    assert.deepEqual(
      mapPointerToWorld({ clientX: 500, clientY: 250, viewport, fov: 60, cameraZ: 10 }),
      { x: 0, y: 0 },
    );
  });

  it('uses perspective frustum dimensions to map canvas edges', () => {
    const point = mapPointerToWorld({ clientX: 900, clientY: 50, viewport, fov: 60, cameraZ: 10 });
    const visibleHeight = 2 * Math.tan(Math.PI / 6) * 10;

    assert.ok(Math.abs(point.x - visibleHeight) < 1e-12);
    assert.ok(Math.abs(point.y - visibleHeight / 2) < 1e-12);
  });
});
