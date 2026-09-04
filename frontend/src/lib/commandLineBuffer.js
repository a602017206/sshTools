/** 终端当前输入行缓冲：跟踪可打印字符、退格、回车提交与 Ctrl+C 清空。
 *  不把 Tab(\\t) 记入缓冲——Tab 通常交给远端 shell 做补全，完整命令应以 xterm 可见行为准。 */
export function createCommandLineBuffer() {
  let line = '';

  function push(data) {
    const submitted = [];
    let sawShellTab = false;

    for (let i = 0; i < data.length; i += 1) {
      const ch = data[i];

      if (ch === '\x03') {
        line = '';
        continue;
      }

      if (ch === '\t') {
        // shell 补全：不写入本地缓冲，由调用方稍后从 xterm 可见行同步
        sawShellTab = true;
        continue;
      }

      if (ch === '\x7f' || ch === '\b') {
        if (line.length > 0) {
          line = line.slice(0, -1);
        }
        continue;
      }

      if (ch === '\r') {
        submitted.push(line);
        line = '';
        if (data[i + 1] === '\n') {
          i += 1;
        }
        continue;
      }

      if (ch === '\n') {
        submitted.push(line);
        line = '';
        continue;
      }

      if (ch < ' ') {
        continue;
      }

      line += ch;
    }

    return { submitted, sawShellTab };
  }

  function flushPending() {
    if (line === '') {
      return { submitted: [] };
    }
    const pending = line;
    line = '';
    return { submitted: [pending] };
  }

  function getLine() {
    return line;
  }

  /** 用 xterm 可见命令覆盖本地缓冲（Tab 补全后同步）。 */
  function replaceLine(next) {
    line = String(next || '');
  }

  return { push, flushPending, getLine, replaceLine };
}
