const THREE_URL = 'https://cdn.jsdelivr.net/npm/three@0.170.0/build/three.module.js';
const DEFAULT_PALETTE = {
  bg: '#0b1214',
  accent: '#2dd4bf',
  particle: '#5eead4',
  line: '#2dd4bf66',
};
const MAX_EDGES = 200;

export function shouldStartScene({ matchMedia, webglAvailable }) {
  const reducedMotion = matchMedia?.('(prefers-reduced-motion: reduce)')?.matches;
  return Boolean(webglAvailable) && !reducedMotion;
}

export function parseCssColor(value) {
  const color = String(value ?? '');
  const match = /^#([0-9a-f]{6})([0-9a-f]{2})$/i.exec(color);
  if (!match) return { color, opacity: 1 };

  return {
    color: `#${match[1]}`,
    opacity: Number.parseInt(match[2], 16) / 255,
  };
}

/**
 * Advance particle positions one frame. Clamps into bounds and turns both
 * current and base velocities inward on bounce so drift never escapes.
 */
export function stepParticles({
  positions,
  velocities,
  baseVelocities,
  count,
  bounds,
  pointer = { active: false },
}) {
  for (let index = 0; index < count; index += 1) {
    const offset = index * 3;
    const velocityOffset = index * 2;
    let vx = velocities[velocityOffset];
    let vy = velocities[velocityOffset + 1];

    if (pointer.active) {
      const dx = pointer.worldX - positions[offset];
      const dy = pointer.worldY - positions[offset + 1];
      const distanceSquared = dx * dx + dy * dy;
      if (distanceSquared < 16 && distanceSquared > 0.01) {
        const force = Math.min(0.003, 0.02 / distanceSquared);
        vx += dx * force;
        vy += dy * force;
      }
    }

    vx += (baseVelocities[velocityOffset] - vx) * 0.006;
    vy += (baseVelocities[velocityOffset + 1] - vy) * 0.006;
    vx *= 0.995;
    vy *= 0.995;
    vx = Math.max(-0.025, Math.min(0.025, vx));
    vy = Math.max(-0.025, Math.min(0.025, vy));

    let x = positions[offset] + vx;
    let y = positions[offset + 1] + vy;

    if (x > bounds.x) {
      x = bounds.x;
      vx = -Math.abs(vx) || -Math.abs(baseVelocities[velocityOffset]) || -0.003;
      baseVelocities[velocityOffset] = -Math.abs(baseVelocities[velocityOffset]) || -0.003;
    } else if (x < -bounds.x) {
      x = -bounds.x;
      vx = Math.abs(vx) || Math.abs(baseVelocities[velocityOffset]) || 0.003;
      baseVelocities[velocityOffset] = Math.abs(baseVelocities[velocityOffset]) || 0.003;
    }

    if (y > bounds.y) {
      y = bounds.y;
      vy = -Math.abs(vy) || -Math.abs(baseVelocities[velocityOffset + 1]) || -0.003;
      baseVelocities[velocityOffset + 1] = -Math.abs(baseVelocities[velocityOffset + 1]) || -0.003;
    } else if (y < -bounds.y) {
      y = -bounds.y;
      vy = Math.abs(vy) || Math.abs(baseVelocities[velocityOffset + 1]) || 0.003;
      baseVelocities[velocityOffset + 1] = Math.abs(baseVelocities[velocityOffset + 1]) || 0.003;
    }

    velocities[velocityOffset] = vx;
    velocities[velocityOffset + 1] = vy;
    positions[offset] = x;
    positions[offset + 1] = y;
  }
}

export function mapPointerToWorld({ clientX, clientY, viewport, fov, cameraZ }) {
  const width = Math.max(1, viewport?.width || 1);
  const height = Math.max(1, viewport?.height || 1);
  const normalizedX = ((clientX - (viewport?.left || 0)) / width) * 2 - 1;
  const normalizedY = 1 - ((clientY - (viewport?.top || 0)) / height) * 2;
  const visibleHeight = 2 * Math.tan((fov * Math.PI) / 360) * cameraZ;

  return {
    x: normalizedX * (visibleHeight * width / height) / 2,
    y: normalizedY * visibleHeight / 2,
  };
}

