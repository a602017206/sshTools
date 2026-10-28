/**
 * Feature showcase carousel helpers.
 */

export function wrapIndex(index, length) {
  if (!length) return 0;
  return ((index % length) + length) % length;
}

export function nextIndex(index, length) {
  return wrapIndex(index + 1, length);
}

export function prevIndex(index, length) {
  return wrapIndex(index - 1, length);
}

export function initCarousel(root, {
  intervalMs = 5200,
  matchMedia = globalThis.matchMedia,
} = {}) {
  if (!root) return { destroy() {}, goTo() {}, next() {}, prev() {} };

  const slides = [...root.querySelectorAll('[data-carousel-slide]')];
  const dots = [...root.querySelectorAll('[data-carousel-dot]')];
  const prevBtn = root.querySelector('[data-carousel-prev]');
  const nextBtn = root.querySelector('[data-carousel-next]');
  let index = Math.max(0, slides.findIndex((slide) => slide.classList.contains('is-active')));
  if (index < 0) index = 0;
  let timer = 0;
  const reduced = (() => {
    try {
      return Boolean(matchMedia?.('(prefers-reduced-motion: reduce)')?.matches);
    } catch {
      return false;
    }
  })();

  function render() {
    slides.forEach((slide, i) => {
      const active = i === index;
      slide.classList.toggle('is-active', active);
      slide.setAttribute('aria-hidden', active ? 'false' : 'true');
    });
    dots.forEach((dot, i) => {
      const active = i === index;
      dot.classList.toggle('is-active', active);
      dot.setAttribute('aria-current', active ? 'true' : 'false');
    });
  }

  function goTo(next) {
    index = wrapIndex(next, slides.length);
    render();
  }

  function next() {
    goTo(nextIndex(index, slides.length));
  }

  function prev() {
    goTo(prevIndex(index, slides.length));
  }

  function stop() {
    if (timer) {
      clearInterval(timer);
      timer = 0;
    }
  }

  function start() {
    stop();
    if (reduced || slides.length < 2) return;
    timer = globalThis.setInterval(next, intervalMs);
  }

  const onPrev = () => { prev(); start(); };
  const onNext = () => { next(); start(); };
  prevBtn?.addEventListener('click', onPrev);
  nextBtn?.addEventListener('click', onNext);
  dots.forEach((dot, i) => {
    dot.addEventListener('click', () => {
      goTo(i);
      start();
    });
  });

  root.addEventListener('pointerenter', stop);
  root.addEventListener('pointerleave', start);

  render();
  start();

  return {
    goTo,
    next,
    prev,
    destroy() {
      stop();
      prevBtn?.removeEventListener('click', onPrev);
      nextBtn?.removeEventListener('click', onNext);
      root.removeEventListener('pointerenter', stop);
      root.removeEventListener('pointerleave', start);
    },
  };
}
