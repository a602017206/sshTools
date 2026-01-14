<script>
  import { createEventDispatcher } from "svelte";
  import {
    GetConnections,
    AddConnection,
    UpdateConnection,
    RemoveConnection,
    TestConnection,
    SelectSSHKeyFile,
    SavePassword,
    GetPassword,
    HasPassword,
  } from "../../wailsjs/go/main/App.js";
  import { onMount } from "svelte";
  import { showAlert, showError, showConfirm } from "../utils/dialog.js";
  import PasswordPrompt from "./PasswordPrompt.svelte";
  import Settings from "./Settings.svelte";

  export let onConnect = null;

  const dispatch = createEventDispatcher();

  let connections = [];
  let showConnectionForm = false;
  let editingConnection = null;
  let testingConnection = false;
  let testResult = "";

  // Password prompt modal
  let showPasswordPrompt = false;
  let passwordPromptTitle = "";
  let passwordPromptMessage = "";
  let passwordPromptIsPassword = true;
  let passwordPromptShowSave = false;
  let pendingConnection = null;

  // Settings modal
  let showSettings = false;

  let formData = {
    id: "",
    name: "",
    host: "",
    port: 22,
    user: "",
    password: "",
    savePassword: false,
    auth_type: "password",
    key_path: "",
    passphrase: "",
    tags: [],
  };

  onMount(async () => {
    await loadConnections();
  });

  async function loadConnections() {
    try {
      connections = await GetConnections();
      console.log("Loaded connections:", connections);
    } catch (error) {
      console.error("Failed to load connections:", error);
      connections = [];
    }
  }

  function openSettings() {
    showSettings = true;
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
      password: "",
      savePassword: false,
      auth_type: connection.auth_type || "password",
      key_path: connection.key_path || "",
      passphrase: "",
      tags: connection.tags || [],
    };
    showConnectionForm = true;
  }

  async function handleSaveConnection() {
    if (!formData.name || !formData.host || !formData.user) {
      await showAlert("请填写必填字段（连接名称、主机地址、用户名）");
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
        tags: formData.tags,
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
      console.error("Failed to save connection:", error);
      await showError("保存连接失败: " + error);
    }
  }

  async function handleRemoveConnection(id) {
    console.log("🔴 handleRemoveConnection called for id:", id);

    const confirmed = await showConfirm("确定要删除此连接吗？");
    if (!confirmed) {
      console.log("用户取消了删除操作");
      return;
    }

    try {
      await RemoveConnection(id);
      await loadConnections();
      console.log("连接已删除:", id);
    } catch (error) {
      console.error("Failed to remove connection:", error);
      await showError("删除连接失败: " + error);
    }
  }

  async function handleTestConnection() {
    if (!formData.host || !formData.user) {
      await showAlert("请填写主机地址和用户名");
      return;
    }

    // Validate based on auth type
    if (formData.auth_type === "password") {
      if (!formData.password) {
        await showAlert("请输入密码以测试连接");
        return;
      }
    } else if (formData.auth_type === "key") {
      if (!formData.key_path) {
        await showAlert("请选择 SSH 密钥文件");
        return;
      }
    }

    testingConnection = true;
    testResult = "";

    try {
      const authValue =
        formData.auth_type === "key" ? formData.key_path : formData.password;
      await TestConnection(
        formData.host,
        parseInt(formData.port),
        formData.user,
        formData.auth_type,
        authValue,
        formData.passphrase || "",
      );
      testResult = "✓ 连接成功";
    } catch (error) {
      console.error("Connection test failed:", error);
      testResult = "✗ 连接失败: " + error;
    } finally {
      testingConnection = false;
    }
  }

  async function handleConnect(connection) {
    console.log("🔵 handleConnect called:", connection);

    if (!onConnect) {
      console.error("onConnect callback not provided");
      return;
    }

    if (connection.auth_type === "key") {
      // For key auth, use saved key path and prompt for passphrase
      pendingConnection = connection;
      passwordPromptTitle = "密钥 Passphrase";
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
          console.log("Using saved password");
          onConnect(connection, password, "");
          return;
        }
      } catch (error) {
        console.error("Failed to get saved password:", error);
      }

      // No saved password, prompt user
      pendingConnection = connection;
      passwordPromptTitle = "输入密码";
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

    if (connection.auth_type === "key") {
      // For key auth, value is the passphrase
      onConnect(connection, connection.key_path, value);
    } else {
      // For password auth, value is the password
      if (save) {
        // Save password for future use
        SavePassword(connection.id, value).catch((err) => {
          console.error("Failed to save password:", err);
        });
      }
      onConnect(connection, value, "");
    }
  }

  function handlePasswordCancel() {
    showPasswordPrompt = false;
    pendingConnection = null;
    console.log("User cancelled password input");
  }

  function handleEditConnection(connection) {
    console.log("handleEditConnection called:", connection);
    showEditConnectionForm(connection);
  }

  async function handleSelectKeyFile() {
    try {
      const filePath = await SelectSSHKeyFile();
      if (filePath) {
        formData.key_path = filePath;
      }
    } catch (error) {
      console.error("Failed to select key file:", error);
      await showError("选择密钥文件失败: " + error);
    }
  }

  function resetForm() {
    formData = {
      id: "",
      name: "",
      host: "",
      port: 22,
      user: "",
      password: "",
      savePassword: false,
      auth_type: "password",
      key_path: "",
      passphrase: "",
      tags: [],
    };
    testResult = "";
  }

  function cancelForm() {
    resetForm();
    showConnectionForm = false;
    editingConnection = null;
  }

  // 使用window方法暴露全局函数供onclick使用
  if (typeof window !== "undefined") {
    window.sshToolsConnect = async (index) => {
      const connection = connections[index];
      if (connection) {
        await handleConnect(connection);
      }
    };

    window.sshToolsEdit = (index) => {
      const connection = connections[index];
      if (connection) {
        handleEditConnection(connection);
      }
    };

    window.sshToolsDelete = async (id) => {
      await handleRemoveConnection(id);
    };
  }
