<script>
  import { createEventDispatcher } from 'svelte';
  import Icon from './Icon.svelte';

  export let file;
  export let selected = false;

  const dispatch = createEventDispatcher();

  // Get file icon name based on type
  function getFileIconName(file) {
    if (file.is_dir) return 'folder';
    if (file.is_symlink) return 'link';

    // File extensions
    const ext = file.name.split('.').pop().toLowerCase();
    const iconMap = {
      'txt': 'file-text',
      'md': 'file-text',
      'pdf': 'file-text',
      'doc': 'file-text',
      'docx': 'file-text',
      'xls': 'file-text',
      'xlsx': 'file-text',
      'ppt': 'file-text',
      'pptx': 'file-text',
      'zip': 'file-archive',
      'tar': 'file-archive',
      'gz': 'file-archive',
      'rar': 'file-archive',
      '7z': 'file-archive',
      'jpg': 'file-image',
      'jpeg': 'file-image',
      'png': 'file-image',
      'gif': 'file-image',
      'svg': 'file-image',
      'mp3': 'file-music',
      'wav': 'file-music',
      'mp4': 'file-video',
      'avi': 'file-video',
      'mkv': 'file-video',
      'js': 'file-code',
      'ts': 'file-code',
      'py': 'file-code',
      'go': 'file-code',
      'java': 'file-code',
      'c': 'file-code',
      'cpp': 'file-code',
      'rs': 'file-code',
      'sh': 'file-code',
      'json': 'file-code',
      'xml': 'file-code',
      'yaml': 'file-code',
      'yml': 'file-code',
    };

    return iconMap[ext] || 'file';
  }
  
  // Get icon color
  function getIconColor(file) {
    if (file.is_dir) return 'var(--accent-primary)';
    if (file.is_symlink) return 'var(--accent-success)';
    return 'var(--text-secondary)';
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
    const diff = now.getTime() - date.getTime();
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
  <div class="col-icon">
    <Icon name={getFileIconName(file)} size={18} color={selected ? 'white' : getIconColor(file)} />
  </div>
  
  <div class="col-name" title={file.name}>
    {file.name}
    {#if file.is_symlink && file.link_target}
      <span class="symlink-target">→ {file.link_target}</span>
    {/if}
  </div>
  
  <div class="col-size">
    {#if !file.is_dir}
      {formatSize(file.size)}
    {:else}
      -
    {/if}
  </div>
  
  <div class="col-date">
    {formatDate(file.mod_time)}
  </div>
</div>

<style>
  .file-item {
    display: grid;
    grid-template-columns: 32px 1fr 100px 140px;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.15s ease;
    user-select: none;
    border-bottom: 1px solid transparent;
  }

  .file-item:hover {
    background: var(--bg-hover);
  }

  .file-item.selected {
    background: var(--accent-primary);
    color: white;
  }

  .file-item.directory .col-name {
    font-weight: 600;
  }

  .col-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .col-name {
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    padding-right: 12px;
  }

  .symlink-target {
    color: var(--text-secondary);
    font-size: 11px;
    margin-left: 6px;
  }
  
  .file-item.selected .symlink-target {
    color: rgba(255, 255, 255, 0.7);
  }

  .col-size {
    font-size: 12px;
    color: var(--text-secondary);
    text-align: right;
    padding-right: 16px;
  }
  
  .col-date {
    font-size: 12px;
    color: var(--text-secondary);
    text-align: right;
  }
  
  .file-item.selected .col-size,
  .file-item.selected .col-date {
    color: rgba(255, 255, 255, 0.8);
  }
</style>
