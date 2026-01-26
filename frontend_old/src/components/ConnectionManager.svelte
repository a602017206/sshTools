<script>
  import { GetConnections, AddConnection, UpdateConnection, RemoveConnection, TestConnection, SelectSSHKeyFile, SavePassword, GetPassword, HasPassword } from '../../wailsjs/go/main/App.js';
  import { onMount } from 'svelte';
  import { showAlert, showError, showConfirm } from '../utils/dialog.js';
  import PasswordPrompt from './PasswordPrompt.svelte';

  export let onConnect = null;

  let connections = [];
  let showConnectionForm = false;
  let editingConnection = null;
  let testingConnection = false;
  let testResult = '';

  // Password prompt modal
  let showPasswordPrompt = false;
  let passwordPromptTitle = '';
  let passwordPromptMessage = '';
  let passwordPromptIsPassword = true;
  let passwordPromptShowSave = false;
  let pendingConnection = null;

  // Connection form data
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
    passphrase: '',
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
      passphrase: '',
      tags: connection.tags || []
    };
    showConnectionForm = true;
  }

  async function handleSaveConnection() {
    if (!formData.name || !formData.host || !formData.user) {
      await showAlert('请填写必填字段（连接名称、主机地址、用户名）');
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
        // Update existing connection
        await UpdateConnection(connectionData);
      } else {
        // Add new connection
        await AddConnection(connectionData);
      }

      // TODO: Save password to credential store if savePassword is true
      if (formData.savePassword && formData.password) {
        console.log('Saving password for connection:', connectionData.id);
        // Will implement credential storage later
      }

      await loadConnections();
      resetForm();
      showConnectionForm = false;
      editingConnection = null;
    } catch (error) {
      console.error('Failed to save connection:', error);
      await showError('保存连接失败: ' + error);
    }
  }

  async function handleRemoveConnection(id) {
    console.log('🔴 handleRemoveConnection called for id:', id);

    const confirmed = await showConfirm('确定要删除此连接吗？');
    if (!confirmed) {
      console.log('用户取消了删除操作');
      return;
    }

    try {
      await RemoveConnection(id);
      await loadConnections();
      console.log('连接已删除:', id);
    } catch (error) {
      console.error('Failed to remove connection:', error);
      await showError('删除连接失败: ' + error);
    }
  }

  async function handleTestConnection() {
    if (!formData.host || !formData.user) {
      await showAlert('请填写主机地址和用户名');
      return;
    }

    // Validate based on auth type
    if (formData.auth_type === 'password') {
      if (!formData.password) {
        await showAlert('请输入密码以测试连接');
        return;
      }
    } else if (formData.auth_type === 'key') {
      if (!formData.key_path) {
        await showAlert('请选择 SSH 密钥文件');
        return;
      }
    }

    testingConnection = true;
    testResult = '';

    try {
      const authValue = formData.auth_type === 'key' ? formData.key_path : formData.password;
      await TestConnection(
        formData.host,
        parseInt(formData.port),
        formData.user,
        formData.auth_type,
        authValue,
        formData.passphrase || ''
      );
      testResult = '✓ 连接成功';
    } catch (error) {
      console.error('Connection test failed:', error);
      testResult = '✗ 连接失败: ' + error;
    } finally {
      testingConnection = false;
    }
  }

  async function handleConnect(connection) {
    console.log('🔵 handleConnect called:', connection);

    if (!onConnect) {
      console.error('onConnect callback not provided');
      await showError('错误：onConnect 回调未提供');
      return;
    }

    if (connection.auth_type === 'key') {
      // For key auth, use saved key path and prompt for passphrase
      pendingConnection = connection;
      passwordPromptTitle = '密钥 Passphrase';
      passwordPromptMessage = `连接到 ${connection.name}\n如果密钥已加密，请输入 Passphrase（否则留空）：`;
      passwordPromptIsPassword = true;
      passwordPromptShowSave = false;
      showPasswordPrompt = true;
    } else {
      // For password auth, try to get saved password first
      let password = null;
      try {
        const hasSaved = await HasPassword(connection.id);
        if (hasSaved) {
          password = await GetPassword(connection.id);
          console.log('Using saved password');
          onConnect(connection, password, '');
          return;
        }
      } catch (error) {
        console.error('Failed to get saved password:', error);
      }

      // No saved password, prompt user
      pendingConnection = connection;
      passwordPromptTitle = '输入密码';
      passwordPromptMessage = `连接到 ${connection.name}\n请输入密码：`;
      passwordPromptIsPassword = true;
      passwordPromptShowSave = true;
      showPasswordPrompt = true;
    }
  }

  function handlePasswordSubmit(event) {
    const { value, save } = event.detail;
    showPasswordPrompt = false;

    if (!pendingConnection) return;

    const connection = pendingConnection;
    pendingConnection = null;

    if (connection.auth_type === 'key') {
      // For key auth, value is the passphrase
      onConnect(connection, connection.key_path, value);
    } else {
      // For password auth, value is the password
      if (save) {
        // Save password for future use
        SavePassword(connection.id, value).catch(err => {
          console.error('Failed to save password:', err);
        });
      }
      onConnect(connection, value, '');
    }
  }

  function handlePasswordCancel() {
    showPasswordPrompt = false;
    pendingConnection = null;
    console.log('User cancelled password input');
  }

  function handleEditConnection(connection) {
    console.log('handleEditConnection called:', connection);
    showEditConnectionForm(connection);
  }

  async function handleSelectKeyFile() {
    try {
      const filePath = await SelectSSHKeyFile();
      if (filePath) {
        formData.key_path = filePath;
      }
    } catch (error) {
      console.error('Failed to select key file:', error);
      await showError('选择密钥文件失败: ' + error);
    }
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
      passphrase: '',
      tags: []
    };
    testResult = '';
  }

  function cancelForm() {
    resetForm();
    showConnectionForm = false;
    editingConnection = null;
  }
