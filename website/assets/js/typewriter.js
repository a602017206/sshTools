/**
 * Terminal-style typewriter for hero demo lines.
 */

export function buildTypeSequence(lines, { pause = 18, clearPause = 28 } = {}) {
  const frames = [];
  for (const line of lines) {
    let buffer = '';
    for (const ch of line) {
      buffer += ch;
      frames.push(buffer);
    }
    for (let i = 0; i < pause; i += 1) frames.push(buffer);
    while (buffer.length) {
      buffer = buffer.slice(0, -1);
      frames.push(buffer);
    }
    for (let i = 0; i < clearPause; i += 1) frames.push('');
  }
  return frames;
}

export function initTypewriter(el, lines, {
  intervalMs = 42,
  matchMedia = globalThis.matchMedia,
} = {}) {
  if (!el) return { destroy() {} };
  const reduced = (() => {
    try {
      return Boolean(matchMedia?.('(prefers-reduced-motion: reduce)')?.matches);
    } catch {
      return false;
    }
  })();

  if (reduced || !lines?.length) {
    el.textContent = lines?.[0] || '';
    return { destroy() {} };
  }

  const frames = buildTypeSequence(lines);
  let frame = 0;
  let timer = 0;

  function tick() {
    el.textContent = frames[frame];
    frame = (frame + 1) % frames.length;
  }

  tick();
  timer = globalThis.setInterval(tick, intervalMs);

  return {
    destroy() {
      if (timer) clearInterval(timer);
    },
  };
}
