<script>
  import { devToolsStore, getRegisteredTools } from '../stores/devtools.js';

  let isDragging = false;
  let startX = 0;
  let startWidth = 0;

  $: tools = getRegisteredTools();
  $: activeTool = tools.find(t => t.id === $devToolsStore.activeTool);

  function handleDragStart(event) {
    isDragging = true;
    startX = event.clientX;
    startWidth = $devToolsStore.width;

    document.addEventListener('mousemove', handleDragMove);
    document.addEventListener('mouseup', handleDragEnd);
    document.body.style.userSelect = 'none';
    document.body.style.cursor = 'col-resize';
  }

  function handleDragMove(event) {
    if (!isDragging) return;
    const delta = startX - event.clientX;
    const newWidth = Math.max(300, Math.min(900, startWidth + delta));
    devToolsStore.setWidth(newWidth);
  }

  function handleDragEnd() {
    if (!isDragging) return;
    isDragging = false;
    document.removeEventListener('mousemove', handleDragMove);
    document.removeEventListener('mouseup', handleDragEnd);
    document.body.style.userSelect = '';
    document.body.style.cursor = '';
  }

  function selectTool(toolId) {
    devToolsStore.setActiveTool(toolId);
  }

  function handleClose() {
    devToolsStore.close();
  }
</script>

