<script>
  import { createEventDispatcher } from 'svelte';
  import { portalToBody } from '../lib/contextMenu.js';
  import {
    FILE_MANAGER_MENU_HEIGHT_BLANK,
    FILE_MANAGER_MENU_HEIGHT_FILE,
    FILE_MANAGER_MENU_WIDTH,
    FILE_MANAGER_SUBMENU_WIDTH,
    getFileManagerMenuFlags,
    getSubmenuPlacement,
    isMacPlatform,
    shiftMenuTopForInlineMore,
  } from '../lib/fileManagerContextMenu.js';

  export let x = 0;
  export let y = 0;
  export let file = null;
  export let currentPath = '/';
  export let history = [];
  export let historyEnabled = true;
  export let clipboard = null;
  export let moreOpen = false;
  export let rootWidth = 0;
  export let rootHeight = 0;

  const dispatch = createEventDispatcher();
  const isMac = typeof navigator !== 'undefined' && isMacPlatform(navigator.userAgent || navigator.platform);
  const meta = isMac ? '⌘' : 'Ctrl+';
  const altShift = isMac ? '⌥⇧' : 'Alt+Shift+';

  $: flags = getFileManagerMenuFlags({
    file,
    currentPath,
    history,
    historyEnabled,
    clipboard,
  });
  $: submenuPlacement = getSubmenuPlacement(x, rootWidth);
  $: menuTop = moreOpen && submenuPlacement === 'down'
    ? shiftMenuTopForInlineMore(y, rootHeight, file ? FILE_MANAGER_MENU_HEIGHT_FILE : FILE_MANAGER_MENU_HEIGHT_BLANK)
    : y;

  function act(id) {
    dispatch('action', id);
  }
</script>

<div
  class="file-manager__menu ops-flyout fixed z-[80] rounded-xl text-xs py-1"
  style={`left: ${x}px; top: ${menuTop}px; width: ${FILE_MANAGER_MENU_WIDTH}px;`}
  use:portalToBody
  on:mousedown|stopPropagation
  on:click|stopPropagation
  on:mouseleave={() => dispatch('more', false)}
