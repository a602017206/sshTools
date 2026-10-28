import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  mapPointerToWorld,
  parseCssColor,
  shouldStartScene,
  stepParticles,
} from '../assets/js/scene.js';

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
    const point = mapPointerToWorld({ clientX: 500, clientY: 250, viewport, fov: 60, cameraZ: 10 });
    assert.equal(point.x, 0);
    assert.equal(point.y, 0);
    assert.equal(point.ndcX, 0);
    assert.equal(point.ndcY, 0);
  });

  it('uses perspective frustum dimensions to map canvas edges', () => {
    const point = mapPointerToWorld({ clientX: 900, clientY: 50, viewport, fov: 60, cameraZ: 10 });
    const visibleHeight = 2 * Math.tan(Math.PI / 6) * 10;

    assert.ok(Math.abs(point.x - visibleHeight) < 1e-12);
    assert.ok(Math.abs(point.y - visibleHeight / 2) < 1e-12);
    assert.equal(point.ndcX, 1);
    assert.equal(point.ndcY, 1);
  });
});

describe('stepParticles', () => {
  it('keeps particles inside the viewport over many frames', () => {
    const count = 120;
    const bounds = { x: 8, y: 4.5 };
    const positions = new Float32Array(count * 3);
    const velocities = new Float32Array(count * 2);
    const baseVelocities = new Float32Array(count * 2);
    let seed = 42;
    const random = () => {
      seed = (seed * 16807) % 2147483647;
      return (seed - 1) / 2147483646;
    };

    for (let index = 0; index < count; index += 1) {
      const positionOffset = index * 3;
      const velocityOffset = index * 2;
      positions[positionOffset] = (random() * 2 - 1) * bounds.x;
      positions[positionOffset + 1] = (random() * 2 - 1) * bounds.y;
      baseVelocities[velocityOffset] = (random() * 2 - 1) * 0.02;
      baseVelocities[velocityOffset + 1] = (random() * 2 - 1) * 0.02;
      velocities[velocityOffset] = baseVelocities[velocityOffset];
      velocities[velocityOffset + 1] = baseVelocities[velocityOffset + 1];
    }

    const zVelocities = new Float32Array(count);
    const baseZVelocities = new Float32Array(count);
    bounds.z = 6;
    for (let index = 0; index < count; index += 1) {
      positions[index * 3 + 2] = (random() * 2 - 1) * bounds.z;
      baseZVelocities[index] = (random() * 2 - 1) * 0.015;
      zVelocities[index] = baseZVelocities[index];
    }

    for (let frame = 0; frame < 5000; frame += 1) {
      stepParticles({
        positions,
        velocities,
        baseVelocities,
        count,
        bounds,
        pointer: { active: false },
        zVelocities,
        baseZVelocities,
      });

      for (let index = 0; index < count; index += 1) {
        const positionOffset = index * 3;
        assert.ok(Math.abs(positions[positionOffset]) <= bounds.x);
        assert.ok(Math.abs(positions[positionOffset + 1]) <= bounds.y);
        assert.ok(Math.abs(positions[positionOffset + 2]) <= bounds.z);
      }
    }
  });
});