{#if $devToolsStore.isOpen}
  <div class="devtools-panel" style="width: {$devToolsStore.width}px">
    <!-- 拖动手柄 -->
    <div
      class="resize-handle"
      class:dragging={isDragging}
      on:mousedown={handleDragStart}
      role="separator"
      aria-label="调整工具面板宽度"
    ></div>

    <!-- 侧边栏：工具列表 -->
    <div class="tools-sidebar">
      <div class="sidebar-header">
        <div class="header-title">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 4.754a3.246 3.246 0 1 0 0 6.492 3.246 3.246 0 0 0 0-6.492zM5.754 8a2.246 2.246 0 1 1 4.492 0 2.246 2.246 0 0 1-4.492 0z"/>
            <path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 0 1-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 0 1-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 0 1 .52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 0 1 1.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 0 1 1.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 0 1 .52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 0 1-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 0 1-1.255-.52l-.094-.319z"/>
          </svg>
          <span class="title">开发工具</span>
        </div>
        <button
          class="close-btn"
          on:click={handleClose}
          title="关闭工具面板 (Esc)"
          aria-label="关闭"
        >
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z"/>
          </svg>
        </button>
      </div>

      <div class="tools-list">
        {#each tools as tool (tool.id)}
          <button
            class="tool-item"
            class:active={$devToolsStore.activeTool === tool.id}
            on:click={() => selectTool(tool.id)}
            title={tool.description}
          >
            <span class="tool-icon">{tool.icon}</span>
            <div class="tool-info">
              <span class="tool-name">{tool.name}</span>
              {#if tool.description}
                <span class="tool-desc">{tool.description}</span>
              {/if}
            </div>
          </button>
        {/each}

        {#if tools.length === 0}
          <div class="empty-message">
            <div class="empty-icon">🔧</div>
            <div class="empty-text">暂无可用工具</div>
            <div class="empty-hint">工具正在加载中...</div>
          </div>
        {/if}
      </div>

      <div class="sidebar-footer">
        <div class="footer-info">
          共 {tools.length} 个工具
        </div>
      </div>
    </div>

    <!-- 主内容区：工具视图 -->
    <div class="tools-content">
      {#if activeTool}
        <svelte:component this={activeTool.component} />
      {:else}
        <div class="welcome-screen">
          <div class="welcome-icon">
            <svg width="64" height="64" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 4.754a3.246 3.246 0 1 0 0 6.492 3.246 3.246 0 0 0 0-6.492zM5.754 8a2.246 2.246 0 1 1 4.492 0 2.246 2.246 0 0 1-4.492 0z"/>
              <path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 0 1-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 0 1-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 0 1 .52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 0 1 1.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 0 1 1.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 0 1 .52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 0 1-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 0 1-1.255-.52l-.094-.319z"/>
            </svg>
          </div>
          <h3>👈 选择一个工具开始使用</h3>
          <p>从左侧列表中选择你需要的开发工具</p>
          <div class="welcome-shortcuts">
            <div class="shortcut">
              <kbd>Esc</kbd>
              <span>关闭面板</span>
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .devtools-panel {
    position: fixed;
    right: 0;
    top: 0;
    height: 100vh;
    background: var(--bg-primary);
    border-left: 1px solid var(--border-primary);
    display: flex;
    z-index: 1000;
    box-shadow: -2px 0 12px rgba(0, 0, 0, 0.08);
    animation: slideIn 0.25s ease-out;
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
    }
    to {
      transform: translateX(0);
    }
  }

  .resize-handle {
    position: absolute;
    left: 0;
    top: 0;
    width: 5px;
    height: 100%;
    background: transparent;
    cursor: col-resize;
    z-index: 10;
    transition: background 0.2s;
  }

  .resize-handle:hover,
  .resize-handle.dragging {
    background: var(--accent-primary);
  }

  .tools-sidebar {
    width: 220px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-primary);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
  }

  .sidebar-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    border-bottom: 1px solid var(--border-primary);
    background: var(--bg-secondary);
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .header-title svg {
    color: var(--accent-primary);
  }

  .title {
    font-weight: 600;
    font-size: 14px;
    color: var(--text-primary);
  }

  .close-btn {
    width: 24px;
    height: 24px;
    padding: 0;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .tools-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }

  .tool-item {
    width: 100%;
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px;
    border: none;
    background: transparent;
    color: var(--text-primary);
    cursor: pointer;
    border-radius: 6px;
    text-align: left;
    transition: all 0.2s;
    margin-bottom: 4px;
  }

  .tool-item:hover {
    background: var(--bg-hover);
    transform: translateX(2px);
  }

  .tool-item.active {
    background: var(--accent-primary);
    color: white;
  }

  .tool-icon {
    font-size: 20px;
    line-height: 1;
    flex-shrink: 0;
  }

  .tool-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .tool-name {
    font-size: 13px;
    font-weight: 600;
    line-height: 1.3;
  }

  .tool-desc {
    font-size: 11px;
    opacity: 0.8;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .tool-item.active .tool-desc {
    opacity: 0.9;
  }

  .empty-message {
    padding: 40px 20px;
    text-align: center;
    color: var(--text-tertiary);
  }

  .empty-icon {
    font-size: 48px;
    margin-bottom: 12px;
    opacity: 0.3;
  }

  .empty-text {
    font-size: 14px;
    margin-bottom: 4px;
  }

  .empty-hint {
    font-size: 12px;
    opacity: 0.7;
  }

  .sidebar-footer {
    padding: 10px 12px;
    border-top: 1px solid var(--border-primary);
    background: var(--bg-secondary);
  }

  .footer-info {
    font-size: 11px;
    color: var(--text-tertiary);
    text-align: center;
  }

  .tools-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: var(--bg-primary);
  }

  .welcome-screen {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    text-align: center;
    color: var(--text-secondary);
  }

  .welcome-icon {
    margin-bottom: 20px;
    opacity: 0.2;
    color: var(--accent-primary);
  }

  .welcome-screen h3 {
    font-size: 18px;
    margin-bottom: 8px;
    color: var(--text-primary);
    font-weight: 600;
  }

  .welcome-screen p {
    font-size: 14px;
    margin-bottom: 24px;
    opacity: 0.8;
  }

  .welcome-shortcuts {
    display: flex;
    gap: 16px;
    justify-content: center;
  }

  .shortcut {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-tertiary);
  }

  kbd {
    padding: 4px 8px;
    border: 1px solid var(--border-primary);
    border-radius: 4px;
    background: var(--bg-secondary);
    color: var(--text-primary);
    font-family: 'SF Mono', 'Monaco', monospace;
    font-size: 11px;
    font-weight: 600;
  }

  /* 滚动条样式 */
  .tools-list::-webkit-scrollbar {
    width: 8px;
  }

  .tools-list::-webkit-scrollbar-track {
    background: var(--bg-secondary);
  }

  .tools-list::-webkit-scrollbar-thumb {
    background: var(--border-primary);
    border-radius: 4px;
  }

  .tools-list::-webkit-scrollbar-thumb:hover {
    background: var(--text-tertiary);
  }
</style>
