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
  const linePositions = new Float32Array(MAX_EDGES * 6);
  const pointer = { x: 0, y: 0, active: false };
  const bounds = { x: 10, y: 6 };

  const scene = new THREE.Scene();
  const camera = new THREE.PerspectiveCamera(55, 1, 0.1, 100);
  const particleGeometry = new THREE.BufferGeometry();
  const lineGeometry = new THREE.BufferGeometry();
  const particleMaterial = new THREE.PointsMaterial({ color: palette.particle, size: 0.07, transparent: true, opacity: 0.85 });
  const lineMaterial = new THREE.LineBasicMaterial({ color: palette.line, transparent: true, opacity: 0.45 });
  const particles = new THREE.Points(particleGeometry, particleMaterial);
  const lines = new THREE.LineSegments(lineGeometry, lineMaterial);

  for (let index = 0; index < count; index += 1) {
    const offset = index * 3;
    positions[offset] = (Math.random() - 0.5) * bounds.x * 2;
    positions[offset + 1] = (Math.random() - 0.5) * bounds.y * 2;
    positions[offset + 2] = (Math.random() - 0.5) * 2;
    velocities[index * 2] = (Math.random() - 0.5) * 0.012;
    velocities[index * 2 + 1] = (Math.random() - 0.5) * 0.012;
  }

  particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
  lineGeometry.setAttribute('position', new THREE.BufferAttribute(linePositions, 3));
  lineGeometry.setDrawRange(0, 0);
  scene.add(particles, lines);
  camera.position.z = 16;

  function resize() {
    const { width, height } = canvas.getBoundingClientRect();
    const safeWidth = Math.max(1, width || canvas.clientWidth || 1);
    const safeHeight = Math.max(1, height || canvas.clientHeight || 1);
    camera.aspect = safeWidth / safeHeight;
    camera.updateProjectionMatrix();
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.setSize(safeWidth, safeHeight, false);
    bounds.x = 10 * camera.aspect;
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
    for (let index = 0; index < count; index += 1) {
      const offset = index * 3;
      const velocityOffset = index * 2;
      let { x: vx, y: vy } = { x: velocities[velocityOffset], y: velocities[velocityOffset + 1] };

      if (pointer.active) {
        const dx = pointer.x * bounds.x - positions[offset];
        const dy = pointer.y * bounds.y - positions[offset + 1];
        const distanceSquared = dx * dx + dy * dy;
        if (distanceSquared < 16 && distanceSquared > 0.01) {
          const force = Math.min(0.003, 0.02 / distanceSquared);
          vx += dx * force;
          vy += dy * force;
        }
      }

      vx *= 0.995;
      vy *= 0.995;
      velocities[velocityOffset] = Math.max(-0.025, Math.min(0.025, vx));
      velocities[velocityOffset + 1] = Math.max(-0.025, Math.min(0.025, vy));
      positions[offset] += velocities[velocityOffset];
      positions[offset + 1] += velocities[velocityOffset + 1];

      if (Math.abs(positions[offset]) > bounds.x) velocities[velocityOffset] *= -1;
      if (Math.abs(positions[offset + 1]) > bounds.y) velocities[velocityOffset + 1] *= -1;
    }
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
    pointer.x = event.clientX / window.innerWidth * 2 - 1;
    pointer.y = -(event.clientY / window.innerHeight * 2 - 1);
    pointer.active = true;
  }

  function setPalette(nextPalette) {
    palette = { ...palette, ...nextPalette };
    particleMaterial.color.set(palette.particle);
    lineMaterial.color.set(palette.line);
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
    window.addEventListener('pointermove', onPointerMove, { passive: true });
    window.addEventListener('resize', resize, { passive: true });
    document.addEventListener('visibilitychange', onVisibilityChange);
    animationFrame = requestAnimationFrame(render);

    return { setPalette, destroy, pause, resume };
  } catch (error) {
    destroy();
    throw error;
  }
}