function detectWebGL() {
  try {
    const canvas = document.createElement('canvas');
    return Boolean(canvas.getContext('webgl') || canvas.getContext('experimental-webgl'));
  } catch {
    return false;
  }
}

/**
 * Mount an animated particle field on a canvas.
 *
 * @param {HTMLCanvasElement | { canvas: HTMLCanvasElement, palette?: object, density?: string }} target
 * @param {{ palette?: object, density?: string }} options
 */
export async function mountScene(target, options = {}) {
  const config = target?.canvas ? target : { ...options, canvas: target };
  const { canvas, density = 'default' } = config;
  let palette = { ...DEFAULT_PALETTE, ...config.palette };

  if (!canvas) {
    throw new TypeError('mountScene requires a canvas');
  }
  if (!detectWebGL()) {
    return null;
  }

  const THREE = await import(THREE_URL);
  let renderer;
  let animationFrame = 0;
  let destroyed = false;
  let paused = false;

  const count = density === 'light' ? 50 : 120;
  const positions = new Float32Array(count * 3);
  const velocities = new Float32Array(count * 2);
  const baseVelocities = new Float32Array(count * 2);
  const linePositions = new Float32Array(MAX_EDGES * 6);
  const pointer = { worldX: 0, worldY: 0, active: false };
  const bounds = { x: 10, y: 6 };

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(55, 1, 0.1, 100);
  const particleGeometry = new THREE.BufferGeometry();
  const lineGeometry = new THREE.BufferGeometry();
  const initialLineColor = parseCssColor(palette.line);
  const particleMaterial = new THREE.PointsMaterial({ color: palette.particle, size: 0.07, transparent: true, opacity: 0.85 });
  const lineMaterial = new THREE.LineBasicMaterial({
    color: initialLineColor.color,
    transparent: initialLineColor.opacity < 1,
    opacity: initialLineColor.opacity,
  });
  const particles = new THREE.Points(particleGeometry, particleMaterial);
  const lines = new THREE.LineSegments(lineGeometry, lineMaterial);

  particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
  lineGeometry.setAttribute('position', new THREE.BufferAttribute(linePositions, 3));
  lineGeometry.setDrawRange(0, 0);
  scene.add(particles, lines);
  camera.position.z = 16;

  function scatterParticles() {
    for (let index = 0; index < count; index += 1) {
      const offset = index * 3;
      const velocityOffset = index * 2;
      const angle = Math.random() * Math.PI * 2;
      const speed = 0.003 + Math.random() * 0.005;
      positions[offset] = (Math.random() - 0.5) * bounds.x * 2;
      positions[offset + 1] = (Math.random() - 0.5) * bounds.y * 2;
      positions[offset + 2] = (Math.random() - 0.5) * 2;
      baseVelocities[velocityOffset] = Math.cos(angle) * speed;
      baseVelocities[velocityOffset + 1] = Math.sin(angle) * speed;
      velocities[velocityOffset] = baseVelocities[velocityOffset];
      velocities[velocityOffset + 1] = baseVelocities[velocityOffset + 1];
    }
    particleGeometry.attributes.position.needsUpdate = true;
  }

  function clampParticlesToBounds() {
    for (let index = 0; index < count; index += 1) {
      const offset = index * 3;
      positions[offset] = Math.max(-bounds.x, Math.min(bounds.x, positions[offset]));
      positions[offset + 1] = Math.max(-bounds.y, Math.min(bounds.y, positions[offset + 1]));
    }
    particleGeometry.attributes.position.needsUpdate = true;
  }

  function resize() {
    const { width, height } = canvas.getBoundingClientRect();
    const safeWidth = Math.max(1, width || canvas.clientWidth || 1);
    const safeHeight = Math.max(1, height || canvas.clientHeight || 1);
    camera.aspect = safeWidth / safeHeight;
    camera.updateProjectionMatrix();
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.setSize(safeWidth, safeHeight, false);
    bounds.y = Math.tan((camera.fov * Math.PI) / 360) * camera.position.z;
    bounds.x = bounds.y * camera.aspect;
    clampParticlesToBounds();
  }

  function rebuildLines() {
    let edges = 0;
    const maxDistanceSquared = 2.2 ** 2;

    for (let from = 0; from < count && edges < MAX_EDGES; from += 1) {
      const fromOffset = from * 3;
      for (let to = from + 1; to < count && edges < MAX_EDGES; to += 1) {
        const toOffset = to * 3;
        const dx = positions[fromOffset] - positions[toOffset];
        const dy = positions[fromOffset + 1] - positions[toOffset + 1];
        if (dx * dx + dy * dy > maxDistanceSquared) continue;

        const lineOffset = edges * 6;
        linePositions.set(positions.subarray(fromOffset, fromOffset + 3), lineOffset);
        linePositions.set(positions.subarray(toOffset, toOffset + 3), lineOffset + 3);
        edges += 1;
      }
    }

    lineGeometry.setDrawRange(0, edges * 2);
    lineGeometry.attributes.position.needsUpdate = true;
  }

  function updateParticles() {
    stepParticles({
      positions,
      velocities,
      baseVelocities,
      count,
      bounds,
      pointer,
    });
    particleGeometry.attributes.position.needsUpdate = true;
  }

  function render() {
    if (destroyed || paused) return;
    updateParticles();
    rebuildLines();
    renderer.render(scene, camera);
    animationFrame = requestAnimationFrame(render);
  }

  function onPointerMove(event) {
    const mapped = mapPointerToWorld({
      clientX: event.clientX,
      clientY: event.clientY,
      viewport: canvas.getBoundingClientRect(),
      fov: camera.fov,
      cameraZ: camera.position.z,
    });
    pointer.worldX = mapped.x;
    pointer.worldY = mapped.y;
    pointer.active = true;
  }

  function setPalette(nextPalette) {
    palette = { ...palette, ...nextPalette };
    const lineColor = parseCssColor(palette.line);
    particleMaterial.color.set(palette.particle);
    lineMaterial.color.set(lineColor.color);
    lineMaterial.opacity = lineColor.opacity;
    lineMaterial.transparent = lineColor.opacity < 1;
    renderer.setClearColor(palette.bg, 0);
  }

  function pause() {
    if (paused || destroyed) return;
    paused = true;
    cancelAnimationFrame(animationFrame);
  }

  function resume() {
    if (!paused || destroyed) return;
    paused = false;
    animationFrame = requestAnimationFrame(render);
  }

  function onVisibilityChange() {
    if (document.hidden) pause();
    else resume();
  }

  function destroy() {
    if (destroyed) return;
    destroyed = true;
    cancelAnimationFrame(animationFrame);
    window.removeEventListener('pointermove', onPointerMove);
    window.removeEventListener('resize', resize);
    document.removeEventListener('visibilitychange', onVisibilityChange);
    particleGeometry.dispose();
    lineGeometry.dispose();
    particleMaterial.dispose();
    lineMaterial.dispose();
    renderer?.dispose();
  }

  try {
    renderer = new THREE.WebGLRenderer({ canvas, alpha: true, antialias: true });
    setPalette(palette);
    resize();
    scatterParticles();
    window.addEventListener('pointermove', onPointerMove, { passive: true });
    window.addEventListener('resize', resize, { passive: true });
    document.addEventListener('visibilitychange', onVisibilityChange);
    if (document.visibilityState === 'hidden') pause();
    else animationFrame = requestAnimationFrame(render);

    return { setPalette, destroy, pause, resume };
  } catch (error) {
    destroy();
    throw error;
  }
}
