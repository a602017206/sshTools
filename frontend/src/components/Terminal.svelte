<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import '@xterm/xterm/css/xterm.css';
  import { themeStore } from '../stores.js';
  import { get } from 'svelte/store';

  export let sessionId = null;
  export let onData = null;
  export let onResize = null;
  export let onZModemTransfer = null;

  let terminalElement;
  let terminal;
  let fitAddon;
  let Zmodem = null;
  let zsentry = null;
  let zsession = null;
  let skip_zmodem = false;
  // 从 themeStore 获取初始主题值
  let currentTheme = get(themeStore);

  // 深色主题配置
  const darkTheme = {
    background: '#1e1e1e',
    foreground: '#d4d4d4',
    cursor: '#d4d4d4',
    black: '#000000',
    red: '#cd3131',
    green: '#0dbc79',
    yellow: '#e5e510',
    blue: '#2472c8',
    magenta: '#bc3fbc',
    cyan: '#11a8cd',
    white: '#e5e5e5',
    brightBlack: '#666666',
    brightRed: '#f14c4c',
    brightGreen: '#23d18b',
    brightYellow: '#f5f543',
    brightBlue: '#3b8eea',
    brightMagenta: '#d670d6',
    brightCyan: '#29b8db',
    selectionBackground: 'rgba(2, 136, 209, 0.3)',
    selectionForeground: undefined,
    selectionInactiveBackground: 'rgba(2, 136, 209, 0.15)',
  };

  // 浅色主题配置
  const lightTheme = {
    background: '#ffffff',
    foreground: '#333333',
    cursor: '#333333',
    black: '#000000',
    red: '#cd3131',
    green: '#00bc00',
    yellow: '#949800',
    blue: '#0451a5',
    magenta: '#bc05bc',
    cyan: '#0598bc',
    white: '#555555',
    brightBlack: '#666666',
    brightRed: '#cd3131',
    brightGreen: '#14ce14',
    brightYellow: '#b5ba00',
    brightBlue: '#0451a5',
    brightMagenta: '#bc05bc',
    brightCyan: '#0598bc',
    brightWhite: '#a5a5a5',
    selectionBackground: 'rgba(225, 245, 254, 0.8)',
    selectionForeground: undefined,
    selectionInactiveBackground: 'rgba(225, 245, 254, 0.5)',
  };

  // 订阅主题变化
  const unsubscribe = themeStore.subscribe(state => {
    currentTheme = state;
    if (terminal) {
      terminal.options.theme = currentTheme === 'light' ? lightTheme : darkTheme;
    }
  });

  onMount(async () => {
    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: currentTheme === 'light' ? lightTheme : darkTheme,
      allowProposedApi: true,
      scrollback: 1000,
      convertEol: false // 禁用自动换行转换，让后端控制换行
    });

    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.loadAddon(new WebLinksAddon());

    terminal.open(terminalElement);

    fitAddon.fit();

    // 动态导入 zmodem.js
    try {
      const zmodemModule = await import('zmodem.js');
      Zmodem = zmodemModule.default || zmodemModule;

      // 初始化 ZMODEM Sentry
      zsentry = new Zmodem.Sentry({
        to_terminal: (octets) => {
          // ZMODEM 数据不写入终端
          // ZMODEM 会通过后端返回（ssh:output 事件）来显示
          // 直接写入会导致重复
          console.log('ZMODEM to_terminal received (not writing):', octets.length, 'bytes');
        },
        sender: (octets) => {
          // 发送 ZMODEM 数据到 SSH 会话
          if (onData && sessionId) {
            // 将字节数组转换为字符串
            let str = '';
            for (let i = 0; i < octets.length; i++) {
              str += String.fromCharCode(octets[i]);
            }
            onData(sessionId, str);
          }
        },
        on_detect: (detection) => {
          console.log('ZMODEM detected:', detection.type);

          // 确认 ZMODEM 会话
          zsession = detection.confirm();

          if (zsession.type === "receive") {
            // rz: 服务器发送文件到客户端（下载）
            handleZModemReceive(zsession);
          } else {
            // sz: 服务器请求客户端发送文件（上传）
            handleZModemSend(zsession);
          }
        },
        on_retract: () => {
          console.log('ZMODEM retracted');
          zsession = null;
        }
      });

      console.log('ZMODEM initialized successfully');
    } catch (error) {
      console.error('Failed to initialize ZMODEM:', error);
      // 即使 ZMODEM 初始化失败，终端仍然可以正常工作
    }

    terminal.onData((data) => {
      if (onData && sessionId) {
        onData(sessionId, data);
      }
    });

    terminal.onResize(({ cols, rows }) => {
      if (onResize && sessionId) {
        onResize(sessionId, cols, rows);
      }
    });

    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
    });
    resizeObserver.observe(terminalElement);

    return () => {
      resizeObserver.disconnect();
    };
  });

  onDestroy(() => {
    if (unsubscribe) {
      unsubscribe();
    }
    if (terminal) {
      terminal.dispose();
    }
  });

  // 处理文件下载（rz - 服务器发送文件）
  function handleZModemReceive(session) {
    // 开始 ZMODEM 传输
    skip_zmodem = true;

    session.on("offer", (xfer) => {
      const details = xfer.get_details();

      // 询问用户是否下载
      const shouldDownload = confirm(
        `服务器发送文件\n\n文件名: ${details.name}\n大小: ${details.size || '未知'} 字节\n\n是否接收此文件?`
      );

      if (shouldDownload) {
        xfer.accept().then(() => {
          // 文件接收完成，现在保存文件
          const payload = xfer.get_payload();
          saveFileToDisk(payload, details.name);
        });
      } else {
        xfer.skip();
      }
    });

    session.on("session_end", () => {
      console.log('ZMODEM 接收会话结束');
      zsession = null;
      setTimeout(() => {
        skip_zmodem = false;
      }, 100); // 延迟重置，确保所有数据都处理完
    });
  }

  // 保存文件到本地
  function saveFileToDisk(payload, filename) {
    try {
      // 在 Wails 环境中，我们需要调用后端保存文件
      // 暂时使用 Blob URL 下载（在浏览器环境有效）
      const blob = new Blob([payload], { type: 'application/octet-stream' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      console.log('文件已保存:', filename);
    } catch (e) {
      console.error('保存文件失败:', e);
      alert('保存文件失败: ' + e.message);
    }
  }

  // 处理文件上传（sz - 客户端发送文件）
  async function handleZModemSend(session) {
    // 开始 ZMODEM 传输
    skip_zmodem = true;

    // 创建文件选择器
    const input = document.createElement('input');
    input.type = 'file';
    input.multiple = true;

    return new Promise((resolve) => {
      input.onchange = async (e) => {
        const files = Array.from(e.target.files);

        if (files.length === 0) {
          session.close();
          resolve();
          return;
        }

        try {
          // 发送文件
          for (const file of files) {
            const fileDetails = {
              name: file.name,
              size: file.size,
              mtime: new Date(file.lastModified),
            };

            await sendFile(session, file, fileDetails);
          }

          // 关闭会话
          await session.close();

          // 重置 ZMODEM 标志
          setTimeout(() => {
            skip_zmodem = false;
          }, 100);
        } catch (error) {
          console.error('发送文件失败:', error);
          session.close();
          skip_zmodem = false;
        }
        resolve();
      };

      input.click();
    });
  }

  // 发送单个文件
  async function sendFile(session, file, details) {
    return new Promise((resolve, reject) => {
      session.send_offer(details).then((xfer) => {
        if (!xfer) {
          // 文件被服务器拒绝
          resolve();
          return;
        }

        const reader = new FileReader();
        const chunkSize = 8192; // 8KB chunks

        reader.onload = (e) => {
          const buffer = e.target.result;
          let offset = 0;

          // 分块发送文件
          function sendChunk() {
            if (offset < buffer.byteLength) {
              const chunk = new Uint8Array(buffer, offset, Math.min(chunkSize, buffer.byteLength - offset));
              xfer.send(chunk);
              offset += chunkSize;
              // 继续发送下一块
              setTimeout(sendChunk, 0);
            } else {
              // 文件发送完成
              xfer.end().then(resolve).catch(reject);
            }
          }

          sendChunk();
        };

        reader.onerror = () => {
          reject(new Error('读取文件失败'));
        };

        reader.readAsArrayBuffer(file);
      });
    });
  }

  export function write(data) {
    if (!terminal) return;

    // 如果是字节数组（来自后端），直接写入
    if (typeof data !== 'string') {
      terminal.write(data);
      return;
    }

    // 用户输入的文本数据，发送到后端
    // 不要进行任何 ZMODEM 处理，因为这是用户输入，不是 ZMODEM 协议数据
    terminal.write(data); // 本地回显
  }

  export function writeln(data) {
    if (terminal) {
      terminal.writeln(data);
    }
  }

  export function clear() {
    if (terminal) {
      terminal.clear();
    }
  }

  export function focus() {
    if (terminal) {
      terminal.focus();
    }
  }

  export function getSize() {
    if (terminal) {
      return {
        cols: terminal.cols,
        rows: terminal.rows
      };
    }
    return { cols: 80, rows: 24 };
  }
</script>

<div class="terminal-container" bind:this={terminalElement}>
  <!-- xterm 终端将在这里渲染 -->
</div>

<style>
  .terminal-container {
    width: 100%;
    height: 100%;
    background-color: var(--bg-primary);
    transition: background-color 0.2s ease;
  }

  :global(.xterm) {
    height: 100%;
    padding: 10px;
  }

  :global(.xterm-selection-layer .selection-bar) {
    position: absolute;
    top: -1px;
    bottom: -1px;
    left: -1px;
    right: -1px;
    border: 1px dashed #0288D1;
    pointer-events: none;
    z-index: 10;
  }

  :global(.xterm .xterm-selection) {
    border: 1px dashed #0288D1 !important;
    box-sizing: border-box;
  }
</style>
