<script>
  export let onCreateConnection = () => {};
  export let onFocusSidebar = () => {};
</script>

<div class="db-workspace-empty" role="status">
  <div class="empty-card">
    <p class="kicker">数据库</p>
    <h2>还没有数据库会话</h2>
    <p class="detail">从左侧资源树连接实例，或新建数据库连接。</p>
    <div class="actions">
      <button type="button" class="secondary" on:click={onFocusSidebar}>从资源树连接</button>
      <button type="button" class="primary" on:click={onCreateConnection}>新建数据库连接</button>
    </div>
  </div>
</div>

<style>
  .db-workspace-empty {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 32px;
    text-align: center;
    /* 让背后氛围色块透出来，卡片才能显出毛玻璃 */
    background:
      radial-gradient(ellipse 55% 45% at 50% 42%, color-mix(in srgb, var(--bg-glow-1) 28%, transparent), transparent 70%),
      transparent;
    color: var(--text-primary);
  }

  .empty-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    max-width: 420px;
    padding: 28px 32px;
    border-radius: 20px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg-strong);
    box-shadow: var(--shadow-glass), var(--shadow-md);
    backdrop-filter: blur(var(--glass-blur)) saturate(var(--glass-saturate));
    -webkit-backdrop-filter: blur(var(--glass-blur)) saturate(var(--glass-saturate));
  }

  .kicker {
    margin: 0;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--ops-signal);
    font-family: var(--terminal-font-family);
  }

  h2 {
    margin: 0;
    font-size: 22px;
    letter-spacing: -0.03em;
  }

  .detail {
    margin: 0;
    max-width: 360px;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    justify-content: center;
    margin-top: 12px;
  }

  button {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    font-weight: 600;
    min-height: 40px;
    padding: 0 16px;
    border-radius: 12px;
    border: 1px solid var(--glass-border);
    transition: border-color var(--trans-fast), color var(--trans-fast), filter var(--trans-fast), transform var(--trans-fast), box-shadow var(--trans-fast);
  }

  button:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 2px;
  }

  button:active {
    transform: scale(0.98);
  }

  .secondary {
    background: var(--glass-bg);
    color: var(--text-primary);
    box-shadow: var(--shadow-glass);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

  .secondary:hover {
    border-color: var(--ops-signal);
    color: var(--ops-signal);
  }

  .primary {
    background: var(--ops-signal);
    border-color: var(--ops-signal);
    color: #fff;
    box-shadow: 0 8px 24px color-mix(in srgb, var(--ops-signal) 35%, transparent);
  }

  .primary:hover {
    filter: brightness(1.05);
  }
</style>
