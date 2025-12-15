<script>
  import { GetConnections, AddConnection, UpdateConnection, RemoveConnection, TestConnection } from '../../wailsjs/go/main/App.js';
  import { onMount } from 'svelte';

  export let onConnect = null;

  let connections = [];
  let showConnectionForm = false;
  let editingConnection = null;
  let testingConnection = false;
  let testResult = '';

  let formData = {
    id: '',
    name: '',
    host: '',
    port: 22,
    user: '',
    password: '',
    savePassword: false,
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
      console.log('Loaded connections:', connections);
    } catch (error) {
      console.error('Failed to load connections:', error);
      connections = [];
    }
  }

  function showNewConnectionForm() {
    editingConnection = null;
    resetForm();
    showConnectionForm = true;
  }

  function showEditConnectionForm(connection) {
    editingConnection = connection;
    formData = {
      id: connection.id,
      name: connection.name,
      host: connection.host,
      port: connection.port,
      user: connection.user,
      password: '',
      savePassword: false,
      auth_type: connection.auth_type || 'password',
      key_path: connection.key_path || '',
      tags: connection.tags || []
    };
    showConnectionForm = true;
  }

  async function handleSaveConnection() {
    if (!formData.name || !formData.host || !formData.user) {
      alert('请填写必填字段（连接名称、主机地址、用户名）');
      return;
    }

    try {
      const connectionData = {
        id: formData.id || `conn_${Date.now()}`,
        name: formData.name,
        host: formData.host,
        port: parseInt(formData.port),
        user: formData.user,
        auth_type: formData.auth_type,
        key_path: formData.key_path,
        tags: formData.tags
      };

      if (editingConnection) {
        await UpdateConnection(connectionData);
      } else {
        await AddConnection(connectionData);
      }

      await loadConnections();
      resetForm();
      showConnectionForm = false;
      editingConnection = null;
    } catch (error) {
      console.error('Failed to save connection:', error);
      alert('保存连接失败: ' + error);
    }
  }

  async function handleRemoveConnection(id) {
    console.log('🔴 handleRemoveConnection called for id:', id);

    if (!window.confirm('确定要删除此连接吗？')) {
      console.log('用户取消了删除操作');
      return;
    }

    try {
      await RemoveConnection(id);
      await loadConnections();
      console.log('连接已删除:', id);
    } catch (error) {
      console.error('Failed to remove connection:', error);
      alert('删除连接失败: ' + error);
    }
  }

  async function handleTestConnection() {
    if (!formData.host || !formData.user) {
      alert('请填写主机地址和用户名');
      return;
    }

    if (!formData.password) {
      alert('请输入密码以测试连接');
      return;
    }

    testingConnection = true;
    testResult = '';

    try {
      await TestConnection(
        formData.host,
        parseInt(formData.port),
        formData.user,
        formData.password
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
    console.log('🔵 handleConnect called:', connection);

    if (!onConnect) {
      console.error('onConnect callback not provided');
      return;
    }

    const password = window.prompt(`连接到 ${connection.name}\n请输入密码：`);
    if (password) {
      onConnect(connection, password);
    } else {
      console.log('用户取消了密码输入');
    }
  }

  function handleEditConnection(connection) {
    console.log('handleEditConnection called:', connection);
    showEditConnectionForm(connection);
  }

  function resetForm() {
    formData = {
      id: '',
      name: '',
      host: '',
      port: 22,
      user: '',
      password: '',
      savePassword: false,
      auth_type: 'password',
      key_path: '',
      tags: []
    };
    testResult = '';
  }

  function cancelForm() {
    resetForm();
    showConnectionForm = false;
    editingConnection = null;
  }

  // 使用window方法暴露全局函数供onclick使用
  if (typeof window !== 'undefined') {
    window.sshToolsConnect = handleConnect;
    window.sshToolsEdit = handleEditConnection;
    window.sshToolsDelete = handleRemoveConnection;
  }
</script>

<div class="manager">
  <div class="header-bar">
    <h2>SSH 连接</h2>
    <!-- 使用原生onclick -->
    <button class="new-btn" onclick="document.getElementById('new-conn-trigger').click()">
      + 新建连接
    </button>
    <button id="new-conn-trigger" style="display:none" on:click={showNewConnectionForm}></button>
  </div>

  {#if showConnectionForm}
    <div class="form-box">
      <h3>{editingConnection ? '编辑连接' : '新建连接'}</h3>

      <div class="field">
        <label>连接名称 *</label>
        <input type="text" bind:value={formData.name} placeholder="例如: 生产服务器" />
      </div>

      <div class="field">
        <label>主机地址 *</label>
        <input type="text" bind:value={formData.host} placeholder="例如: 192.168.1.100" />
      </div>

      <div class="field-row">
        <div class="field">
          <label>端口</label>
          <input type="number" bind:value={formData.port} />
        </div>
        <div class="field">
          <label>用户名 *</label>
          <input type="text" bind:value={formData.user} placeholder="例如: root" />
        </div>
      </div>

      <div class="field">
        <label>密码</label>
        <input type="password" bind:value={formData.password} placeholder="用于测试连接" />
      </div>

      {#if testResult}
        <div class="result {testResult.includes('成功') ? 'success' : 'error'}">
          {testResult}
        </div>
      {/if}

      <div class="actions">
        <button on:click={cancelForm}>取消</button>
        <button on:click={handleTestConnection} disabled={testingConnection}>
          {testingConnection ? '测试中...' : '测试连接'}
        </button>
        <button on:click={handleSaveConnection} class="primary">保存</button>
      </div>
    </div>
  {/if}

  <div class="list">
    {#if connections.length === 0}
      <div class="empty">
        <p>暂无连接</p>
        <p>点击"新建连接"开始添加</p>
      </div>
    {:else}
      {#each connections as connection (connection.id)}
        <div class="item">
          <div class="info">
            <div class="name">{connection.name}</div>
            <div class="details">{connection.user}@{connection.host}:{connection.port}</div>
          </div>
          <div class="item-actions">
            <!-- 使用原生onclick和全局函数 -->
            <button
              class="act-btn connect-btn"
              onclick="window.sshToolsConnect({JSON.stringify(connection).replace(/"/g, '&quot;')})"
            >
              连接
            </button>
            <button
              class="act-btn edit-btn"
              onclick="window.sshToolsEdit({JSON.stringify(connection).replace(/"/g, '&quot;')})"
            >
              编辑
            </button>
            <button
              class="act-btn delete-btn"
              onclick="window.sshToolsDelete('{connection.id}')"
            >
              删除
            </button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .manager {
    height: 100%;
    padding: 20px;
    background: #252526;
    color: #ccc;
    overflow-y: auto;
  }

  .header-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  h2 {
    margin: 0;
    font-size: 18px;
  }

  h3 {
    margin: 0 0 15px 0;
    font-size: 16px;
  }

  .new-btn,
  .act-btn,
  button {
    padding: 8px 16px;
    background: #3c3c3c;
    color: #ccc;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
  }

  .new-btn:hover,
  button:hover {
    background: #505050;
  }

  .new-btn {
    background: #0e639c;
    color: white;
  }

  .new-btn:hover {
    background: #1177bb;
  }

  .form-box {
    background: #1e1e1e;
    padding: 20px;
    border-radius: 6px;
    margin-bottom: 20px;
  }

  .field {
    margin-bottom: 15px;
  }

  .field-row {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 15px;
  }

  label {
    display: block;
    margin-bottom: 5px;
    font-size: 13px;
  }

  input, select {
    width: 100%;
    padding: 8px;
    background: #3c3c3c;
    border: 1px solid #555;
    border-radius: 3px;
    color: #ccc;
    font-size: 13px;
    box-sizing: border-box;
  }

  .result {
    padding: 10px;
    border-radius: 3px;
    margin-bottom: 15px;
    font-size: 13px;
  }

  .result.success {
    background: #1e3a1e;
    color: #4caf50;
  }

  .result.error {
    background: #3a1e1e;
    color: #f44336;
  }

  .actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .primary {
    background: #0e639c !important;
    color: white !important;
  }

  .primary:hover {
    background: #1177bb !important;
  }

  .list {
    margin-top: 20px;
  }

  .empty {
    text-align: center;
    padding: 40px;
    color: #858585;
  }

  .item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 15px;
    background: #1e1e1e;
    border-radius: 6px;
    margin-bottom: 10px;
  }

  .item:hover {
    background: #2a2d2e;
  }

  .info {
    flex: 1;
  }

  .name {
    font-weight: 500;
    margin-bottom: 5px;
  }

  .details {
    font-size: 12px;
    color: #858585;
  }

  .item-actions {
    display: flex;
    gap: 8px;
  }

  .act-btn {
    padding: 6px 12px;
    font-size: 12px;
  }

  .connect-btn {
    background: #0e639c !important;
    color: white !important;
  }

  .connect-btn:hover {
    background: #1177bb !important;
  }

  .edit-btn {
    background: #3c3c3c !important;
  }

  .edit-btn:hover {
    background: #505050 !important;
  }

  .delete-btn {
    background: #3c3c3c !important;
  }

  .delete-btn:hover {
    background: #a03030 !important;
  }
</style>