</script>

<div class="connection-manager">
  <div class="header">
    <h2>SSH 连接</h2>
    <button class="btn-new" on:click={showNewConnectionForm} type="button">
      + 新建连接
    </button>
  </div>

  {#if showConnectionForm}
    <div class="connection-form">
      <h3>{editingConnection ? '编辑连接' : '新建连接'}</h3>

      <div class="form-group">
        <label for="conn-name">连接名称 *</label>
        <input
          id="conn-name"
          type="text"
          bind:value={formData.name}
          placeholder="例如: 生产服务器"
        />
      </div>

      <div class="form-group">
        <label for="conn-host">主机地址 *</label>
        <input
          id="conn-host"
          type="text"
          bind:value={formData.host}
          placeholder="例如: 192.168.1.100 或 example.com"
        />
      </div>

      <div class="form-row">
        <div class="form-group">
          <label for="conn-port">端口</label>
          <input
            id="conn-port"
            type="number"
            bind:value={formData.port}
          />
        </div>
        <div class="form-group">
          <label for="conn-user">用户名 *</label>
          <input
            id="conn-user"
            type="text"
            bind:value={formData.user}
            placeholder="例如: root"
          />
        </div>
      </div>

      <div class="form-group">
        <label for="conn-auth">认证方式</label>
        <select id="conn-auth" bind:value={formData.auth_type}>
          <option value="password">密码</option>
          <option value="key">SSH 密钥</option>
        </select>
      </div>

      {#if formData.auth_type === 'password'}
        <div class="form-group">
          <label for="conn-password">密码</label>
          <input
            id="conn-password"
            type="password"
            bind:value={formData.password}
            placeholder="连接时使用的密码"
            autocomplete="off"
          />
          <div class="checkbox-group">
            <label class="checkbox-label">
              <input
                type="checkbox"
                bind:checked={formData.savePassword}
                disabled
              />
              <span class="checkbox-text">保存密码（开发中 - 将使用系统密钥链加密存储）</span>
            </label>
          </div>
        </div>
      {:else if formData.auth_type === 'key'}
        <div class="form-group">
          <label for="conn-keypath">SSH 私钥文件</label>
          <div class="key-file-selector">
            <input
              id="conn-keypath"
              type="text"
              bind:value={formData.key_path}
              placeholder="点击选择密钥文件 (例如: ~/.ssh/id_rsa)"
              readonly
            />
            <button class="btn-select-file" on:click={handleSelectKeyFile} type="button">
              选择文件
            </button>
          </div>
        </div>
        <div class="form-group">
          <label for="conn-passphrase">Passphrase（可选）</label>
          <input
            id="conn-passphrase"
            type="password"
            bind:value={formData.passphrase}
            placeholder="如果密钥已加密，请输入 passphrase"
            autocomplete="off"
          />
          <div class="hint-text">
            如果您的 SSH 密钥文件已加密，请输入 passphrase。否则留空即可。
          </div>
        </div>
      {/if}

      {#if testResult}
        <div class="test-result" class:success={testResult.includes('成功')} class:error={testResult.includes('失败')}>
          {testResult}
        </div>
      {/if}

      <div class="form-actions">
        <button class="btn-secondary" on:click={cancelForm} type="button">
          取消
        </button>
        <button
          class="btn-secondary"
          on:click={handleTestConnection}
          disabled={testingConnection}
          type="button"
        >
          {testingConnection ? '测试中...' : '测试连接'}
        </button>
        <button class="btn-primary" on:click={handleSaveConnection} type="button">
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
      {#each connections as connection (connection.id)}
        <div class="connection-item">
          <div class="connection-info">
            <div class="connection-name">{connection.name}</div>
            <div class="connection-details">
              {connection.user}@{connection.host}:{connection.port}
            </div>
          </div>
          <div class="connection-actions">
            <button
              class="btn-connect"
              on:click={() => handleConnect(connection)}
              type="button"
            >
              连接
            </button>
            <button
              class="btn-edit"
              on:click={() => handleEditConnection(connection)}
              type="button"
            >
              编辑
            </button>
            <button
              class="btn-delete"
              on:click={() => handleRemoveConnection(connection.id)}
              type="button"
            >
              删除
            </button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<PasswordPrompt
  bind:visible={showPasswordPrompt}
  title={passwordPromptTitle}
  message={passwordPromptMessage}
  isPassword={passwordPromptIsPassword}
  showSaveOption={passwordPromptShowSave}
  on:submit={handlePasswordSubmit}
  on:cancel={handlePasswordCancel}
/>

<style>
  .connection-manager {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 20px;
    background-color: #252526;
    color: #cccccc;
    -webkit-app-region: no-drag !important;
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
    transition: background-color 0.2s;
    -webkit-app-region: no-drag;
  }

  .btn-new:hover {
    background-color: #1177bb;
  }

  .btn-new:active {
    background-color: #0d5a8f;
  }

  .connection-form {
    background-color: #1e1e1e;
    padding: 20px;
    border-radius: 6px;
    margin-bottom: 20px;
    border: 1px solid #3c3c3c;
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

  input[type="text"],
  input[type="number"],
  input[type="password"],
  select {
    width: 100%;
    padding: 8px 10px;
    background-color: #3c3c3c;
    border: 1px solid #555555;
    border-radius: 3px;
    color: #cccccc;
    font-size: 13px;
    box-sizing: border-box;
    transition: border-color 0.2s;
    -webkit-app-region: no-drag;
  }

  input[type="text"]:focus,
  input[type="number"]:focus,
  input[type="password"]:focus,
  select:focus {
    outline: none;
    border-color: #0e639c;
  }

  input[type="text"]:hover,
  input[type="number"]:hover,
  input[type="password"]:hover,
  select:hover {
    border-color: #666666;
  }

  .checkbox-group {
    margin-top: 8px;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    cursor: pointer;
    margin-bottom: 0;
  }

  .checkbox-label input[type="checkbox"] {
    margin-right: 8px;
    cursor: pointer;
    -webkit-app-region: no-drag;
  }

  .checkbox-label input[type="checkbox"]:disabled {
    cursor: not-allowed;
  }

  .checkbox-text {
    font-size: 12px;
    color: #999999;
  }

  .key-file-selector {
    display: flex;
    gap: 10px;
  }

  .key-file-selector input {
    flex: 1;
    background-color: #2a2a2a;
    cursor: default;
  }

  .btn-select-file {
    padding: 8px 16px;
    background-color: #0e639c;
    color: white;
    white-space: nowrap;
  }

  .btn-select-file:hover {
    background-color: #1177bb;
  }

  .hint-text {
    font-size: 11px;
    color: #858585;
    margin-top: 5px;
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

  .btn-primary,
  .btn-secondary {
    padding: 8px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    transition: background-color 0.2s;
    -webkit-app-region: no-drag;
  }

  .btn-primary {
    background-color: #0e639c;
    color: white;
  }

  .btn-primary:hover {
    background-color: #1177bb;
  }

  .btn-primary:active {
    background-color: #0d5a8f;
  }

  .btn-secondary {
    background-color: #3c3c3c;
    color: #cccccc;
  }

  .btn-secondary:hover {
    background-color: #505050;
  }

  .btn-secondary:active {
    background-color: #2a2a2a;
  }

  .btn-secondary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .connections-list {
    flex: 1;
    overflow-y: auto;
    -webkit-app-region: no-drag !important;
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
    border: 1px solid transparent;
    transition: all 0.2s;
    -webkit-app-region: no-drag !important;
  }

  .connection-item:hover {
    background-color: #2a2d2e;
    border-color: #3c3c3c;
  }

  .connection-info {
    flex: 1;
  }

  .connection-name {
    font-weight: 500;
    margin-bottom: 5px;
    font-size: 14px;
  }

  .connection-details {
    font-size: 12px;
    color: #858585;
  }

  .connection-actions {
    display: flex;
    gap: 8px;
    -webkit-app-region: no-drag !important;
  }

  .btn-connect,
  .btn-edit,
  .btn-delete {
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    cursor: pointer !important;
    font-size: 12px;
    transition: background-color 0.2s;
    pointer-events: auto !important;
    position: relative;
    z-index: 10;
    -webkit-app-region: no-drag !important;
  }

  .btn-connect {
    background-color: #0e639c;
    color: white;
  }

  .btn-connect:hover {
    background-color: #1177bb;
  }

  .btn-connect:active {
    background-color: #0d5a8f;
  }

  .btn-edit {
    background-color: #3c3c3c;
    color: #cccccc;
  }

  .btn-edit:hover {
    background-color: #505050;
  }

  .btn-edit:active {
    background-color: #2a2a2a;
  }

  .btn-delete {
    background-color: #3c3c3c;
    color: #cccccc;
  }

  .btn-delete:hover {
    background-color: #a03030;
  }

  .btn-delete:active {
    background-color: #8a2828;
  }
</style>
