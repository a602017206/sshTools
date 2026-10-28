/**
 * Page UI motion: reveal, tilt, themed cursor shapes, trails, ambient fluff.
 * Respects prefers-reduced-motion and coarse pointers.
 */

import {
  applyCursorStyle,
  getCursorStyle,
  readStoredCursor,
} from './cursor.js';

export function prefersReducedMotion(matchMedia = globalThis.matchMedia) {
  try {
    return Boolean(matchMedia?.('(prefers-reduced-motion: reduce)')?.matches);
  } catch {
    return false;
  }
}

export function supportsCustomCursor(matchMedia = globalThis.matchMedia) {
  if (prefersReducedMotion(matchMedia)) return false;
  try {
    return Boolean(matchMedia?.('(pointer: fine)')?.matches);
  } catch {
    return true;
  }
}

/** Smooth-follow lerp for cursor trail points */
export function lerp(a, b, t) {
  return a + (b - a) * t;
}

export function spawnBurst(root, x, y, {
  count = 8,
  className = 'cursor-spark',
  ttlMs = 520,
  glyph = '',
} = {}) {
  if (!root?.appendChild) return [];
  const nodes = [];
  for (let i = 0; i < count; i += 1) {
    const el = (root.ownerDocument || document).createElement('span');
    el.className = className;
    el.setAttribute('aria-hidden', 'true');
    if (glyph) el.textContent = glyph;
    const angle = (Math.PI * 2 * i) / count + Math.random() * 0.45;
    const dist = 22 + Math.random() * 48;
    el.style.left = `${x}px`;
    el.style.top = `${y}px`;
    el.style.setProperty('--sx', `${Math.cos(angle) * dist}px`);
    el.style.setProperty('--sy', `${Math.sin(angle) * dist}px`);
    el.style.setProperty('--rot', `${(Math.random() * 80 - 40).toFixed(1)}deg`);
    root.appendChild(el);
    nodes.push(el);
    globalThis.setTimeout?.(() => el.remove(), ttlMs);
  }
  return nodes;
}

function ensureAmbientFluff(root) {
  const ambient = root.querySelector?.('.ambient');
  if (!ambient) return;
  if (!ambient.querySelector('.ambient__grain')) {
    const grain = document.createElement('div');
    grain.className = 'ambient__grain';
    ambient.appendChild(grain);
  }
  if (!ambient.querySelector('.ambient__scan')) {
    const scan = document.createElement('div');
    scan.className = 'ambient__scan';
    ambient.appendChild(scan);
  }
  if (!ambient.querySelector('.ambient__sparks')) {
    const sparks = document.createElement('div');
    sparks.className = 'ambient__sparks';
    for (let i = 0; i < 12; i += 1) {
      const spark = document.createElement('span');
      spark.className = 'ambient__spark';
      spark.style.setProperty('--spark-x', `${Math.random() * 100}%`);
      spark.style.setProperty('--spark-delay', `${(Math.random() * 8).toFixed(2)}s`);
      spark.style.setProperty('--spark-dur', `${(5 + Math.random() * 7).toFixed(2)}s`);
      sparks.appendChild(spark);
    }
    ambient.appendChild(sparks);
  }
}

function renderCursorMarkup(style) {
  if (style.id === 'ring') {
    return '<span class="site-cursor__ring"></span><span class="site-cursor__dot"></span>';
  }
  if (style.id === 'spark') {
    return '<span class="site-cursor__star">✦</span><span class="site-cursor__aura"></span>';
  }
  if (style.id === 'comet') {
    return '<span class="site-cursor__comet-core"></span><span class="site-cursor__comet-tail"></span>';
  }
  return `<span class="site-cursor__glyph">${style.glyph}</span><span class="site-cursor__aura"></span>`;
}

function trailCountFor(style) {
  if (style.trail === 'comet') return 12;
  if (style.trail === 'spark') return 9;
  if (style.trail === 'glyphs') return 6;
  return 7;
}

function rebuildTrail(trailLayer, style, doc) {
  trailLayer.innerHTML = '';
  trailLayer.dataset.trail = style.trail;
  const count = trailCountFor(style);
  return Array.from({ length: count }, (_, i) => {
    const node = doc.createElement('span');
    if (style.trail === 'glyphs') {
      node.className = 'cursor-trail__glyph';
      node.textContent = style.glyph;
    } else if (style.trail === 'comet') {
      node.className = 'cursor-trail__comet';
    } else if (style.trail === 'spark') {
      node.className = 'cursor-trail__spark';
      node.textContent = '✧';
    } else {
      node.className = 'cursor-trail__dot';
    }
    node.style.setProperty('--trail-i', String(i));
    trailLayer.appendChild(node);
    return { node, x: 0, y: 0 };
  });
}

