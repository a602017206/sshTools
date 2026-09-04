import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import {
  getXtermTheme,
  resolveTerminalTheme,
} from '../src/lib/terminalTheme.js';

test('终端主题可独立于界面：深色、浅色或跟随界面', () => {
  assert.equal(resolveTerminalTheme('dark', 'light'), 'dark');
  assert.equal(resolveTerminalTheme('light', 'dark'), 'light');
  assert.equal(resolveTerminalTheme('follow', 'light'), 'light');
  assert.equal(resolveTerminalTheme('follow', 'dark'), 'dark');
  assert.equal(resolveTerminalTheme('default', 'light'), 'dark');
  assert.equal(resolveTerminalTheme('', 'light'), 'dark');
});

test('浅色终端与深色终端使用不同背景，便于在亮色 UI 下阅读', () => {
  const dark = getXtermTheme('dark');
  const light = getXtermTheme('light');
  assert.notEqual(dark.background, light.background);
  assert.match(dark.background, /^#0/);
  assert.match(light.background, /^#f/);
});

test('设置页把整体外观和终端主题分成两项', async () => {
  const dialog = await readFile(new URL('../src/components/GlobalSettingsDialog.svelte', import.meta.url), 'utf8');
  const terminal = await readFile(new URL('../src/components/Terminal.svelte', import.meta.url), 'utf8');
  const app = await readFile(new URL('../src/App.svelte', import.meta.url), 'utf8');

  assert.match(dialog, /终端主题/);
  assert.match(dialog, /TERMINAL_THEME_PRESETS/);
  assert.match(dialog, /draft\.terminal_theme = option\.id/);
  assert.match(app, /terminal_theme: settings\.terminal_theme/);
  assert.match(terminal, /getXtermTheme/);
  assert.doesNotMatch(terminal, /明暗 UI 下均保持深色可读区/);
});
