<script>
  export let tabs = [];              // Array of {id, name, displayName, connectionName, userAtHost, isActive}
  export let activeTabId = null;     // Currently active tab ID
  export let onTabChange = null;     // (tabId) => void
  export let onTabClose = null;      // (tabId) => void
  export let onTabRename = null;     // (tabId, newName) => void
  export let onNewTab = null;        // () => void

  let editingTabId = null;           // Tab currently being renamed
  let editingName = '';              // Temporary name while editing
  let editInputRef = null;           // Reference to edit input

  function handleTabClick(tabId) {
    if (editingTabId) return; // Don't switch if editing
    if (onTabChange) {
      onTabChange(tabId);
    }
  }

  function handleTabDoubleClick(tab) {
    // Start editing
    editingTabId = tab.id;
    editingName = tab.name;

    // Focus input after render
    setTimeout(() => {
      if (editInputRef) {
        editInputRef.select();
      }
    }, 10);
  }

  function handleEditKeydown(event, tabId) {
    if (event.key === 'Enter') {
      confirmRename(tabId);
    } else if (event.key === 'Escape') {
      cancelRename();
    }
  }

  function confirmRename(tabId) {
    const trimmedName = editingName.trim();

    // Validate: non-empty, max 50 characters
    if (trimmedName && trimmedName.length <= 50) {
      if (onTabRename) {
        onTabRename(tabId, trimmedName);
      }
    }

    // Exit edit mode
    editingTabId = null;
    editingName = '';
  }

  function cancelRename() {
    editingTabId = null;
    editingName = '';
  }

  function handleCloseClick(event, tabId) {
    event.stopPropagation(); // Prevent tab activation
    if (onTabClose) {
      onTabClose(tabId);
    }
  }

  function handleNewTabClick() {
    if (onNewTab) {
      onNewTab();
    }
  }

  // Auto-scroll to active tab
  $: if (activeTabId) {
    setTimeout(() => {
      const activeTabElement = document.getElementById(`tab-${activeTabId}`);
      if (activeTabElement) {
        activeTabElement.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'nearest' });
      }
    }, 100);
  }
</script>

<div class="tab-bar">
  {#each tabs as tab (tab.id)}
    <div
      id="tab-{tab.id}"
      class="tab"
      class:active={tab.id === activeTabId}
      on:click={() => handleTabClick(tab.id)}
      on:dblclick={() => handleTabDoubleClick(tab)}
      title={tab.userAtHost}
    >
      <!-- Status indicator -->
      <div class="tab-status"></div>

      <!-- Tab title (editable on double-click) -->
      {#if editingTabId === tab.id}
        <input
          type="text"
          class="tab-title-input"
          bind:value={editingName}
          bind:this={editInputRef}
          on:keydown={(e) => handleEditKeydown(e, tab.id)}
          on:blur={() => confirmRename(tab.id)}
          maxlength="50"
        />
      {:else}
        <div class="tab-title">
          {tab.name}
        </div>
      {/if}

      <!-- Close button -->
      <button
        class="tab-close"
        on:click={(e) => handleCloseClick(e, tab.id)}
        title="关闭"
      >
        ×
      </button>
    </div>
  {/each}

  <!-- Add tab button -->
  <button class="add-tab-btn" on:click={handleNewTabClick} title="新建连接">
    +
  </button>
</div>

<style>
  .tab-bar {
    display: flex;
    align-items: center;
    padding: 0 8px;
    gap: 4px;
    height: 100%;
    overflow-x: auto;
    overflow-y: hidden;
  }

  /* Hide scrollbar for cleaner look */
  .tab-bar::-webkit-scrollbar {
    height: 4px;
  }

  .tab-bar::-webkit-scrollbar-track {
    background: transparent;
  }

  .tab-bar::-webkit-scrollbar-thumb {
    background: var(--scrollbar-thumb);
    border-radius: 2px;
  }

  .tab-bar::-webkit-scrollbar-thumb:hover {
    background: var(--scrollbar-thumb-hover);
  }

  .tab {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    background: var(--bg-hover);
    border: 1px solid transparent;
    border-radius: 4px 4px 0 0;
    cursor: pointer;
    user-select: none;
    transition: background 0.15s;
    min-width: 120px;
    max-width: 200px;
    position: relative;
    flex-shrink: 0;
  }

  .tab.active {
    background: var(--bg-primary);
    border-color: var(--border-active);
    border-bottom-color: var(--bg-primary);
  }

  .tab:hover:not(.active) {
    background: var(--bg-hover);
  }

  .tab-status {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--accent-success);
    margin-right: 8px;
    flex-shrink: 0;
  }

  .tab-title {
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    color: var(--text-primary);
    min-width: 0;
  }

  .tab-title-input {
    flex: 1;
    background: var(--bg-input);
    border: 1px solid var(--border-active);
    color: var(--text-primary);
    padding: 2px 4px;
    font-size: 13px;
    outline: none;
    border-radius: 2px;
    min-width: 0;
  }

  .tab-close {
    width: 16px;
    height: 16px;
    margin-left: 8px;
    padding: 0;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 18px;
    line-height: 16px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tab-close:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
  }

  .add-tab-btn {
    padding: 8px 12px;
    background: transparent;
    color: var(--text-secondary);
    border: none;
    cursor: pointer;
    font-size: 18px;
    line-height: 1;
    flex-shrink: 0;
    border-radius: 4px;
    transition: background 0.15s, color 0.15s;
  }

  .add-tab-btn:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }
</style>
