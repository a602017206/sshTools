<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import '@xterm/xterm/css/xterm.css';
  import { ClipboardGetText, ClipboardSetText } from '../../wailsjs/runtime/runtime.js';
  import { getTerminalShortcutAction, shouldScrollToBottomBeforeArrowKey } from '../lib/terminalShortcuts.js';
  import { getXtermTheme, resolveTerminalThemeFromSettings } from '../lib/terminalTheme.js';
  import { decodeTerminalOutput, normalizeTerminalCharset, terminalContextMenuItems, TERMINAL_CHARSET_OPTIONS } from '../lib/terminalCharset.js';
  import { getViewportMenuPosition, portalToBody } from '../lib/contextMenu.js';
  import { createCommandLineBuffer } from '../lib/commandLineBuffer.js';
  import { pickSuggestFill, shouldOfferSuggest } from '../lib/commandSuggest.js';

  export let sessionId = null;
  export let encoding = 'utf-8';
  export let encodingEnabled = false;
  export let onEncodingChange = null;
  export let onData = null;
  export let onResize = null;
  export let onZModemTransfer = null;
  export let connectionId = null;
  export let commandSuggestEnabled = false;
  export let commandSuggestLimit = 8;

  let terminalElement;
  let terminal;
  let fitAddon;
  let Zmodem = null;
  let zsentry = null;
  let zsession = null;
  let skip_zmodem = false;
  let zmodemProgress = null;
  let zmodemActiveOffer = null;
  let zmodemActive = false;

  // ZMODEM 下载状态（响应式）
  let zmodemDownloadOffer = null;
  let zmodemDownloadAction = 'pending'; // 'pending' | 'accepting' | 'skipping' | 'completed'
  let zmodemDownloadError = null;
  let zmodemDownloadSavedPath = null;
  let zmodemTransferModal = null;
  let handleAppearanceUpdated = null;
  let contextMenu = null;
  const commandLineBuffer = createCommandLineBuffer();
  let suggestItems = [];
  let suggestSelectedIndex = 0;
  let suggestVisible = false;
  let suggestDebounceTimer = null;
  let suggestRequestSeq = 0;
  $: currentEncoding = normalizeTerminalCharset(encoding);
  $: suggestOpen = suggestVisible && suggestItems.length > 0;
  $: if ((!commandSuggestEnabled || !connectionId) && (suggestVisible || suggestItems.length > 0)) {
    clearSuggestOverlay();
  }

  function suggestionText(entry) {
    return entry?.command || entry?.Command || '';
  }

  function clearSuggestOverlay() {
    suggestItems = [];
    suggestSelectedIndex = 0;
    suggestVisible = false;
  }

  function recordSubmittedCommands(submitted) {
    if (!connectionId || !commandSuggestEnabled) {
      return;
    }
    const RecordCommand = window.wailsBindings?.RecordCommand;
    if (typeof RecordCommand !== 'function') {
      return;
    }
    for (const cmd of submitted) {
      if (!cmd || !String(cmd).trim()) {
        continue;
      }
      Promise.resolve(RecordCommand(connectionId, cmd)).catch(() => {});
    }
  }

  function scheduleSuggestRefresh() {
    if (suggestDebounceTimer) {
      clearTimeout(suggestDebounceTimer);
      suggestDebounceTimer = null;
    }
    const line = commandLineBuffer.getLine();
    if (!shouldOfferSuggest(line, commandSuggestEnabled) || !connectionId) {
      clearSuggestOverlay();
      return;
    }
    suggestDebounceTimer = setTimeout(() => {
      refreshSuggestions(line);
    }, 150);
  }

  async function refreshSuggestions(prefix) {
    const SuggestCommands = window.wailsBindings?.SuggestCommands;
    if (typeof SuggestCommands !== 'function' || !connectionId) {
      clearSuggestOverlay();
      return;
    }
    if (!shouldOfferSuggest(prefix, commandSuggestEnabled)) {
      clearSuggestOverlay();
      return;
    }
    const seq = ++suggestRequestSeq;
    try {
      const limit = Number(commandSuggestLimit) > 0 ? Number(commandSuggestLimit) : 8;
      const rows = await SuggestCommands(connectionId, prefix, limit);
      if (seq !== suggestRequestSeq) {
        return;
      }
      const items = Array.isArray(rows) ? rows.filter((row) => suggestionText(row)) : [];
      suggestItems = items;
      suggestSelectedIndex = 0;
      suggestVisible = items.length > 0 && shouldOfferSuggest(commandLineBuffer.getLine(), commandSuggestEnabled);
    } catch (error) {
      console.warn('SuggestCommands failed:', error);
      if (seq === suggestRequestSeq) {
        clearSuggestOverlay();
      }
    }
  }

  /** Tab/点击填入：退格清空远端当前行后发送建议文本，不发送 \\r。
   *  依赖远端 echo 回显来同步本地行缓冲；无回显或缓冲不同步时填入可能不准。 */
  function applySuggestionFill(suggestion) {
    const currentLine = commandLineBuffer.getLine();
    const fill = pickSuggestFill(currentLine, suggestion);
    if (fill == null || fill === '') {
      return;
    }
    const payload = '\x7f'.repeat(currentLine.length) + fill;
    commandLineBuffer.push(payload);
    clearSuggestOverlay();
    if (onData && sessionId) {
      onData(sessionId, payload);
    }
  }

  function handleOutgoingInput(data) {
    const { submitted } = commandLineBuffer.push(data);
    if (submitted.length) {
      recordSubmittedCommands(submitted);
      clearSuggestOverlay();
    } else {
      scheduleSuggestRefresh();
    }
    if (onData && sessionId) {
      onData(sessionId, data);
    }
  }

  function handleSuggestKeyEvent(event) {
    if (!suggestOpen) {
      return true;
    }
    if (event.key === 'Tab') {
      event.preventDefault();
      const text = suggestionText(suggestItems[suggestSelectedIndex]);
      if (text) {
        applySuggestionFill(text);
      }
      return false;
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      suggestSelectedIndex = (suggestSelectedIndex - 1 + suggestItems.length) % suggestItems.length;
      return false;
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      suggestSelectedIndex = (suggestSelectedIndex + 1) % suggestItems.length;
      return false;
    }
    if (event.key === 'Escape') {
      event.preventDefault();
      clearSuggestOverlay();
      return false;
    }
    // Enter：不采纳建议，交给终端原义提交
    return true;
  }
  function readTerminalXtermTheme(settings) {
    const resolved = settings
      ? resolveTerminalThemeFromSettings(settings)
      : (typeof document !== 'undefined' && document.documentElement.dataset.terminalTheme) || 'dark';
    return getXtermTheme(resolved);
  }

  function applyTerminalXtermTheme(settings) {
    if (!terminal) {
      return;
    }
    terminal.options.theme = readTerminalXtermTheme(settings);
    if (typeof terminal.refresh === 'function' && terminal.rows) {
      terminal.refresh(0, terminal.rows - 1);
    }
  }

  function readTerminalTypography() {
    if (typeof document === 'undefined') {
      return {
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace'
      };
    }

    const rootStyles = getComputedStyle(document.documentElement);
    const fontSizeValue = Number.parseInt(rootStyles.getPropertyValue('--terminal-font-size'), 10);
    const fontFamilyValue = rootStyles.getPropertyValue('--terminal-font-family').trim();

    return {
      fontSize: Number.isFinite(fontSizeValue) ? fontSizeValue : 14,
      fontFamily: fontFamilyValue || 'Menlo, Monaco, "Courier New", monospace'
    };
  }

  function applyTerminalTypography() {
    if (!terminal) {
      return;
    }
    const typography = readTerminalTypography();
    terminal.options.fontSize = typography.fontSize;
    terminal.options.fontFamily = typography.fontFamily;
    fitAddon?.fit();
  }

  async function copyToClipboard(text) {
    if (!text) {
      return;
    }

    try {
      const ok = await ClipboardSetText(text);
      if (ok !== false) {
        return;
      }
      console.warn('Wails clipboard copy returned false, falling back to browser clipboard');
    } catch (error) {
      console.warn('Wails clipboard copy failed, falling back to browser clipboard:', error);
    }

    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return;
    }

    throw new Error('Clipboard API is unavailable');
  }

  async function readClipboardText() {
    try {
      return await ClipboardGetText();
    } catch (error) {
      console.warn('Wails clipboard paste failed, falling back to browser clipboard:', error);
    }

    if (navigator.clipboard?.readText) {
      return navigator.clipboard.readText();
    }

    throw new Error('Clipboard API is unavailable');
  }

  function copySelection() {
    const text = terminal?.getSelection?.() || '';
    if (!text) {
      return;
    }
    copyToClipboard(text).catch(error => {
      console.error('Failed to copy terminal selection:', error);
    });
  }

  function pasteText(text) {
    if (text && sessionId) {
      handleOutgoingInput(text);
    }
  }

  async function pasteFromClipboard() {
    try {
      const text = await readClipboardText();
      pasteText(text ?? '');
    } catch (error) {
      console.error('Failed to paste from clipboard:', error);
    }
  }

  function writeDecoded(octets) {
    if (!terminal) return;
    const decoded = decodeTerminalOutput(octets, currentEncoding);
    terminal.write(decoded);
  }

  function closeContextMenu() {
    contextMenu = null;
  }

  function openContextMenu(event) {
    event.preventDefault();
    event.stopPropagation();
    const selectedText = terminal?.getSelection?.() || '';
    contextMenu = {
      selectedText,
      ...getViewportMenuPosition({
        clientX: event.clientX,
        clientY: event.clientY,
        menuWidth: 160,
        menuHeight: 80,
        viewWidth: typeof window !== 'undefined' ? window.innerWidth : Number.POSITIVE_INFINITY,
        viewHeight: typeof window !== 'undefined' ? window.innerHeight : Number.POSITIVE_INFINITY,
      })
    };
  }

  function handleEncodingSelect(event) {
    event.stopPropagation();
    const next = normalizeTerminalCharset(event.currentTarget.value);
    onEncodingChange?.(next);
  }

  function handleContextMenuAction(id) {
    const selectedText = contextMenu?.selectedText || '';
    closeContextMenu();
    if (id === 'copy' && selectedText) {
      copyToClipboard(selectedText).catch((error) => {
        console.error('Failed to copy terminal selection:', error);
      });
      return;
    }
    if (id === 'paste') {
      pasteFromClipboard();
    }
  }

  onMount(async () => {
    const typography = readTerminalTypography();

    terminal = new Terminal({
      cursorBlink: true,
      fontSize: typography.fontSize,
      fontFamily: typography.fontFamily,
      theme: readTerminalXtermTheme(),
      allowProposedApi: true,
      scrollback: 1000,
      scrollOnUserInput: true,
      convertEol: true, // 启用自动换行转换，确保 \n 转换为 \r\n，光标回到行首
      rightClickSelectsWord: false,
      macOptionClickForcesSelection: true, // macOS Option+Click 强制选择
    });

    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.loadAddon(new WebLinksAddon());

    terminal.open(terminalElement);

    fitAddon.fit();

    // 常见复制、粘贴快捷键。无选区的 Ctrl+C 保持发送中断信号的终端语义。
    // macOS 原生 Edit 菜单会先处理 Cmd+C；复制路径不要 preventDefault，以便触发 copy 事件。
    terminal.attachCustomKeyEventHandler((event) => {
      if (event.type !== 'keydown') {
        return true;
      }

      if (handleSuggestKeyEvent(event) === false) {
        return false;
      }

      if (shouldScrollToBottomBeforeArrowKey(event, terminal.buffer.active.viewportY)) {
        terminal.scrollToBottom();
      }

      const action = getTerminalShortcutAction(event, terminal.hasSelection());
      if (action === 'copy') {
        copySelection();
        return false;
      }
      if (action === 'noop') {
        return false;
      }
      if (action === 'paste') {
        event.preventDefault();
        pasteFromClipboard();
        return false;
      }
      return true;
    });

    // 承接浏览器 / macOS Edit 菜单的 copy 事件：xterm 选区不在 DOM selection 中。
    terminalElement.addEventListener('copy', (event) => {
      const text = terminal?.getSelection?.() || '';
      if (!text) {
        return;
      }
      event.clipboardData?.setData('text/plain', text);
      event.preventDefault();
      copyToClipboard(text).catch(error => {
        console.error('Failed to copy terminal selection via copy event:', error);
      });
    });

    terminalElement.addEventListener('paste', (event) => {
      const text = event.clipboardData?.getData('text/plain');
      if (text) {
        event.preventDefault();
        pasteText(text);
      }
    });
    terminalElement.addEventListener('contextmenu', openContextMenu);

    // 动态导入 zmodem.js
    try {
      const zmodemModule = await import('zmodem.js');
      Zmodem = zmodemModule.default || zmodemModule;

      // 初始化 ZMODEM Sentry
      zsentry = new Zmodem.Sentry({
        to_terminal: (octets) => {
          // 非 ZMODEM 数据写入终端
          writeDecoded(new Uint8Array(octets));
        },
        sender: (octets) => {
          // 发送 ZMODEM 数据到 SSH 会话
          if (sessionId) {
            if (onZModemTransfer) {
              onZModemTransfer(sessionId, new Uint8Array(octets));
              return;
            }
            if (onData) {
              // 将字节数组转换为字符串
              let str = '';
              for (let i = 0; i < octets.length; i++) {
                str += String.fromCharCode(octets[i]);
              }
              onData(sessionId, str);
            }
          }
        },
        on_detect: (detection) => {
          console.log('ZMODEM detected:', detection.type);
          console.log('Detection object:', detection);

          // 确认 ZMODEM 会话
          zsession = detection.confirm();

          console.log('ZMODEM confirmed, type:', zsession.type, 'has zsession:', !!zsession);

          if (!zsession) {
            console.error('ERROR: zsession is null after confirm()');
            return;
          }

          zmodemActive = true;
          if (terminal) {
            terminal.options.disableStdin = true;
          }

          if (zsession.type === "receive") {
            // rz: 服务器发送文件到客户端（下载）
            console.log('Calling handleZModemReceive for download');
            handleZModemReceive(zsession);
          } else {
            // sz: 服务器请求客户端发送文件（上传）
            console.log('Calling handleZModemSend for upload');
            handleZModemSend(zsession);
          }
        },
        on_retract: () => {
          console.log('ZMODEM retracted');
          zsession = null;
          zmodemActive = false;
          if (terminal) {
            terminal.options.disableStdin = false;
          }
        }
      });

      console.log('ZMODEM initialized successfully');
    } catch (error) {
      console.error('Failed to initialize ZMODEM:', error);
      // 即使 ZMODEM 初始化失败，终端仍然可以正常工作
    }

    terminal.onData((data) => {
      handleOutgoingInput(data);
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

    handleAppearanceUpdated = (event) => {
      applyTerminalTypography();
      applyTerminalXtermTheme(event?.detail);
    };
    window.addEventListener('app:appearance-updated', handleAppearanceUpdated);

    return () => {
      resizeObserver.disconnect();
      if (handleAppearanceUpdated) {
        window.removeEventListener('app:appearance-updated', handleAppearanceUpdated);
      }
    };
  });

  onDestroy(() => {
    if (suggestDebounceTimer) {
      clearTimeout(suggestDebounceTimer);
      suggestDebounceTimer = null;
    }
    closeContextMenu();
    if (terminal) {
      terminal.dispose();
    }
  });

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

  function encodeBinaryString(octets) {
    let binary = '';
    for (let i = 0; i < octets.length; i++) {
      binary += String.fromCharCode(octets[i]);
    }
    return binary;
  }

  async function collectPayloads(payloads) {
    const chunks = [];
    let totalLength = 0;

    for await (const chunk of payloads) {
      chunks.push(chunk);
      totalLength += chunk.length;
    }

    const combined = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      combined.set(chunk, offset);
      offset += chunk.length;
    }

    return combined;
  }

  async function saveZmodemPayloads(payloads, filename) {
    const octets = await collectPayloads(payloads);
    const { SaveBinaryFile } = window.wailsBindings || {};

    if (typeof SaveBinaryFile === 'function') {
      const encoded = btoa(encodeBinaryString(octets));
      return await SaveBinaryFile(filename, encoded);
    }

    saveFileToDisk(octets, filename);
    return null;
  }

  // 显示 ZMODEM 进度条
  function showZmodemProgress(totalFiles) {
    if (zmodemTransferModal) return;

    zmodemTransferModal = document.createElement('div');
    zmodemTransferModal.className = 'zmodem-progress-modal';
    zmodemTransferModal.innerHTML = `
      <div class="zmodem-progress-content">
        <div class="zmodem-progress-header">
          <span>ZMODEM 文件传输</span>
        </div>
        <div class="zmodem-progress-body">
          <div class="zmodem-progress-item">
            <span id="zmodem-file-name">准备传输...</span>
            <div class="zmodem-progress-bar-container">
              <div class="zmodem-progress-bar" id="zmodem-progress-bar"></div>
            </div>
            <span id="zmodem-progress-text">0%</span>
          </div>
          <div class="zmodem-progress-details">
            <span id="zmodem-files-progress">1 / ${totalFiles}</span>
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(zmodemTransferModal);
  }

  // 更新 ZMODEM 进度
  function updateZmodemProgress(fileIndex, totalFiles, fileName, sent, total) {
    if (!zmodemTransferModal) return;

    const progress = Math.min(100, Math.round((sent / total) * 100));

    document.getElementById('zmodem-file-name').textContent = `正在上传: ${fileName}`;
    document.getElementById('zmodem-progress-bar').style.width = `${progress}%`;
    document.getElementById('zmodem-progress-text').textContent = `${progress}%`;
    document.getElementById('zmodem-files-progress').textContent = `${fileIndex + 1} / ${totalFiles}`;
  }

  // 隐藏 ZMODEM 进度条
  function hideZmodemProgress() {
    if (zmodemTransferModal) {
      document.body.removeChild(zmodemTransferModal);
      zmodemTransferModal = null;
    }
  }

  function handleZModemReceive(session) {
    zmodemDownloadOffer = null;
    zmodemActiveOffer = null;
    zmodemDownloadAction = 'pending';
    zmodemDownloadError = null;
    zmodemDownloadSavedPath = null;

    session.on('offer', (offer) => {
      const details = offer.get_details();
      zmodemActiveOffer = offer;
      zmodemDownloadOffer = details;
      zmodemDownloadAction = 'pending';
      zmodemDownloadError = null;
      zmodemDownloadSavedPath = null;
    });

    session.on('session_end', () => {
      zmodemActiveOffer = null;
      zsession = null;
      zmodemActive = false;
      if (terminal) {
        terminal.options.disableStdin = false;
      }

      if (zmodemDownloadAction === 'accepting') {
        zmodemDownloadAction = 'completed';
        setTimeout(() => {
          zmodemDownloadOffer = null;
          zmodemDownloadAction = 'pending';
          zmodemDownloadSavedPath = null;
        }, 1500);
      } else {
        zmodemDownloadOffer = null;
        zmodemDownloadAction = 'pending';
        zmodemDownloadSavedPath = null;
      }
    });

    session.start();
  }

  async function acceptZmodemDownload() {
    if (!zmodemActiveOffer) {
      return;
    }

    zmodemDownloadAction = 'accepting';
    zmodemDownloadError = null;

    try {
      await zmodemActiveOffer.accept();
      const payloads = zmodemActiveOffer.get_payloads();
      const filename = zmodemActiveOffer.get_details().name;
      zmodemDownloadSavedPath = await saveZmodemPayloads(payloads, filename);
      zmodemActiveOffer = null;
      zmodemDownloadAction = 'completed';
    } catch (error) {
      console.error('接收文件失败:', error);
      zmodemDownloadError = error.message || String(error);
      zmodemDownloadAction = 'pending';
    }
  }

  function skipZmodemDownload() {
    if (zmodemActiveOffer) {
      zmodemActiveOffer.skip();
    }
    zmodemActiveOffer = null;
    zmodemDownloadOffer = null;
    zmodemDownloadAction = 'pending';
    zmodemDownloadError = null;
    zmodemDownloadSavedPath = null;
  }

  // 处理文件上传（sz - 客户端发送文件）
  async function handleZModemSend(session) {
    console.log('handleZModemSend called, session:', session);

    // 创建文件选择器
    const input = document.createElement('input');
    input.type = 'file';
    input.multiple = true;

    return new Promise((resolve) => {
      input.onchange = async (e) => {
        const files = Array.from(e.target.files);
        console.log('Files selected:', files.length, files.map(f => f.name));

        if (files.length === 0) {
          console.log('No files selected, closing session');
          session.close();
          skip_zmodem = false;
          zmodemActive = false;
          if (terminal) {
            terminal.options.disableStdin = false;
          }
          resolve();
          return;
        }

        try {
          // 显示进度条
          showZmodemProgress(files.length);

          // 发送所有文件
          console.log('Starting to send files...');
          for (let i = 0; i < files.length; i++) {
            const file = files[i];
            const fileDetails = {
              name: file.name,
              size: file.size,
              mtime: new Date(file.lastModified),
            };

            console.log(`Sending file ${i + 1}/${files.length}:`, file.name, file.size);
            await sendFile(session, file, fileDetails, i, files.length);
          }

          // 所有文件传输完成，关闭会话
          console.log('All files sent, closing session');
          await session.close();

          // 延迟重置 ZMODEM 标志
          setTimeout(() => {
            skip_zmodem = false;
            hideZmodemProgress();
            zmodemActive = false;
            if (terminal) {
              terminal.options.disableStdin = false;
            }
            resolve(); // Promise resolve
          }, 500);
        } catch (error) {
          console.error('发送文件失败:', error);
          console.error('Error details:', error.stack);
          session.close();
          skip_zmodem = false;
          hideZmodemProgress();
          zmodemActive = false;
          if (terminal) {
            terminal.options.disableStdin = false;
          }
          resolve(); // 即使出错也 resolve
        }
      };

      input.click();
    });
  }

  // 格式化文件大小

  // 格式化文件大小
  function formatFileSize(bytes) {
    if (!bytes) return '0 B';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
  }

  // 发送单个文件
  async function sendFile(session, file, details, fileIndex, totalFiles) {
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
          let totalSent = 0;

          // 分块发送文件
          function sendChunk() {
            if (offset < buffer.byteLength) {
              const chunk = new Uint8Array(buffer, offset, Math.min(chunkSize, buffer.byteLength - offset));
              xfer.send(chunk);
              offset += chunkSize;
              totalSent += chunk.byteLength;

              // 更新进度
              updateZmodemProgress(fileIndex, totalFiles, file.name, totalSent, file.size);

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

    // 转换为字节数组
    let octets;
    if (typeof data === 'string') {
      octets = new Uint8Array(data.split('').map(c => c.charCodeAt(0)));
    } else if (data instanceof Uint8Array) {
      octets = data;
    } else {
      octets = new Uint8Array(data);
    }

    // 将数据喂给 ZMODEM Sentry
    // Sentry 会检测 ZMODEM 序列并调用 on_detect
    if (zsentry && !skip_zmodem) {
      try {
        zsentry.consume(octets);
        // to_terminal 回调会处理非 ZMODEM 数据的显示
        // 不要重复写入终端
        return;
      } catch (error) {
        console.warn('ZMODEM consume failed:', error);
        zsession = null;
        zmodemActiveOffer = null;
        zmodemActive = false;
        skip_zmodem = false;
        if (terminal) {
          terminal.options.disableStdin = false;
        }
      }
    }

    // 如果跳过 ZMODEM 或 Sentry 未初始化，直接写入终端
    writeDecoded(octets);
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

<svelte:window on:click={closeContextMenu} />

<div class="terminal-shell">
  <div
    class="terminal-container ops-terminal-canvas"
    bind:this={terminalElement}
    on:click={closeContextMenu}
  >
    <!-- xterm 终端将在这里渲染 -->
  </div>
  {#if suggestOpen}
    <ul class="command-suggest-overlay" role="listbox" aria-label="常用命令建议">
      {#each suggestItems as item, i (suggestionText(item) + ':' + i)}
        <li
          class="command-suggest-item"
          class:selected={i === suggestSelectedIndex}
          role="option"
          aria-selected={i === suggestSelectedIndex}
          on:mousedown|preventDefault={() => applySuggestionFill(suggestionText(item))}
        >
          <span class="command-suggest-cmd">{suggestionText(item)}</span>
          {#if (item.count ?? item.Count) > 0}
            <span class="command-suggest-count">{item.count ?? item.Count}</span>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
  {#if encodingEnabled}
    <label class="terminal-encoding" on:click|stopPropagation on:mousedown|stopPropagation>
      <span>编码</span>
      <select value={currentEncoding} on:change={handleEncodingSelect}>
        {#each TERMINAL_CHARSET_OPTIONS as option}
          <option value={option.id}>{option.label}</option>
        {/each}
      </select>
    </label>
  {/if}
</div>

{#if contextMenu}
  <div
    class="terminal-context-menu"
    style={`left:${contextMenu.x}px; top:${contextMenu.y}px;`}
    use:portalToBody
    on:click|stopPropagation
    role="menu"
  >
    {#each terminalContextMenuItems(Boolean(contextMenu.selectedText)) as item}
      <button
        type="button"
        disabled={item.disabled}
        on:click={() => handleContextMenuAction(item.id)}
      >
        {item.label}
      </button>
    {/each}
  </div>
{/if}

<!-- ZMODEM 下载对话框（非阻塞） -->
{#if zmodemDownloadOffer}
  <div class="zmodem-download-modal">
    <div class="zmodem-download-content">
      <div class="zmodem-download-header">
        <span class="zmodem-download-title">📥 ZMODEM 文件下载</span>
      </div>

      <div class="zmodem-download-body">
        <div class="zmodem-download-info">
          <div class="zmodem-download-file">
            <span class="zmodem-download-icon">📄</span>
            <div class="zmodem-download-details">
              <span class="zmodem-download-name">{zmodemDownloadOffer.name}</span>
              <span class="zmodem-download-size">{formatFileSize(zmodemDownloadOffer.size)} 字节</span>
            </div>
          </div>
        </div>

        <div class="zmodem-download-actions">
          {#if zmodemDownloadError}
            <div class="zmodem-error-message">
              ❌ {zmodemDownloadError}
            </div>
          {/if}

          {#if zmodemDownloadAction === 'pending'}
            <p class="zmodem-download-prompt">服务器正在发送文件，是否接收？</p>
            <div class="zmodem-download-buttons">
              <button
                class="zmodem-download-btn zmodem-btn-reject"
                on:click={skipZmodemDownload}
              >
                拒绝
              </button>
              <button
                class="zmodem-download-btn zmodem-btn-accept"
                on:click={acceptZmodemDownload}
              >
                接收文件
              </button>
            </div>
          {/if}

          {#if zmodemDownloadAction === 'accepting'}
            <div class="zmodem-downloading">
              <span class="zmodem-spinner"></span>
              <span>正在接收文件...</span>
            </div>
          {/if}

          {#if zmodemDownloadAction === 'completed'}
            <div class="zmodem-completed">
              <span class="zmodem-success-icon">✓</span>
              <span>
                文件接收完成，已保存到：{zmodemDownloadSavedPath || zmodemDownloadOffer.name}
              </span>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}


<style>
  .terminal-shell {
    position: relative;
    width: 100%;
    height: 100%;
    min-height: 0;
  }

  .terminal-container {
    width: 100%;
    height: 100%;
    background-color: var(--ops-terminal-bg);
  }

  .terminal-container.ops-terminal-canvas {
    position: relative;
  }

  .terminal-encoding {
    position: absolute;
    top: 8px;
    right: 10px;
    z-index: 6;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 2px 6px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--bg-primary) 82%, transparent);
    border: 1px solid var(--glass-border);
    color: var(--text-secondary);
    font-size: 11px;
  }

  .command-suggest-overlay {
    position: absolute;
    left: 12px;
    bottom: 12px;
    z-index: 8;
    margin: 0;
    padding: 4px 0;
    list-style: none;
    min-width: 200px;
    max-width: min(480px, calc(100% - 24px));
    max-height: 220px;
    overflow-y: auto;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: color-mix(in srgb, var(--ops-terminal-bg, var(--bg-primary)) 92%, transparent);
    color: var(--text-primary);
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.18);
    font-family: var(--terminal-font-family, Menlo, Monaco, "Courier New", monospace);
    font-size: 12px;
  }

  .command-suggest-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 5px 10px;
    cursor: pointer;
    color: var(--text-primary);
  }

  .command-suggest-item:hover,
  .command-suggest-item.selected {
    background: color-mix(in srgb, var(--accent-primary, #0f9f9a) 22%, transparent);
  }

  .command-suggest-cmd {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .command-suggest-count {
    flex-shrink: 0;
    color: var(--text-secondary);
    font-size: 11px;
    opacity: 0.8;
  }

  .terminal-encoding select {
    border: 0;
    background: transparent;
    color: inherit;
    font: inherit;
    padding: 1px 0;
  }

  .terminal-context-menu {
    position: fixed;
    z-index: 140;
    min-width: 148px;
    padding: 6px;
    border-radius: 12px;
    border: 1px solid var(--glass-border);
    background: var(--bg-primary);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.18);
  }

  .terminal-context-menu button {
    display: block;
    width: 100%;
    text-align: left;
    border: 0;
    background: transparent;
    color: inherit;
    padding: 7px 10px;
    border-radius: 8px;
    font-size: 12px;
    cursor: pointer;
  }

  .terminal-context-menu button:hover:not(:disabled) {
    background: var(--bg-secondary);
  }

  .terminal-context-menu button:disabled {
    opacity: 0.45;
    cursor: default;
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

  /* ZMODEM Progress Modal Styles */
  :global(.zmodem-progress-modal) {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
  }

  :global(.zmodem-progress-content) {
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    padding: 24px;
    min-width: 400px;
    max-width: 500px;
  }

  :global(.zmodem-progress-header) {
    font-size: 18px;
    font-weight: 600;
    color: #1a1a1a;
    margin-bottom: 20px;
  }

  :global(.zmodem-progress-body) {
    margin-bottom: 20px;
  }

  :global(.zmodem-progress-item) {
    margin-bottom: 16px;
  }

  :global(#zmodem-file-name) {
    font-size: 14px;
    color: #4a5568;
    margin-bottom: 8px;
    display: block;
  }

  :global(.zmodem-progress-bar-container) {
    background: #f3f4f6;
    border-radius: 4px;
    height: 8px;
    overflow: hidden;
  }

  :global(.zmodem-progress-bar) {
    height: 100%;
    background: linear-gradient(90deg, #6366f1 0%, #4f46e5 100%);
    transition: width 0.3s ease;
    border-radius: 4px;
  }

  :global(#zmodem-progress-text) {
    font-size: 12px;
    color: #6b7280;
    float: right;
  }

  :global(.zmodem-progress-details) {
    display: flex;
    justify-content: flex-end;
    font-size: 13px;
    color: #6b7280;
  }

  /* ZMODEM 下载对话框样式 */
  :global(.zmodem-download-modal) {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
  }

  :global(.zmodem-download-content) {
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    padding: 24px;
    min-width: 400px;
    max-width: 500px;
  }

  :global(.zmodem-download-header) {
    font-size: 18px;
    font-weight: 600;
    color: #1a1a1a;
    margin-bottom: 20px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  :global(.zmodem-download-title) {
    color: #4a5568;
  }

  :global(.zmodem-download-body) {
    margin-bottom: 24px;
  }

  :global(.zmodem-download-info) {
    background: #f3f4f6;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 20px;
  }

  :global(.zmodem-download-file) {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  :global(.zmodem-download-icon) {
    font-size: 32px;
  }

  :global(.zmodem-download-details) {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  :global(.zmodem-download-name) {
    font-weight: 600;
    color: #1a1a1a;
    font-size: 14px;
  }

  :global(.zmodem-download-size) {
    color: #6b7280;
    font-size: 13px;
  }

  :global(.zmodem-download-actions) {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  :global(.zmodem-download-prompt) {
    color: #4a5568;
    font-size: 14px;
    text-align: center;
    margin-bottom: 16px;
  }

  :global(.zmodem-download-buttons) {
    display: flex;
    gap: 12px;
    justify-content: center;
  }

  :global(.zmodem-download-btn) {
    padding: 10px 24px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  :global(.zmodem-download-btn:hover) {
    opacity: 0.9;
    transform: translateY(-1px);
  }

  :global(.zmodem-btn-accept) {
    background: #4f46e5;
    color: white;
  }

  :global(.zmodem-btn-accept:hover) {
    background: #43a047;
  }

  :global(.zmodem-btn-reject) {
    background: #dc2626;
    color: white;
  }

  :global(.zmodem-btn-reject:hover) {
    background: #b91c1c;
  }

  :global(.zmodem-downloading) {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #4a5568;
    font-size: 14px;
  }

  :global(.zmodem-spinner) {
    width: 20px;
    height: 20px;
    border: 2px solid #4f46e5;
    border-top-color: transparent;
    border-radius: 50%;
    animation: zmodem-spin 0.8s linear infinite;
  }

  @keyframes zmodem-spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  :global(.zmodem-completed) {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #22c55e;
    font-size: 14px;
  }

  :global(.zmodem-success-icon) {
    font-size: 24px;
  }

  :global(.zmodem-error-message) {
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 6px;
    padding: 12px 16px;
    color: #dc2626;
    font-size: 13px;
    text-align: center;
  }
</style>