>
  <div class="file-manager__menu-toolbar">
    <button type="button" class="file-manager__menu-icon" title="重命名" disabled={!flags.canRename} on:click={() => act('rename')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M4 20h4L18.5 9.5a2.1 2.1 0 00-3-3L5 17v3z" /></svg>
    </button>
    <button type="button" class="file-manager__menu-icon" title="粘贴" disabled={!flags.canPaste} on:click={() => act('paste')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M8 7V5a2 2 0 012-2h4a2 2 0 012 2v2m-8 0h8m-8 0H6a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V9a2 2 0 00-2-2h-2" /></svg>
    </button>
    <button type="button" class="file-manager__menu-icon" title="剪切" disabled={!flags.canCut} on:click={() => act('cut')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="6" cy="18" r="2.5" /><circle cx="18" cy="18" r="2.5" /><path stroke-linecap="round" d="M8 16.5L20 4M16 16.5L4 4" /></svg>
    </button>
    <button type="button" class="file-manager__menu-icon" title="复制" disabled={!flags.canCopy} on:click={() => act('copy')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="9" y="9" width="11" height="11" rx="2" /><path d="M5 15V5a2 2 0 012-2h10" /></svg>
    </button>
  </div>

  <button class="file-manager__menu-item" type="button" disabled={!flags.canOpen} on:click={() => act('openLocal')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M14 4h6v6M10 14L20 4M18 14v6H4V6h6" /></svg>
      {flags.isDir ? '打开' : '本地打开'}
    </span>
    <span class="file-manager__menu-shortcut">Enter</span>
  </button>
  <button class="file-manager__menu-item" type="button" on:click={() => act('refresh')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v6h6M20 20v-6h-6M5 15a7 7 0 0012.9 2M19 9A7 7 0 006.1 7" /></svg>
      刷新
    </span>
    <span class="file-manager__menu-shortcut">{meta}R</span>
  </button>

  <div class="file-manager__menu-sep"></div>

  {#if flags.canFavorite}
    <button class="file-manager__menu-item" type="button" on:click={() => act('toggleFavorite')}>
      <span class="file-manager__menu-label">
        <svg viewBox="0 0 24 24" fill={flags.isFavorite ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M12 18l-6 3 1.5-7L3 9.5l7-.6L12 2l2 6.9 7 .6-4.5 4.5L18 21z" /></svg>
        {flags.isFavorite ? '取消收藏当前路径' : '收藏当前路径'}
      </span>
    </button>
  {/if}
  <button class="file-manager__menu-item" type="button" disabled={!flags.canDownload} on:click={() => act('downloadTo')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v12m0 0l-4-4m4 4l4-4M5 20h14" /></svg>
      下载至
    </span>
  </button>
  <button class="file-manager__menu-item" type="button" on:click={() => act('uploadFiles')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M12 16V6m0 0l-4 4m4-4l4 4M5 20h14" /></svg>
      选择文件上传
    </span>
  </button>
  <button class="file-manager__menu-item" type="button" on:click={() => act('uploadFolder')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M3 7a2 2 0 012-2h3l2 2h9a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" /><path stroke-linecap="round" stroke-linejoin="round" d="M12 16V10m0 0l-2.5 2.5M12 10l2.5 2.5" /></svg>
      选择文件夹上传
    </span>
  </button>

  <div class="file-manager__menu-sep"></div>

  <button class="file-manager__menu-item" type="button" disabled={!flags.canRename} on:click={() => act('rename')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M4 20h4L18.5 9.5a2.1 2.1 0 00-3-3L5 17v3z" /></svg>
      重命名
    </span>
    <span class="file-manager__menu-shortcut">F2</span>
  </button>
  <button class="file-manager__menu-item is-danger" type="button" disabled={!flags.canDelete} on:click={() => act('delete')}>
    <span class="file-manager__menu-label">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M6 7h12M9 7V5a1 1 0 011-1h4a1 1 0 011 1v2m-7 0v12a1 1 0 001 1h6a1 1 0 001-1V7" /></svg>
      删除
    </span>
    <span class="file-manager__menu-shortcut">Backspace</span>
  </button>

  <div class="file-manager__menu-sep"></div>

  <div
    class="file-manager__menu-more"
    on:mouseenter={() => dispatch('more', true)}
  >
    <button class="file-manager__menu-item" class:is-active={moreOpen} type="button" on:click={() => dispatch('more', !moreOpen)}>
      <span class="file-manager__menu-label">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="6" cy="12" r="1.4" fill="currentColor" /><circle cx="12" cy="12" r="1.4" fill="currentColor" /><circle cx="18" cy="12" r="1.4" fill="currentColor" /></svg>
        更多
      </span>
      <span class="file-manager__menu-caret">{submenuPlacement === 'down' ? '⌄' : '›'}</span>
    </button>

    {#if moreOpen}
      <div
        class="file-manager__submenu py-1"
        class:ops-flyout={submenuPlacement !== 'down'}
        class:rounded-xl={submenuPlacement !== 'down'}
        class:is-left={submenuPlacement === 'left'}
        class:is-down={submenuPlacement === 'down'}
        style={submenuPlacement === 'down' ? '' : `width: ${FILE_MANAGER_SUBMENU_WIDTH}px;`}
      >
        <button class="file-manager__menu-item" type="button" on:click={() => act('copyPath')}>
          <span class="file-manager__menu-label">复制路径</span>
          <span class="file-manager__menu-shortcut">{altShift}C</span>
        </button>
        <button class="file-manager__menu-item" type="button" on:click={() => act('newFolder')}>
          <span class="file-manager__menu-label">新建文件夹</span>
        </button>
        <button class="file-manager__menu-item" type="button" on:click={() => act('newFile')}>
          <span class="file-manager__menu-label">新建文件</span>
          <span class="file-manager__menu-shortcut">{altShift}N</span>
        </button>
        <button class="file-manager__menu-item" type="button" disabled={!flags.canChmod} on:click={() => act('chmod')}>
          <span class="file-manager__menu-label">修改权限</span>
          <span class="file-manager__menu-shortcut">{altShift}M</span>
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .file-manager__menu {
    pointer-events: auto;
    color: var(--text-primary);
  }
  .file-manager__menu-toolbar {
    display: flex;
    gap: 4px;
    padding: 6px 8px 8px;
  }
  .file-manager__menu-icon {
    width: 30px;
    height: 28px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: color-mix(in srgb, var(--glass-bg) 70%, transparent);
    color: var(--text-secondary);
    cursor: pointer;
  }
  .file-manager__menu-icon svg {
    width: 14px;
    height: 14px;
  }
  .file-manager__menu-icon:hover:not(:disabled) {
    color: var(--text-primary);
    background: var(--bg-hover);
  }
  .file-manager__menu-icon:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }
  .file-manager__menu-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    min-height: 30px;
    text-align: left;
    padding: 5px 12px;
    color: var(--text-primary);
    background: transparent;
    border: 0;
    cursor: pointer;
  }
  .file-manager__menu-item:hover:not(:disabled),
  .file-manager__menu-item.is-active {
    background: color-mix(in srgb, var(--ops-signal) 16%, transparent);
  }
  .file-manager__menu-item:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .file-manager__menu-item.is-danger {
    color: var(--ops-alert);
  }
  .file-manager__menu-item.is-danger:hover:not(:disabled) {
    background: color-mix(in srgb, var(--ops-alert) 12%, transparent);
  }
  .file-manager__menu-label {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .file-manager__menu-label svg {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
    color: var(--text-secondary);
  }
  .file-manager__menu-shortcut,
  .file-manager__menu-caret {
    color: var(--text-tertiary);
    font-size: 11px;
    flex-shrink: 0;
  }
  .file-manager__menu-caret {
    font-size: 16px;
    line-height: 1;
  }
  .file-manager__menu-sep {
    height: 1px;
    margin: 4px 8px;
    background: var(--border-secondary);
  }
  .file-manager__menu-more {
    position: relative;
  }
  .file-manager__submenu {
    position: absolute;
    top: -6px;
    left: calc(100% - 6px);
    z-index: 2;
  }
  .file-manager__submenu.is-left {
    left: auto;
    right: calc(100% - 6px);
  }
  .file-manager__submenu.is-down {
    position: relative;
    top: 0;
    left: auto;
    right: auto;
    margin: 0 6px 6px;
    border: 1px solid var(--glass-border);
    border-radius: 10px;
    background: color-mix(in srgb, var(--glass-bg) 55%, transparent);
  }
</style>
