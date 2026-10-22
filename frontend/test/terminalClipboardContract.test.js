import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

import { createXtermOptions } from '../src/lib/xtermOptions.js';
import { getTerminalShortcutAction } from '../src/lib/terminalShortcuts.js';

// 复制粘贴是受保护区域（见 .cursor/rules/terminal-clipboard.mdc）。
// 这里失败说明修改终端时动到了复制粘贴：应撤回改动，而不是修改本测试。

const source = readFileSync(new URL('../src/components/Terminal.svelte', import.meta.url), 'utf8');

function functionBody(name) {
  const start = source.search(new RegExp(`\\n  (async )?function ${name}\\(`));
  assert.notEqual(start, -1, `Terminal.svelte 缺少 ${name}()`);
  const rest = source.slice(start + 1);
  const end = rest.search(/\n  (async )?function \w+\(|\n  onMount\(/);
  return end === -1 ? rest : rest.slice(0, end);
}

function listenerBlock(eventName) {
  const marker = `terminalElement.addEventListener('${eventName}', (event) => {`;
  const start = source.indexOf(marker);
  assert.notEqual(start, -1, `Terminal.svelte 缺少 ${eventName} 监听`);
  const closeStart = source.indexOf('\n    }', start);
  const closeEnd = source.indexOf('\n', closeStart + 1);
  return { body: source.slice(start, closeStart), closing: source.slice(closeStart + 1, closeEnd) };
}

function keyHandlerBody() {
  const start = source.indexOf('terminal.attachCustomKeyEventHandler(');
  assert.notEqual(start, -1, 'Terminal.svelte 缺少 attachCustomKeyEventHandler');
  return source.slice(start, source.indexOf('\n    });', start));
}

test('受保护区域标记仍在，且成对出现', () => {
  assert.equal(source.split('剪贴板受保护区域 开始').length - 1, 2);
  assert.equal(source.split('剪贴板受保护区域 结束').length - 1, 2);
});

test('复制只走 Wails ClipboardSetText，并延后到事件结束后', () => {
  assert.match(source, /import \{[^}]*ClipboardGetText[^}]*ClipboardSetText[^}]*\} from '..\/..\/wailsjs\/runtime\/runtime\.js'/);
  const body = functionBody('copyToClipboard');
  assert.match(body, /setTimeout\(\(\) => \{\s*ClipboardSetText\(text\)/);
});

test('终端组件不使用会把 {} 日志写成空白的浏览器复制接口', () => {
  assert.doesNotMatch(source, /execCommand\(/);
  assert.doesNotMatch(source, /clipboardData\.setData/);
  assert.doesNotMatch(source, /navigator\.clipboard\.writeText/);
});

test('粘贴先读 Wails 剪贴板，失败才回退浏览器，并经 prepareTerminalPaste 处理', () => {
  const read = functionBody('readClipboardText');
  assert.ok(read.indexOf('ClipboardGetText()') < read.indexOf('navigator.clipboard.readText()'));
  assert.match(functionBody('pasteText'), /prepareTerminalPaste\(text, Boolean\(terminal\?\.modes\?\.bracketedPasteMode\)\)/);
});

test('复制文本来自选区快照兜底，tail -f 刷屏时不为空', () => {
  assert.match(functionBody('resolvedSelectionText'), /resolveTerminalCopyText\([\s\S]*selectionMemory\.text/);
  assert.match(source, /terminal\.onSelectionChange\?\.\(\(\) => \{\s*applySelectionEvent\('change'\);/);
  const mousedown = listenerBlock('mousedown');
  assert.match(mousedown.body, /applySelectionEvent\('pointerdown'\);/);
  assert.equal(mousedown.closing, '    }, true);');
  assert.match(source, /window\.addEventListener\('mouseup', handleSelectionMouseUp\);/);
});

test('copy 事件在捕获阶段拦截，抢在 xterm 写入空选区之前', () => {
  const copy = listenerBlock('copy');
  assert.match(copy.body, /event\.preventDefault\(\);/);
  assert.match(copy.body, /event\.stopPropagation\(\);/);
  assert.match(copy.body, /copyToClipboard\(text\);/);
  assert.equal(copy.closing, '    }, true);');
  assert.match(listenerBlock('paste').body, /pasteText\(text\);/);
});

test('快捷键处理中复制、粘贴均拦截并返回 false', () => {
  const handler = keyHandlerBody();
  assert.match(handler, /getTerminalShortcutAction\(event, terminal\.hasSelection\(\)\)/);
  assert.match(handler, /action === 'copy'\) \{\s*event\.preventDefault\(\);\s*copySelection\(\);\s*return false;/);
  assert.match(handler, /action === 'paste'\) \{\s*event\.preventDefault\(\);\s*pasteFromClipboard\(\);\s*return false;/);
});

test('复制粘贴快捷键语义保持不变', () => {
  const key = (init) => ({ key: '', code: '', ctrlKey: false, metaKey: false, altKey: false, shiftKey: false, ...init });
  assert.equal(getTerminalShortcutAction(key({ key: 'c', code: 'KeyC', metaKey: true }), true), 'copy');
  assert.equal(getTerminalShortcutAction(key({ key: 'c', code: 'KeyC', metaKey: true }), false), 'noop');
  assert.equal(getTerminalShortcutAction(key({ key: 'c', code: 'KeyC', ctrlKey: true }), true), 'copy');
  assert.equal(getTerminalShortcutAction(key({ key: 'c', code: 'KeyC', ctrlKey: true }), false), null);
  assert.equal(getTerminalShortcutAction(key({ key: 'v', code: 'KeyV', metaKey: true }), false), 'paste');
  assert.equal(getTerminalShortcutAction(key({ key: 'v', code: 'KeyV', ctrlKey: true }), false), 'paste');
});

test('xterm 不开启 convertEol，避免 vim 等全屏程序画面被拆散', () => {
  assert.equal(createXtermOptions({}).convertEol, false);
});
