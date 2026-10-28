/**
 * Page UI motion: scroll reveal + card tilt. Respects reduced motion.
 */

export function prefersReducedMotion(matchMedia = globalThis.matchMedia) {
  try {
    return Boolean(matchMedia?.('(prefers-reduced-motion: reduce)')?.matches);
  } catch {
    return false;
  }
}

export function initFx({ root = document, matchMedia = globalThis.matchMedia } = {}) {
  if (prefersReducedMotion(matchMedia)) {
    root.querySelectorAll('.reveal').forEach((el) => el.classList.add('is-visible'));
    return { destroy() {} };
  }

  const revealEls = [...root.querySelectorAll('.reveal')];
  const tiltEls = [...root.querySelectorAll('[data-fx-tilt]')];
  const cleanups = [];

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
      el.style.setProperty('--reveal-delay', `${Math.min(index % 6, 5) * 70}ms`);
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

  // Soft page cursor glow
  let glow = root.querySelector('.cursor-glow');
  if (!glow && root.body) {
    glow = root.createElement?.('div') || document.createElement('div');
    glow.className = 'cursor-glow';
    glow.setAttribute('aria-hidden', 'true');
    (root.body || root).appendChild(glow);
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

  return {
    destroy() {
      cleanups.forEach((fn) => fn());
    },
  };
}