function initThemedCursor(root, cleanups) {
  const doc = root.ownerDocument || document;
  const html = doc.documentElement;
  html.classList.add('has-custom-cursor');

  let style = applyCursorStyle(readStoredCursor());

  const cursor = doc.createElement('div');
  cursor.className = 'site-cursor';
  cursor.setAttribute('aria-hidden', 'true');
  cursor.dataset.style = style.id;
  cursor.innerHTML = renderCursorMarkup(style);
  (root.body || root).appendChild(cursor);

  const trailLayer = doc.createElement('div');
  trailLayer.className = 'cursor-trail';
  trailLayer.setAttribute('aria-hidden', 'true');
  (root.body || root).appendChild(trailLayer);

  let trail = rebuildTrail(trailLayer, style, doc);

  let targetX = -100;
  let targetY = -100;
  let curX = -100;
  let curY = -100;
  let prevX = -100;
  let prevY = -100;
  let angle = 0;
  let raf = 0;
  let visible = false;
  let lastSparkAt = 0;

  const interactiveSelector = 'a, button, [role="button"], input, textarea, select, summary, label, .theme-swatch, .theme-mode, .cursor-pick';

  const applyStyle = (nextId) => {
    style = getCursorStyle(nextId);
    cursor.dataset.style = style.id;
    cursor.innerHTML = renderCursorMarkup(style);
    trail = rebuildTrail(trailLayer, style, doc);
  };

  const onCursorEvent = (event) => {
    applyStyle(event.detail?.id || readStoredCursor());
  };
  globalThis.addEventListener?.('sshtools:cursor', onCursorEvent);

  const tick = () => {
    curX = lerp(curX, targetX, 0.32);
    curY = lerp(curY, targetY, 0.32);
    const dx = curX - prevX;
    const dy = curY - prevY;
    const speed = Math.hypot(dx, dy);
    if (speed > 0.35) {
      angle = Math.atan2(dy, dx) * (180 / Math.PI);
    }
    prevX = curX;
    prevY = curY;

    const rotate = (style.trail === 'comet' || style.group === 'zodiac' || style.group === 'pet' || style.id === 'rocket')
      ? ` rotate(${angle + (style.id === 'rocket' ? 45 : 0)}deg)`
      : '';
    cursor.style.transform = `translate3d(${curX}px, ${curY}px, 0)${rotate}`;
    cursor.style.setProperty('--cursor-speed', speed.toFixed(2));

    let px = curX;
    let py = curY;
    for (let i = 0; i < trail.length; i += 1) {
      const point = trail[i];
      const ease = style.trail === 'comet' ? 0.22 - i * 0.012 : 0.18 - i * 0.015;
      point.x = lerp(point.x, px, Math.max(0.06, ease));
      point.y = lerp(point.y, py, Math.max(0.06, ease));
      const fadeRot = style.trail === 'comet' ? ` rotate(${angle}deg)` : '';
      point.node.style.transform = `translate3d(${point.x}px, ${point.y}px, 0)${fadeRot}`;
      point.node.style.opacity = String(Math.max(0.12, 0.9 - i * 0.08 - Math.min(speed, 20) * 0.01));
      px = point.x;
      py = point.y;
    }

    // Motion glitter for spark / heart styles
    if (visible && (style.trail === 'spark' || style.id === 'heart') && speed > 4) {
      const now = performance.now?.() || Date.now();
      if (now - lastSparkAt > 48) {
        lastSparkAt = now;
        const flake = doc.createElement('span');
        flake.className = style.id === 'heart' ? 'cursor-flake cursor-flake--heart' : 'cursor-flake';
        flake.textContent = style.id === 'heart' ? '💕' : '✧';
        flake.style.left = `${curX + (Math.random() * 18 - 9)}px`;
        flake.style.top = `${curY + (Math.random() * 18 - 9)}px`;
        (root.body || root).appendChild(flake);
        globalThis.setTimeout?.(() => flake.remove(), 480);
      }
    }

    raf = globalThis.requestAnimationFrame(tick);
  };
  raf = globalThis.requestAnimationFrame(tick);

  const onMove = (event) => {
    targetX = event.clientX;
    targetY = event.clientY;
    if (!visible) {
      visible = true;
      cursor.classList.add('is-on');
      trailLayer.classList.add('is-on');
      curX = targetX;
      curY = targetY;
      prevX = targetX;
      prevY = targetY;
    }
    const overInteractive = Boolean(event.target?.closest?.(interactiveSelector));
    cursor.classList.toggle('is-link', overInteractive);
  };

  const onDown = (event) => {
    cursor.classList.add('is-press');
    const burstClass = style.burst === 'heart'
      ? 'cursor-spark cursor-spark--heart'
      : style.burst === 'emoji'
        ? 'cursor-spark cursor-spark--emoji'
        : 'cursor-spark';
    const glyph = style.burst === 'heart'
      ? '💖'
      : style.burst === 'emoji'
        ? style.glyph
        : '';
    spawnBurst(root.body || root, event.clientX, event.clientY, {
      count: style.burst === 'emoji' ? 8 : 12,
      className: burstClass,
      ttlMs: 680,
      glyph,
    });
    const ripple = doc.createElement('span');
    ripple.className = 'cursor-ripple';
    ripple.style.left = `${event.clientX}px`;
    ripple.style.top = `${event.clientY}px`;
    (root.body || root).appendChild(ripple);
    globalThis.setTimeout?.(() => ripple.remove(), 700);
  };
  const onUp = () => cursor.classList.remove('is-press');
  const onLeave = () => {
    visible = false;
    cursor.classList.remove('is-on', 'is-press', 'is-link');
    trailLayer.classList.remove('is-on');
  };

  window.addEventListener('pointermove', onMove, { passive: true });
  window.addEventListener('pointerdown', onDown);
  window.addEventListener('pointerup', onUp);
  window.addEventListener('pointerleave', onLeave);
  window.addEventListener('blur', onLeave);

  cleanups.push(() => {
    globalThis.cancelAnimationFrame?.(raf);
    globalThis.removeEventListener?.('sshtools:cursor', onCursorEvent);
    window.removeEventListener('pointermove', onMove);
    window.removeEventListener('pointerdown', onDown);
    window.removeEventListener('pointerup', onUp);
    window.removeEventListener('pointerleave', onLeave);
    window.removeEventListener('blur', onLeave);
    html.classList.remove('has-custom-cursor');
    cursor.remove();
    trailLayer.remove();
  });
}

