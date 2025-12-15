<script>
  import { GetConnections, AddConnection, RemoveConnection, TestConnection } from '../../wailsjs/go/main/App.js';
  import { onMount } from 'svelte';

  export let onConnect = null;

  let connections = [];
  let showNewConnectionForm = false;
  let testingConnection = false;
  let testResult = '';

  // New connection form data
  let newConnection = {
    id: '',
    name: '',
    host: '',
    port: 22,
    user: '',
    auth_type: 'password',
    key_path: '',
    tags: []
  };

  onMount(async () => {
    await loadConnections();
  });

  async function loadConnections() {
    try {
      connections = await GetConnections();
    } catch (error) {
      console.error('Failed to load connections:', error);
    }
  }

  async function handleAddConnection() {
    if (!newConnection.name || !newConnection.host || !newConnection.user) {
      alert('请填写必填字段');
      return;
    }

    newConnection.id = `conn_${Date.now()}`;

    try {
      await AddConnection(newConnection);
      await loadConnections();
      resetForm();
      showNewConnectionForm = false;
    } catch (error) {
      console.error('Failed to add connection:', error);
      alert('添加连接失败: ' + error);
    }
  }

  async function handleRemoveConnection(id) {
    if (!confirm('确定要删除此连接吗？')) {
      return;
    }

    try {
      await RemoveConnection(id);
      await loadConnections();
    } catch (error) {
      console.error('Failed to remove connection:', error);
      alert('删除连接失败: ' + error);
    }
  }

  async function handleTestConnection() {
    if (!newConnection.host || !newConnection.user) {
      alert('请填写主机和用户名');
      return;
    }

    testingConnection = true;
    testResult = '';

    try {
      const password = prompt('请输入密码（仅用于测试）：');
      if (!password) {
        testingConnection = false;
        return;
      }

      await TestConnection(
        newConnection.host,
        parseInt(newConnection.port),
        newConnection.user,
        password
      );
      testResult = '✓ 连接成功';
    } catch (error) {
      console.error('Connection test failed:', error);
      testResult = '✗ 连接失败: ' + error;
    } finally {
      testingConnection = false;
    }
  }

  function handleConnect(connection) {
    if (onConnect) {
      const password = prompt(`连接到 ${connection.name}\n请输入密码：`);
      if (password) {
        onConnect(connection, password);
      }
    }
  }

  function resetForm() {
    newConnection = {
      id: '',
      name: '',
      host: '',
      port: 22,
      user: '',
      auth_type: 'password',
      key_path: '',
      tags: []
    };
    testResult = '';
  }
</script>