</script>

<div class="manager">
  <div class="header-bar">
    <h2>资产</h2>
    <div class="header-actions">
      <button class="icon-btn" on:click={openSettings} title="设置">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path
            d="M14 7v2h-2.1c-.1.5-.3 1-.6 1.4l1.5 1.5-1.4 1.4-1.5-1.5c-.4.3-.9.5-1.4.6V14H7v-2.1c-.5-.1-1-.3-1.4-.6l-1.5 1.5L2.7 11.4l1.5-1.5c-.3-.4-.5-.9-.6-1.4H2V7h2.1c.1-.5.3-1 .6-1.4L3.2 4.1 4.6 2.7l1.5 1.5C6.5 4 7 3.8 7.5 3.7V2h2v1.7c.5.1 1 .3 1.4.6l1.5-1.5 1.4 1.4-1.5 1.5c.3.4.5.9.6 1.4H14zm-5.5 3c1.4 0 2.5-1.1 2.5-2.5S9.9 5 8.5 5 6 6.1 6 7.5 7.1 10 8.5 10z"
          />
        </svg>
      </button>
      <button
        class="icon-btn"
        on:click={() => dispatch("collapse")}
        title="收起侧边栏"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path
            fill-rule="evenodd"
            clip-rule="evenodd"
            d="M14 2.5H2V3.5H14V2.5ZM2 7.5H14V8.5H2V7.5ZM2 12.5H14V13.5H2V12.5ZM5 2.5V13.5H4V2.5H5Z"
          />
        </svg>
      </button>
    </div>
  </div>

  <div class="content-area">
    {#if showConnectionForm}
      <div class="form-box">
        <h3>{editingConnection ? "编辑连接" : "新建连接"}</h3>

        <div class="field">
          <label>连接名称 *</label>
          <input
            type="text"
            bind:value={formData.name}
            placeholder="例如: 生产服务器"
          />
        </div>

        <div class="field">
          <label>主机地址 *</label>
          <input
            type="text"
            bind:value={formData.host}
            placeholder="例如: 192.168.1.100"
          />
        </div>

        <div class="field-row">
          <div class="field">
            <label>端口</label>
            <input type="number" bind:value={formData.port} />
          </div>
          <div class="field">
            <label>用户名 *</label>
            <input
              type="text"
              bind:value={formData.user}
              placeholder="例如: root"
            />
          </div>
        </div>

        <div class="field">
          <label>认证方式</label>
          <select bind:value={formData.auth_type}>
            <option value="password">密码</option>
            <option value="key">SSH 密钥</option>
          </select>
        </div>

        {#if formData.auth_type === "password"}
          <div class="field">
            <label>密码</label>
            <input
              type="password"
              bind:value={formData.password}
              placeholder="用于测试连接"
            />
          </div>
        {:else if formData.auth_type === "key"}
          <div class="field">
            <label>SSH 私钥文件</label>
            <div class="key-file-selector">
              <input
                type="text"
                bind:value={formData.key_path}
                placeholder="点击选择密钥文件"
                readonly
              />
              <button
                class="btn-select-file"
                on:click={handleSelectKeyFile}
                type="button"
              >
                选择文件
              </button>
            </div>
          </div>
          <div class="field">
            <label>Passphrase（可选）</label>
            <input
              type="password"
              bind:value={formData.passphrase}
              placeholder="如果密钥已加密，请输入 passphrase"
            />
            <div class="hint-text">
              如果您的 SSH 密钥文件已加密，请输入 passphrase。否则留空即可。
            </div>
          </div>
        {/if}

        {#if testResult}
          <div
            class="result {testResult.includes('成功') ? 'success' : 'error'}"
          >
            {testResult}
          </div>
        {/if}

        <div class="actions">
          <button on:click={cancelForm}>取消</button>
          <button on:click={handleTestConnection} disabled={testingConnection}>
            {testingConnection ? "测试中..." : "测试连接"}
          </button>
          <button on:click={handleSaveConnection} class="primary">保存</button>
        </div>
      </div>
    {:else}
      <div class="list">
        {#if connections.length === 0}
          <div class="empty">
            <p>暂无连接</p>
            <p>点击下方"新建连接"开始添加</p>
          </div>
        {:else}
          {#each connections as connection, index (connection.id)}
            <div class="item">
              <div class="info">
                <div class="name">{connection.name}</div>
                <div class="details">
                  {connection.user}@{connection.host}:{connection.port}
                </div>
              </div>
              <div class="item-actions">
                <!-- 使用原生onclick和索引 -->
                <button
                  class="act-btn connect-btn"
                  onclick="window.sshToolsConnect({index})"
                >
                  连接
                </button>
                <button
                  class="act-btn edit-btn"
                  onclick="window.sshToolsEdit({index})"
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
    {/if}
  </div>

  <div class="footer-bar">
    {#if !showConnectionForm}
      <button
        class="new-btn full-width"
        onclick="document.getElementById('new-conn-trigger').click()"
      >
        + 新建连接
      </button>
      <button
        id="new-conn-trigger"
        style="display:none"
        on:click={showNewConnectionForm}
      ></button>
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

<Settings bind:visible={showSettings} />

<style>
  .manager {
    height: 100%;
    padding: 0;
    background: var(--bg-secondary);
    color: var(--text-primary);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .header-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-secondary);
    background: var(--bg-secondary);
    flex-shrink: 0;
  }

  h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.3px;
  }

  h3 {
    margin: 0 0 20px 0;
    font-size: 18px;
    color: var(--text-primary);
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }

  /* Content with custom scrollbar */
  .content-area {
    flex: 1;
    overflow-y: auto;
    min-height: 0;
    padding: 16px;
  }

  .footer-bar {
    padding: 16px;
    border-top: 1px solid var(--border-secondary);
    background: var(--bg-secondary);
    flex-shrink: 0;
  }

  /* Buttons & Icons */
  .icon-btn {
    width: 28px;
    height: 28px;
    padding: 0;
    background: transparent;
    color: var(--text-secondary);
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .icon-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  button {
    font-family: inherit;
    font-size: 13px;
    font-weight: 500;
    padding: 8px 16px;
    border-radius: 6px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }

  .new-btn {
    background: var(--accent-primary);
    color: white;
    box-shadow: var(--shadow-sm);
    width: 100%;
    padding: 10px;
    font-size: 14px;
  }

  .new-btn:hover {
    background: var(--accent-hover);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }

  /* Form Styling */
  .form-box {
    background: var(--bg-tertiary);
    padding: 24px;
    border-radius: 8px;
    border: 1px solid var(--border-secondary);
    margin-bottom: 20px;
    box-shadow: var(--shadow-md);
  }

  .field {
    margin-bottom: 16px;
  }

  .field-row {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 16px;
  }

  label {
    display: block;
    margin-bottom: 6px;
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
  }

  input,
  select {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-input);
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 13px;
    transition: all 0.2s;
    outline: none;
  }

  input:focus,
  select:focus {
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 2px var(--accent-subtle);
    background: var(--bg-input-focus);
  }

  /* Key File Selector */
  .key-file-selector {
    display: flex;
    gap: 8px;
  }

  .key-file-selector input {
    background: var(--bg-input);
    cursor: default;
  }

  .btn-select-file {
    padding: 0 12px;
    background: var(--bg-hover);
    color: var(--text-primary);
    border: 1px solid var(--border-primary);
  }

  .btn-select-file:hover {
    background: var(--bg-active);
  }

  .hint-text {
    font-size: 12px;
    color: var(--text-tertiary);
    margin-top: 6px;
    line-height: 1.4;
  }

  /* Result Box */
  .result {
    padding: 12px;
    border-radius: 6px;
    margin-bottom: 20px;
    font-size: 13px;
    display: flex;
    align-items: center;
  }

  .result.success {
    background: rgba(16, 185, 129, 0.1);
    color: var(--accent-success);
    border: 1px solid rgba(16, 185, 129, 0.2);
  }

  .result.error {
    background: rgba(239, 68, 68, 0.1);
    color: var(--accent-error);
    border: 1px solid rgba(239, 68, 68, 0.2);
  }

  /* Form Actions */
  .actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid var(--border-secondary);
  }

  .actions button {
    background: var(--bg-hover);
    color: var(--text-primary);
    border: 1px solid var(--border-primary);
  }

  .actions button:hover {
    background: var(--bg-active);
  }

  .actions button.primary {
    background: var(--accent-primary);
    color: white;
    border: 1px solid transparent;
  }

  .actions button.primary:hover {
    background: var(--accent-hover);
  }

  /* Connection List */
  .list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .empty {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-tertiary);
  }

  .empty p:first-child {
    font-size: 16px;
    margin-bottom: 8px;
    color: var(--text-secondary);
  }

  .item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: var(--bg-secondary); /* or bg-tertiary? */
    border: 1px solid transparent;
    border-radius: 8px;
    transition: all 0.2s ease;
    cursor: default;
  }

  .item:hover {
    background: var(--bg-hover);
    border-color: var(--border-secondary);
    transform: translateY(-1px);
    box-shadow: var(--shadow-sm);
  }

  .info {
    flex: 1;
    min-width: 0;
  }

  .name {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 4px;
    color: var(--text-primary);
  }

  .details {
    font-size: 12px;
    color: var(--text-secondary);
    font-family: monospace; /* For IP/User alignment */
  }

  .item-actions {
    display: flex;
    gap: 6px;
    opacity: 0.6;
    transition: opacity 0.2s;
  }

  .item:hover .item-actions {
    opacity: 1;
  }

  .act-btn {
    padding: 6px 10px;
    font-size: 12px;
    border-radius: 4px;
    background: var(--bg-input);
    color: var(--text-secondary);
    border: 1px solid var(--border-secondary);
  }

  .act-btn:hover {
    background: var(--bg-active);
    color: var(--text-primary);
    border-color: var(--border-primary);
  }

  .connect-btn {
    background: var(--accent-subtle);
    color: var(--accent-primary);
    border-color: transparent;
    font-weight: 600;
  }

  .connect-btn:hover {
    background: var(--accent-primary);
    color: white;
  }

  .delete-btn:hover {
    background: rgba(239, 68, 68, 0.1);
    color: var(--accent-error);
    border-color: rgba(239, 68, 68, 0.3);
  }
</style>