export function initFx({ root = document, matchMedia = globalThis.matchMedia } = {}) {
  if (prefersReducedMotion(matchMedia)) {
    root.querySelectorAll?.('.reveal')?.forEach?.((el) => el.classList.add('is-visible'));
    return { destroy() {} };
  }

  const revealEls = [...(root.querySelectorAll?.('.reveal') || [])];
  const tiltEls = [...(root.querySelectorAll?.('[data-fx-tilt]') || [])];
  const cleanups = [];

  ensureAmbientFluff(root);

  if (typeof IntersectionObserver === 'function' && revealEls.length) {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          entry.target.classList.add('is-visible');
          observer.unobserve(entry.target);
        }
      },
      { threshold: 0.12, rootMargin: '0px 0px -8% 0px' },
    );
    revealEls.forEach((el, index) => {
      el.style?.setProperty?.('--reveal-delay', `${Math.min(index % 6, 5) * 70}ms`);
      observer.observe(el);
    });
    cleanups.push(() => observer.disconnect());
  } else {
    revealEls.forEach((el) => el.classList.add('is-visible'));
  }

  for (const el of tiltEls) {
    const onMove = (event) => {
      const rect = el.getBoundingClientRect();
      const px = (event.clientX - rect.left) / Math.max(1, rect.width);
      const py = (event.clientY - rect.top) / Math.max(1, rect.height);
      const rx = (0.5 - py) * 10;
      const ry = (px - 0.5) * 12;
      el.style.setProperty('--tilt-x', `${rx.toFixed(2)}deg`);
      el.style.setProperty('--tilt-y', `${ry.toFixed(2)}deg`);
      el.style.setProperty('--tilt-px', `${(px * 100).toFixed(1)}%`);
      el.style.setProperty('--tilt-py', `${(py * 100).toFixed(1)}%`);
      el.classList.add('is-tilting');
    };
    const onLeave = () => {
      el.classList.remove('is-tilting');
      el.style.setProperty('--tilt-x', '0deg');
      el.style.setProperty('--tilt-y', '0deg');
    };
    el.addEventListener('pointermove', onMove);
    el.addEventListener('pointerleave', onLeave);
    cleanups.push(() => {
      el.removeEventListener('pointermove', onMove);
      el.removeEventListener('pointerleave', onLeave);
    });
  }

  let glow = root.querySelector?.('.cursor-glow');
  if (!glow && root.body) {
    glow = document.createElement('div');
    glow.className = 'cursor-glow';
    glow.setAttribute('aria-hidden', 'true');
    root.body.appendChild(glow);
  }
  if (glow) {
    const onPointer = (event) => {
      glow.style.transform = `translate3d(${event.clientX}px, ${event.clientY}px, 0)`;
      glow.classList.add('is-on');
    };
    const onLeave = () => glow.classList.remove('is-on');
    window.addEventListener('pointermove', onPointer, { passive: true });
    window.addEventListener('pointerleave', onLeave);
    cleanups.push(() => {
      window.removeEventListener('pointermove', onPointer);
      window.removeEventListener('pointerleave', onLeave);
      glow.remove();
    });
  }

  if (supportsCustomCursor(matchMedia) && root.body) {
    initThemedCursor(root, cleanups);
  }

  return {
    destroy() {
      cleanups.forEach((fn) => fn());
    },
  };
}