<div class="connection-manager">
  <div class="header">
    <h2>SSH 连接</h2>
    <button class="btn-new" on:click={() => showNewConnectionForm = !showNewConnectionForm}>
      {showNewConnectionForm ? '取消' : '+ 新建连接'}
    </button>
  </div>

  {#if showNewConnectionForm}
    <div class="new-connection-form">
      <h3>新建连接</h3>
      <div class="form-group">
        <label>连接名称 *</label>
        <input type="text" bind:value={newConnection.name} placeholder="例如: 生产服务器" />
      </div>
      <div class="form-group">
        <label>主机地址 *</label>
        <input type="text" bind:value={newConnection.host} placeholder="例如: 192.168.1.100" />
      </div>
      <div class="form-row">
        <div class="form-group">
          <label>端口</label>
          <input type="number" bind:value={newConnection.port} />
        </div>
        <div class="form-group">
          <label>用户名 *</label>
          <input type="text" bind:value={newConnection.user} placeholder="例如: root" />
        </div>
      </div>
      <div class="form-group">
        <label>认证方式</label>
        <select bind:value={newConnection.auth_type}>
          <option value="password">密码</option>
          <option value="key" disabled>SSH密钥（开发中）</option>
        </select>
      </div>
      {#if testResult}
        <div class="test-result" class:success={testResult.includes('成功')} class:error={testResult.includes('失败')}>
          {testResult}
        </div>
      {/if}
      <div class="form-actions">
        <button class="btn-secondary" on:click={handleTestConnection} disabled={testingConnection}>
          {testingConnection ? '测试中...' : '测试连接'}
        </button>
        <button class="btn-primary" on:click={handleAddConnection}>
          保存
        </button>
      </div>
    </div>
  {/if}

  <div class="connections-list">
    {#if connections.length === 0}
      <div class="empty-state">
        <p>暂无连接</p>
        <p class="hint">点击"新建连接"开始添加</p>
      </div>
    {:else}
      {#each connections as connection}
        <div class="connection-item">
          <div class="connection-info">
            <div class="connection-name">{connection.name}</div>
            <div class="connection-details">
              {connection.user}@{connection.host}:{connection.port}
            </div>
          </div>
          <div class="connection-actions">
            <button class="btn-connect" on:click={() => handleConnect(connection)}>
              连接
            </button>
            <button class="btn-delete" on:click={() => handleRemoveConnection(connection.id)}>
              删除
            </button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .connection-manager {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 20px;
    background-color: #252526;
    color: #cccccc;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  h3 {
    margin: 0 0 15px 0;
    font-size: 16px;
    font-weight: 500;
  }

  .btn-new {
    padding: 8px 16px;
    background-color: #0e639c;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
  }

  .btn-new:hover {
    background-color: #1177bb;
  }

  .new-connection-form {
    background-color: #1e1e1e;
    padding: 20px;
    border-radius: 6px;
    margin-bottom: 20px;
  }

  .form-group {
    margin-bottom: 15px;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 15px;
  }

  label {
    display: block;
    margin-bottom: 5px;
    font-size: 13px;
    color: #cccccc;
  }

  input, select {
    width: 100%;
    padding: 8px;
    background-color: #3c3c3c;
    border: 1px solid #3c3c3c;
    border-radius: 3px;
    color: #cccccc;
    font-size: 13px;
    box-sizing: border-box;
  }

  input:focus, select:focus {
    outline: none;
    border-color: #0e639c;
  }

  .test-result {
    padding: 10px;
    border-radius: 3px;
    margin-bottom: 15px;
    font-size: 13px;
  }

  .test-result.success {
    background-color: #1e3a1e;
    color: #4caf50;
  }

  .test-result.error {
    background-color: #3a1e1e;
    color: #f44336;
  }

  .form-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .btn-primary, .btn-secondary {
    padding: 8px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
  }

  .btn-primary {
    background-color: #0e639c;
    color: white;
  }

  .btn-primary:hover {
    background-color: #1177bb;
  }

  .btn-secondary {
    background-color: #3c3c3c;
    color: #cccccc;
  }

  .btn-secondary:hover {
    background-color: #505050;
  }

  .btn-secondary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .connections-list {
    flex: 1;
    overflow-y: auto;
  }

  .empty-state {
    text-align: center;
    padding: 40px;
    color: #858585;
  }

  .empty-state p {
    margin: 5px 0;
  }

  .hint {
    font-size: 12px;
  }

  .connection-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 15px;
    background-color: #1e1e1e;
    border-radius: 6px;
    margin-bottom: 10px;
  }

  .connection-item:hover {
    background-color: #2a2d2e;
  }

  .connection-info {
    flex: 1;
  }

  .connection-name {
    font-weight: 500;
    margin-bottom: 5px;
  }

  .connection-details {
    font-size: 12px;
    color: #858585;
  }

  .connection-actions {
    display: flex;
    gap: 10px;
  }

  .btn-connect, .btn-delete {
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
  }

  .btn-connect {
    background-color: #0e639c;
    color: white;
  }

  .btn-connect:hover {
    background-color: #1177bb;
  }

  .btn-delete {
    background-color: #3c3c3c;
    color: #cccccc;
  }

  .btn-delete:hover {
    background-color: #a03030;
  }
</style>
