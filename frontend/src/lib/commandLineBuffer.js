/** 终端当前输入行缓冲：跟踪可打印字符、退格、回车提交与 Ctrl+C 清空。 */
export function createCommandLineBuffer() {
  let line = '';

  function push(data) {
    const submitted = [];

    for (let i = 0; i < data.length; i += 1) {
      const ch = data[i];

      if (ch === '\x03') {
        line = '';
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

      if (ch < ' ' && ch !== '\t') {
        continue;
      }

      line += ch;
    }

    return { submitted };
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

  return { push, flushPending, getLine };
}
