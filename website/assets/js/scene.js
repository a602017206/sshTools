const THREE_URL = 'https://cdn.jsdelivr.net/npm/three@0.170.0/build/three.module.js';
const DEFAULT_PALETTE = {
  bg: '#0b1214',
  accent: '#2dd4bf',
  particle: '#5eead4',
  line: '#2dd4bf66',
};
const MAX_EDGES = 420;

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
  zVelocities,
  baseZVelocities,
}) {
  const hasZ = Boolean(zVelocities && baseZVelocities && bounds.z != null);

  for (let index = 0; index < count; index += 1) {
    const offset = index * 3;
    const velocityOffset = index * 2;
    let vx = velocities[velocityOffset];
    let vy = velocities[velocityOffset + 1];
    let vz = hasZ ? zVelocities[index] : 0;

    if (pointer.active) {
      const dx = pointer.worldX - positions[offset];
      const dy = pointer.worldY - positions[offset + 1];
      const dz = hasZ ? (pointer.worldZ ?? 0) - positions[offset + 2] : 0;
      const distanceSquared = dx * dx + dy * dy + dz * dz;
      if (distanceSquared < 64 && distanceSquared > 0.04) {
        const force = Math.min(0.012, 0.08 / distanceSquared);
        vx += dx * force;
        vy += dy * force;
        if (hasZ) vz += dz * force * 0.55;
      }
    }

    vx += (baseVelocities[velocityOffset] - vx) * 0.01;
    vy += (baseVelocities[velocityOffset + 1] - vy) * 0.01;
    vx *= 0.993;
    vy *= 0.993;
    vx = Math.max(-0.05, Math.min(0.05, vx));
    vy = Math.max(-0.05, Math.min(0.05, vy));

    let x = positions[offset] + vx;
    let y = positions[offset + 1] + vy;

    if (x > bounds.x) {
      x = bounds.x;
      vx = -Math.abs(vx) || -Math.abs(baseVelocities[velocityOffset]) || -0.004;
      baseVelocities[velocityOffset] = -Math.abs(baseVelocities[velocityOffset]) || -0.004;
    } else if (x < -bounds.x) {
      x = -bounds.x;
      vx = Math.abs(vx) || Math.abs(baseVelocities[velocityOffset]) || 0.004;
      baseVelocities[velocityOffset] = Math.abs(baseVelocities[velocityOffset]) || 0.004;
    }

    if (y > bounds.y) {
      y = bounds.y;
      vy = -Math.abs(vy) || -Math.abs(baseVelocities[velocityOffset + 1]) || -0.004;
      baseVelocities[velocityOffset + 1] = -Math.abs(baseVelocities[velocityOffset + 1]) || -0.004;
    } else if (y < -bounds.y) {
      y = -bounds.y;
      vy = Math.abs(vy) || Math.abs(baseVelocities[velocityOffset + 1]) || 0.004;
      baseVelocities[velocityOffset + 1] = Math.abs(baseVelocities[velocityOffset + 1]) || 0.004;
    }

    velocities[velocityOffset] = vx;
    velocities[velocityOffset + 1] = vy;
    positions[offset] = x;
    positions[offset + 1] = y;

    if (hasZ) {
      vz += (baseZVelocities[index] - vz) * 0.01;
      vz *= 0.993;
      vz = Math.max(-0.04, Math.min(0.04, vz));
      let z = positions[offset + 2] + vz;
      if (z > bounds.z) {
        z = bounds.z;
        vz = -Math.abs(vz) || -Math.abs(baseZVelocities[index]) || -0.003;
        baseZVelocities[index] = -Math.abs(baseZVelocities[index]) || -0.003;
      } else if (z < -bounds.z) {
        z = -bounds.z;
        vz = Math.abs(vz) || Math.abs(baseZVelocities[index]) || 0.003;
        baseZVelocities[index] = Math.abs(baseZVelocities[index]) || 0.003;
      }
      zVelocities[index] = vz;
      positions[offset + 2] = z;
    }
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
    ndcX: normalizedX,
    ndcY: normalizedY,
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

function createGlowTexture(THREE) {
  const size = 128;
  const canvas = document.createElement('canvas');
  canvas.width = size;
  canvas.height = size;
  const ctx = canvas.getContext('2d');
  const gradient = ctx.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2);
  gradient.addColorStop(0, 'rgba(255,255,255,1)');
  gradient.addColorStop(0.22, 'rgba(255,255,255,0.9)');
  gradient.addColorStop(0.5, 'rgba(255,255,255,0.28)');
  gradient.addColorStop(1, 'rgba(255,255,255,0)');
  ctx.fillStyle = gradient;
  ctx.fillRect(0, 0, size, size);
  const texture = new THREE.CanvasTexture(canvas);
  texture.needsUpdate = true;
  return texture;
}

function createGridPlane(THREE, size, divisions, color, opacity) {
  const group = new THREE.Group();
  const material = new THREE.LineBasicMaterial({
    color,
    transparent: true,
    opacity,
    depthWrite: false,
  });
  const half = size / 2;
  const step = size / divisions;
  const positions = [];
  for (let i = 0; i <= divisions; i += 1) {
    const edge = -half + i * step;
    positions.push(-half, 0, edge, half, 0, edge);
    positions.push(edge, 0, -half, edge, 0, half);
  }
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
  group.add(new THREE.LineSegments(geometry, material));
  group.userData.material = material;
  group.userData.geometry = geometry;
  return group;
}

/**
 * Full-viewport ops constellation scene.
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
  let clock = 0;

  const full = density !== 'light';
  const count = full ? 280 : 110;
  const starCount = full ? 160 : 60;
  const positions = new Float32Array(count * 3);
  const velocities = new Float32Array(count * 2);
  const baseVelocities = new Float32Array(count * 2);
  const zVelocities = new Float32Array(count);
  const baseZVelocities = new Float32Array(count);
  const starPositions = new Float32Array(starCount * 3);
  const linePositions = new Float32Array(MAX_EDGES * 6);
  const pointer = {
    worldX: 0,
    worldY: 0,
    worldZ: 0,
    ndcX: 0,
    ndcY: 0,
    active: false,
  };
  const cameraTarget = { x: 0, y: 0 };
  const bounds = { x: 14, y: 8, z: 12 };
  const cameraDistance = full ? 18 : 16;

  const scene = new THREE.Scene();
  const fogColor = new THREE.Color(palette.bg);
  scene.fog = new THREE.FogExp2(fogColor, full ? 0.028 : 0.034);
  const camera = new THREE.PerspectiveCamera(58, 1, 0.1, 160);
  const glowTexture = createGlowTexture(THREE);

  const particleGeometry = new THREE.BufferGeometry();
  const starGeometry = new THREE.BufferGeometry();
  const lineGeometry = new THREE.BufferGeometry();
  const initialLineColor = parseCssColor(palette.line);

  const particleMaterial = new THREE.PointsMaterial({
    color: palette.particle,
    map: glowTexture,
    size: full ? 0.55 : 0.42,
    transparent: true,
    opacity: 0.95,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
    sizeAttenuation: true,
  });
  const starMaterial = new THREE.PointsMaterial({
    color: palette.accent,
    map: glowTexture,
    size: full ? 0.22 : 0.16,
    transparent: true,
    opacity: 0.55,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
    sizeAttenuation: true,
  });
  const lineMaterial = new THREE.LineBasicMaterial({
    color: initialLineColor.color,
    transparent: true,
    opacity: Math.min(0.7, initialLineColor.opacity + 0.22),
    depthWrite: false,
    blending: THREE.AdditiveBlending,
  });

  const particles = new THREE.Points(particleGeometry, particleMaterial);
  const stars = new THREE.Points(starGeometry, starMaterial);
  const lines = new THREE.LineSegments(lineGeometry, lineMaterial);
  particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
  starGeometry.setAttribute('position', new THREE.BufferAttribute(starPositions, 3));
  lineGeometry.setAttribute('position', new THREE.BufferAttribute(linePositions, 3));
  lineGeometry.setDrawRange(0, 0);
  scene.add(stars, particles, lines);

  // Multiple constellation hubs across the viewport
  const hubs = [];
  const hubSpecs = full
    ? [
        { x: 0, y: 0.2, z: 0, scale: 1.35 },
        { x: -7.5, y: 3.2, z: -2, scale: 0.85 },
        { x: 8.2, y: -2.4, z: -1.5, scale: 0.95 },
        { x: -5.5, y: -3.8, z: 2, scale: 0.7 },
        { x: 6.5, y: 4.1, z: 1.2, scale: 0.75 },
      ]
    : [
        { x: 0, y: 0, z: 0, scale: 1 },
        { x: -5, y: 2.5, z: -1, scale: 0.7 },
        { x: 5.5, y: -2, z: 1, scale: 0.75 },
      ];

  for (const hub of hubSpecs) {
    const coreGeo = new THREE.OctahedronGeometry(hub.scale * 0.9, 0);
    const coreMat = new THREE.MeshBasicMaterial({
      color: palette.accent,
      wireframe: true,
      transparent: true,
      opacity: 0.38,
    });
    const core = new THREE.Mesh(coreGeo, coreMat);
    core.position.set(hub.x, hub.y, hub.z);

    const glowGeo = new THREE.OctahedronGeometry(hub.scale * 0.35, 0);
    const glowMat = new THREE.MeshBasicMaterial({
      color: palette.particle,
      transparent: true,
      opacity: 0.55,
    });
    const glow = new THREE.Mesh(glowGeo, glowMat);
    glow.position.copy(core.position);

    scene.add(core, glow);
    hubs.push({
      core,
      glow,
      coreGeo,
      coreMat,
      glowGeo,
      glowMat,
      spin: 0.2 + Math.random() * 0.35,
    });
  }

  const ringMats = [];
  const rings = [];
  const ringSpecs = full
    ? [
        { r: 5.5, tube: 0.02, tilt: 0.55, speed: 0.16, y: 0 },
        { r: 8.5, tube: 0.015, tilt: -0.7, speed: -0.1, y: 0.4 },
        { r: 11.5, tube: 0.012, tilt: 1.05, speed: 0.07, y: -0.3 },
        { r: 7.2, tube: 0.01, tilt: 0.2, speed: 0.13, y: 1.2 },
      ]
    : [
        { r: 4.2, tube: 0.014, tilt: 0.45, speed: 0.12, y: 0 },
        { r: 6.8, tube: 0.01, tilt: -0.6, speed: -0.08, y: 0.2 },
      ];

  for (const spec of ringSpecs) {
    const geo = new THREE.TorusGeometry(spec.r, spec.tube, 10, 128);
    const mat = new THREE.MeshBasicMaterial({
      color: palette.accent,
      transparent: true,
      opacity: 0.28,
      depthWrite: false,
    });
    const mesh = new THREE.Mesh(geo, mat);
    mesh.position.y = spec.y;
    mesh.rotation.x = Math.PI / 2 + spec.tilt;
    mesh.userData.speed = spec.speed;
    mesh.userData.geo = geo;
    scene.add(mesh);
    rings.push(mesh);
    ringMats.push(mat);
  }

  const grid = createGridPlane(THREE, full ? 48 : 32, full ? 24 : 16, palette.accent, full ? 0.12 : 0.08);
  grid.position.y = full ? -7.5 : -5.8;
  grid.position.z = -2;
  scene.add(grid);

  camera.position.set(0, 0, cameraDistance);
  camera.lookAt(0, 0, 0);

  const disposables = [
    particleGeometry,
    starGeometry,
    lineGeometry,
    particleMaterial,
    starMaterial,
    lineMaterial,
    glowTexture,
    grid.userData.geometry,
    grid.userData.material,
    ...rings.map((r) => r.userData.geo),
    ...ringMats,
    ...hubs.flatMap((h) => [h.coreGeo, h.coreMat, h.glowGeo, h.glowMat]),
  ];

  function updateFrustumBounds() {
    const distance = cameraDistance;
    const halfH = Math.tan((camera.fov * Math.PI) / 360) * distance;
    bounds.y = halfH * 1.15;
    bounds.x = halfH * camera.aspect * 1.15;
    bounds.z = full ? 14 : 9;
  }

  function scatterParticles() {
    for (let index = 0; index < count; index += 1) {
      const offset = index * 3;
      const velocityOffset = index * 2;
      const angle = Math.random() * Math.PI * 2;
      const speed = 0.006 + Math.random() * 0.012;
      // Bias a portion toward hubs so the field feels structured but still fills the view
      if (Math.random() < 0.35) {
        const hub = hubSpecs[Math.floor(Math.random() * hubSpecs.length)];
        const radius = 1.2 + Math.random() * 3.5;
        const theta = Math.random() * Math.PI * 2;
        positions[offset] = hub.x + Math.cos(theta) * radius;
        positions[offset + 1] = hub.y + Math.sin(theta) * radius * 0.7;
        positions[offset + 2] = hub.z + (Math.random() - 0.5) * 3;
      } else {
        positions[offset] = (Math.random() - 0.5) * bounds.x * 2;
        positions[offset + 1] = (Math.random() - 0.5) * bounds.y * 2;
        positions[offset + 2] = (Math.random() - 0.5) * bounds.z * 2;
      }
      baseVelocities[velocityOffset] = Math.cos(angle) * speed;
      baseVelocities[velocityOffset + 1] = Math.sin(angle) * speed;
      velocities[velocityOffset] = baseVelocities[velocityOffset];
      velocities[velocityOffset + 1] = baseVelocities[velocityOffset + 1];
      baseZVelocities[index] = (Math.random() - 0.5) * 0.014;
      zVelocities[index] = baseZVelocities[index];
    }
    for (let index = 0; index < starCount; index += 1) {
      const offset = index * 3;
      starPositions[offset] = (Math.random() - 0.5) * bounds.x * 2.4;
      starPositions[offset + 1] = (Math.random() - 0.5) * bounds.y * 2.4;
      starPositions[offset + 2] = -4 - Math.random() * (full ? 18 : 12);
    }
    particleGeometry.attributes.position.needsUpdate = true;
    starGeometry.attributes.position.needsUpdate = true;
  }

  function clampParticlesToBounds() {
    for (let index = 0; index < count; index += 1) {
      const offset = index * 3;
      positions[offset] = Math.max(-bounds.x, Math.min(bounds.x, positions[offset]));
      positions[offset + 1] = Math.max(-bounds.y, Math.min(bounds.y, positions[offset + 1]));
      positions[offset + 2] = Math.max(-bounds.z, Math.min(bounds.z, positions[offset + 2]));
    }
    particleGeometry.attributes.position.needsUpdate = true;
  }

  function resize() {
    const { width, height } = canvas.getBoundingClientRect();
    const safeWidth = Math.max(1, width || canvas.clientWidth || window.innerWidth || 1);
    const safeHeight = Math.max(1, height || canvas.clientHeight || window.innerHeight || 1);
    camera.aspect = safeWidth / safeHeight;
    camera.updateProjectionMatrix();
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.setSize(safeWidth, safeHeight, false);
    updateFrustumBounds();
    clampParticlesToBounds();
  }

  function rebuildLines() {
    let edges = 0;
    const maxDistanceSquared = (full ? 4.2 : 3.2) ** 2;

    for (let from = 0; from < count && edges < MAX_EDGES; from += 1) {
      const fromOffset = from * 3;
      for (let to = from + 1; to < count && edges < MAX_EDGES; to += 1) {
        const toOffset = to * 3;
        const dx = positions[fromOffset] - positions[toOffset];
        const dy = positions[fromOffset + 1] - positions[toOffset + 1];
        const dz = positions[fromOffset + 2] - positions[toOffset + 2];
        if (dx * dx + dy * dy + dz * dz > maxDistanceSquared) continue;

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
      zVelocities,
      baseZVelocities,
    });
    particleGeometry.attributes.position.needsUpdate = true;
  }

  function updateAtmosphere(dt) {
    clock += dt;
    for (const hub of hubs) {
      hub.core.rotation.y += dt * hub.spin;
      hub.core.rotation.x += dt * hub.spin * 0.35;
      hub.glow.rotation.y -= dt * hub.spin * 1.4;
      hub.glowMat.opacity = 0.4 + Math.sin(clock * 2.4 + hub.spin * 8) * 0.22;
      hub.coreMat.opacity = 0.28 + Math.sin(clock * 1.5 + hub.spin * 5) * 0.1;
    }

    lineMaterial.opacity = Math.min(
      0.75,
      (parseCssColor(palette.line).opacity || 0.4) + 0.18 + Math.sin(clock * 1.6) * 0.1,
    );
    starMaterial.opacity = 0.4 + Math.sin(clock * 0.9) * 0.12;

    for (const ring of rings) {
      ring.rotation.z += dt * ring.userData.speed;
      ring.rotation.y += dt * ring.userData.speed * 0.4;
    }

    grid.rotation.y = Math.sin(clock * 0.12) * 0.12;
    stars.rotation.y = clock * 0.015;

    const targetX = pointer.active ? pointer.ndcX * 2.4 : 0;
    const targetY = pointer.active ? pointer.ndcY * 1.4 : 0;
    cameraTarget.x += (targetX - cameraTarget.x) * Math.min(1, dt * 2.2);
    cameraTarget.y += (targetY - cameraTarget.y) * Math.min(1, dt * 2.2);
    camera.position.x = cameraTarget.x;
    camera.position.y = cameraTarget.y;
    camera.position.z = cameraDistance;
    camera.lookAt(cameraTarget.x * 0.2, cameraTarget.y * 0.15, 0);
  }

  let lastTs = performance.now();

  function render(now = performance.now()) {
    if (destroyed || paused) return;
    const dt = Math.min(0.05, (now - lastTs) / 1000);
    lastTs = now;
    updateParticles();
    rebuildLines();
    updateAtmosphere(dt);
    renderer.render(scene, camera);
    animationFrame = requestAnimationFrame(render);
  }

  function onPointerMove(event) {
    const mapped = mapPointerToWorld({
      clientX: event.clientX,
      clientY: event.clientY,
      viewport: canvas.getBoundingClientRect(),
      fov: camera.fov,
      cameraZ: cameraDistance,
    });
    pointer.worldX = mapped.x;
    pointer.worldY = mapped.y;
    pointer.worldZ = 0;
    pointer.ndcX = mapped.ndcX;
    pointer.ndcY = mapped.ndcY;
    pointer.active = true;
  }

  function setPalette(nextPalette) {
    palette = { ...palette, ...nextPalette };
    const lineColor = parseCssColor(palette.line);
    particleMaterial.color.set(palette.particle);
    starMaterial.color.set(palette.accent);
    lineMaterial.color.set(lineColor.color);
    lineMaterial.opacity = Math.min(0.75, lineColor.opacity + 0.2);
    for (const hub of hubs) {
      hub.coreMat.color.set(palette.accent);
      hub.glowMat.color.set(palette.particle);
    }
    for (const mat of ringMats) mat.color.set(palette.accent);
    if (grid.userData.material) grid.userData.material.color.set(palette.accent);
    fogColor.set(palette.bg);
    scene.fog.color.copy(fogColor);
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
    lastTs = performance.now();
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
    for (const item of disposables) item?.dispose?.();
    renderer?.dispose();
  }

  try {
    renderer = new THREE.WebGLRenderer({ canvas, alpha: true, antialias: true });
    renderer.setClearColor(palette.bg, 0);
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
