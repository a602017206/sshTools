/** Docker 容器日志纯函数（便于单测）：事件名、尾部行数、节流拼接、自动滚底判定。 */

export const LOG_TAIL_OPTIONS = [100, 200, 500, 1000];
export const DEFAULT_LOG_TAIL = 200;
/** 前端日志缓冲上限（字符），超出后丢弃最旧内容。 */
export const MAX_LOG_CHARS = 400000;
export const LOG_FLUSH_INTERVAL_MS = 80;
export const AUTO_SCROLL_THRESHOLD_PX = 24;

export function dockerLogsEventName(sessionId, containerId) {
  return `docker:logs:${sessionId}:${containerId}`;
}

export function normalizeLogTail(value) {
  const n = Number(value);
  return LOG_TAIL_OPTIONS.includes(n) ? n : DEFAULT_LOG_TAIL;
}

/** 事件 payload 为 { chunk } 或纯字符串。 */
export function logChunkText(payload) {
  if (typeof payload === 'string') return payload;
  const chunk = payload?.chunk;
  return typeof chunk === 'string' ? chunk : '';
}

/** 追加文本；超过 maxChars 时保留尾部，并尽量从整行边界截断。 */
export function appendLogText(current, chunk, maxChars = MAX_LOG_CHARS) {
  const next = String(current || '') + String(chunk || '');
  if (next.length <= maxChars) return next;
  const cut = next.length - maxChars;
  if (next[cut - 1] === '\n') return next.slice(cut);
  const newline = next.indexOf('\n', cut);
  return newline >= 0 && newline < next.length - 1 ? next.slice(newline + 1) : next.slice(cut);
}

/** 距底部在阈值内视为"贴底"，应继续自动滚动。 */
export function isNearBottom(metrics, threshold = AUTO_SCROLL_THRESHOLD_PX) {
  if (!metrics) return true;
  const { scrollTop = 0, scrollHeight = 0, clientHeight = 0 } = metrics;
  return scrollHeight - scrollTop - clientHeight <= threshold;
}

/**
 * 节流拼接器：push 缓存文本块，interval 内最多 flush 一次（合并为一个字符串）。
 * timers 可注入以便测试。
 */
export function createThrottledAppender(flush, interval = LOG_FLUSH_INTERVAL_MS, timers = {}) {
  const setTimer = timers.setTimeout || setTimeout;
  const clearTimer = timers.clearTimeout || clearTimeout;
  let pending = [];
  let timer = null;

  function run() {
    timer = null;
    if (pending.length === 0) return;
    const text = pending.join('');
    pending = [];
    flush(text);
  }

  return {
    push(text) {
      if (!text) return;
      pending.push(text);
      if (timer === null) timer = setTimer(run, interval);
    },
    /** 立即输出缓存 */
    flush() {
      if (timer !== null) {
        clearTimer(timer);
        timer = null;
      }
      run();
    },
    /** 丢弃缓存（清空 / 销毁时） */
    cancel() {
      if (timer !== null) {
        clearTimer(timer);
        timer = null;
      }
      pending = [];
    }
  };
}

/** 概览 Mounts 等列表字段统一为字符串数组。 */
export function asStringList(value) {
  return Array.isArray(value) ? value.map((item) => String(item)) : [];
}
