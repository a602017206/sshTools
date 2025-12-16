<script>
  import { createEventDispatcher } from 'svelte';

  export let file;
  export let selected = false;

  const dispatch = createEventDispatcher();

  // Get file icon based on type
  function getFileIcon(file) {
    if (file.is_dir) return '📁';
    if (file.is_symlink) return '🔗';

    // File extensions
    const ext = file.name.split('.').pop().toLowerCase();
    const iconMap = {
      'txt': '📄',
      'md': '📝',
      'pdf': '📕',
      'doc': '📘',
      'docx': '📘',
      'xls': '📗',
      'xlsx': '📗',
      'ppt': '📙',
      'pptx': '📙',
      'zip': '📦',
      'tar': '📦',
      'gz': '📦',
      'rar': '📦',
      '7z': '📦',
      'jpg': '🖼️',
      'jpeg': '🖼️',
      'png': '🖼️',
      'gif': '🖼️',
      'svg': '🖼️',
      'mp3': '🎵',
      'wav': '🎵',
      'mp4': '🎬',
      'avi': '🎬',
      'mkv': '🎬',
      'js': '📜',
      'ts': '📜',
      'py': '📜',
      'go': '📜',
      'java': '📜',
      'c': '📜',
      'cpp': '📜',
      'rs': '📜',
      'sh': '⚙️',
      'json': '📋',
      'xml': '📋',
      'yaml': '📋',
      'yml': '📋',
    };

    return iconMap[ext] || '📄';
  }

  // Format file size
  function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
  }

  // Format date
  function formatDate(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now - date;
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
      return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    } else if (days < 7) {
      return `${days}天前`;
    } else {
      return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' });
    }
  }

  function handleClick(event) {
    dispatch('click', { file, event });
  }

  function handleDoubleClick() {
    dispatch('dblclick', file);
  }

  function handleContextMenu(event) {
    event.preventDefault();
    dispatch('contextmenu', { file, event });
  }
</script>

<div
  class="file-item"
  class:selected
  class:directory={file.is_dir}
  on:click={handleClick}
  on:dblclick={handleDoubleClick}
  on:contextmenu={handleContextMenu}
  role="button"
  tabindex="0"
>
  <div class="file-icon">{getFileIcon(file)}</div>
  <div class="file-info">
    <div class="file-name">
      {file.name}
      {#if file.is_symlink && file.link_target}
        <span class="symlink-target">→ {file.link_target}</span>
      {/if}
    </div>
    <div class="file-meta">
      {#if !file.is_dir}
        <span class="file-size">{formatSize(file.size)}</span>
        <span class="separator">•</span>
      {/if}
      <span class="file-date">{formatDate(file.mod_time)}</span>
    </div>
  </div>
</div>

<style>
  .file-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
    border-radius: 4px;
    transition: background-color 0.15s;
    user-select: none;
  }

  .file-item:hover {
    background: var(--bg-hover);
  }

  .file-item.selected {
    background: var(--accent-primary);
    color: white;
  }

  .file-item.selected .file-meta {
    color: rgba(255, 255, 255, 0.8);
  }

  .file-item.directory {
    font-weight: 500;
  }

  .file-icon {
    font-size: 20px;
    margin-right: 12px;
    flex-shrink: 0;
    width: 24px;
    text-align: center;
  }

  .file-info {
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }

  .file-name {
    font-size: 13px;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-item.selected .file-name {
    color: white;
  }

  .symlink-target {
    color: var(--text-secondary);
    font-size: 11px;
    margin-left: 6px;
  }

  .file-item.selected .symlink-target {
    color: rgba(255, 255, 255, 0.7);
  }

  .file-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: 2px;
  }

  .separator {
    opacity: 0.5;
  }
</style>
