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
  import { onMount, onDestroy } from "svelte";
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
  let showCreateMenu = false;
  let passwordVisible = false;
  let passwordLoading = false;
  let hasSavedPassword = false;
  let passwordLoaded = false;
  let passwordLoadError = "";
  let passwordFetchInFlight = false;

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

  $: updateModalOpenClass(showConnectionForm);

  function updateModalOpenClass(isOpen) {
    if (typeof document === "undefined") return;
    document.body.classList.toggle("modal-open", isOpen);
  }

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

  function toggleCreateMenu() {
    showCreateMenu = !showCreateMenu;
  }

  function closeCreateMenu() {
    showCreateMenu = false;
  }

  function handleCreateConnection() {
    showCreateMenu = false;
    showNewConnectionForm();
  }

  async function handleCreateGroup() {
    showCreateMenu = false;
    await showAlert("新建组功能待实现");
  }

  async function handleTodoCreate(label) {
    showCreateMenu = false;
    await showAlert(`${label}功能待实现`);
  }

  function showNewConnectionForm() {
    editingConnection = null;
    resetForm();
    passwordVisible = false;
    passwordLoading = false;
    hasSavedPassword = false;
    passwordLoaded = false;
    passwordLoadError = "";
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
    passwordVisible = false;
    passwordLoading = false;
    hasSavedPassword = false;
    passwordLoaded = false;
    passwordLoadError = "";
    refreshSavedPasswordState(connection.id);
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

      // Save password if user provided one and chose to save it
      if (formData.auth_type === "password" && formData.savePassword && formData.password) {
        try {
          await SavePassword(connectionData.id, formData.password);
          console.log("Password saved for connection:", connectionData.id);
        } catch (error) {
          console.error("Failed to save password:", error);
          await showError("密码保存失败: " + error);
        }
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
        } else {
          // No saved password, show提示 before prompting
          console.log("No saved password found for connection:", connection.id);
          await showAlert(`未保存密码\n连接 ${connection.name} 需要输入密码`);
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
    passwordVisible = false;
    passwordLoading = false;
    hasSavedPassword = false;
    passwordLoaded = false;
    passwordLoadError = "";
  }

  function cancelForm() {
    resetForm();
    showConnectionForm = false;
    editingConnection = null;
  }

  onDestroy(() => {
    updateModalOpenClass(false);
  });

  async function refreshSavedPasswordState(connectionId) {
    if (!connectionId) return;
    try {
      hasSavedPassword = await HasPassword(connectionId);
    } catch (error) {
      console.error("Failed to check saved password:", error);
      hasSavedPassword = false;
    }
  }

  function togglePasswordVisibility() {
    const nextVisible = !passwordVisible;
    passwordVisible = nextVisible;
    if (nextVisible) {
      ensureSavedPasswordLoaded();
    }
  }

  $: if (passwordVisible) {
    ensureSavedPasswordLoaded();
  }

  async function ensureSavedPasswordLoaded() {
    if (passwordFetchInFlight) return;
    if (passwordLoaded || formData.password || !formData.id) return;
    if (!passwordVisible) return;

    passwordFetchInFlight = true;
    passwordLoading = true;
    passwordLoadError = "";
    try {
      const saved = await HasPassword(formData.id);
      hasSavedPassword = saved;
      if (!saved) {
        return;
      }
      formData.password = await GetPassword(formData.id);
      passwordLoaded = true;
    } catch (error) {
      console.error("Failed to load saved password:", error);
      passwordLoadError = "读取已保存密码失败";
    } finally {
      passwordLoading = false;
      passwordFetchInFlight = false;
    }
  }

  // 使用window方法暴露全局函数供onclick使用
  if (typeof window !== "undefined") {
    window.sshToolsConnect = async (connJson) => {
      const connection =
        typeof connJson === "string" ? JSON.parse(connJson) : connJson;
      if (connection) {
        await handleConnect(connection);
      }
    };

    window.sshToolsEdit = (connJson) => {
      const connection =
        typeof connJson === "string" ? JSON.parse(connJson) : connJson;
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
    <div class="header-title">
      <h2>资产</h2>
      <div class="create-menu-wrapper">
        <button class="create-btn" on:click={toggleCreateMenu} title="新建">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M7.25 2.5a.75.75 0 0 1 1.5 0v4.75H13a.75.75 0 0 1 0 1.5H8.75V13a.75.75 0 0 1-1.5 0V8.75H2.5a.75.75 0 0 1 0-1.5h4.75V2.5z"/>
          </svg>
          新建
        </button>
        {#if showCreateMenu}
          <div class="menu-backdrop" on:click={closeCreateMenu}></div>
          <div class="create-menu">
            <button class="menu-item" on:click={handleCreateConnection}>
              新建连接
            </button>
            <button class="menu-item" on:click={handleCreateGroup}>
              新建组
            </button>
            <div class="menu-divider"></div>
            <button class="menu-item disabled" disabled>
              添加数据库连接（待实现）
            </button>
            <button class="menu-item disabled" disabled>
              添加 Docker 连接（待实现）
            </button>
          </div>
        {/if}
      </div>
    </div>
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
      <div class="modal-backdrop" on:click={cancelForm}>
        <div class="modal" on:click|stopPropagation>
          <div class="modal-header">
            <h3>{editingConnection ? "编辑连接" : "新建连接"}</h3>
            <button class="icon-btn" on:click={cancelForm} title="关闭">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M3.72 3.72a.75.75 0 0 1 1.06 0L8 6.94l3.22-3.22a.75.75 0 1 1 1.06 1.06L9.06 8l3.22 3.22a.75.75 0 1 1-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 0 1-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 0 1 0-1.06z"/>
              </svg>
            </button>
          </div>

          <div class="form-box">
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
                <div class="password-field">
                  {#if passwordVisible}
                    <input
                      class="with-toggle"
                      type="text"
                      bind:value={formData.password}
                      placeholder="用于测试连接"
                    />
                  {:else}
                    <input
                      class="with-toggle"
                      type="password"
                      bind:value={formData.password}
                      placeholder="用于测试连接"
                    />
                  {/if}
                  <button
                    type="button"
                    class="toggle-visibility"
                    on:click={togglePasswordVisibility}
                    title={passwordVisible ? "隐藏密码" : "显示已保存密码"}
                    disabled={passwordLoading}
                  >
                    {#if passwordVisible}
                      <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                        <path d="M8 3c3.5 0 6.3 2.2 7.4 5-1.1 2.8-3.9 5-7.4 5S1.7 10.8.6 8C1.7 5.2 4.5 3 8 3zm0 2c-1.7 0-3 1.3-3 3s1.3 3 3 3 3-1.3 3-3-1.3-3-3-3zm0 1.5a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3z"/>
                      </svg>
                    {:else}
                      <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                        <path d="M3.2 2.5a.75.75 0 0 1 1.06 0l9.5 9.5a.75.75 0 1 1-1.06 1.06l-1.6-1.6A7.9 7.9 0 0 1 8 13c-3.5 0-6.3-2.2-7.4-5a8.4 8.4 0 0 1 2.7-3.4L3.2 3.56a.75.75 0 0 1 0-1.06zM8 5c-.4 0-.8.1-1.1.3l1.1 1.1a1.5 1.5 0 0 1 1.9 1.9l1.1 1.1c.2-.3.3-.7.3-1.1 0-1.7-1.3-3-3-3zm-3.3 1.2a6.5 6.5 0 0 0-2.4 1.8c1.1 2.2 3.3 3.7 5.7 3.7.7 0 1.4-.1 2-.3l-1.1-1.1a3 3 0 0 1-4.2-4.1zM8 3c1.6 0 3.1.5 4.4 1.4L11.1 5.7A3 3 0 0 0 6.3 1.9L4.8 3.4A7.7 7.7 0 0 1 8 3z"/>
                      </svg>
                    {/if}
                  </button>
                </div>
                {#if passwordLoadError}
                  <div class="hint-text error-text">{passwordLoadError}</div>
                {/if}
                {#if !hasSavedPassword && editingConnection}
                  <div class="hint-text">未检测到已保存的密码</div>
                {/if}
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
          {#each connections as connection (connection.id)}
            <div class="item">
              <div class="info">
                <div class="name">{connection.name}</div>
                <div class="details">
                  {connection.user}@{connection.host}:{connection.port}
                </div>
              </div>
              <div class="item-actions">
                <button
                  class="act-btn connect-btn"
                  data-connection={JSON.stringify(connection)}
                  onclick="window.sshToolsConnect(this.dataset.connection)"
                >
                  连接
                </button>
                <button
                  class="act-btn edit-btn"
                  data-connection={JSON.stringify(connection)}
                  onclick="window.sshToolsEdit(this.dataset.connection)"
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

  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
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

  .create-menu-wrapper {
    position: relative;
  }

  .create-btn {
    padding: 6px 10px;
    background: var(--accent-subtle);
    color: var(--accent-primary);
    border: 1px solid transparent;
    font-weight: 600;
  }

  .create-btn:hover {
    background: var(--accent-primary);
    color: white;
  }

  .menu-backdrop {
    position: fixed;
    inset: 0;
    background: transparent;
    z-index: 180;
  }

  .create-menu {
    position: absolute;
    top: calc(100% + 8px);
    left: 0;
    width: 220px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-secondary);
    border-radius: 8px;
    box-shadow: var(--shadow-md);
    padding: 6px;
    z-index: 190;
  }

  .menu-item {
    width: 100%;
    padding: 8px 10px;
    background: transparent;
    border: none;
    color: var(--text-primary);
    text-align: left;
    border-radius: 6px;
    font-size: 13px;
  }

  .menu-item:hover {
    background: var(--bg-hover);
  }

  .menu-item.disabled {
    color: var(--text-tertiary);
    cursor: default;
  }

  .menu-item.disabled:hover {
    background: transparent;
  }

  .menu-divider {
    height: 1px;
    background: var(--border-secondary);
    margin: 6px 4px;
  }

  /* Form Styling */
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.45);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    animation: modalFadeIn 160ms ease-out;
  }

  .modal {
    width: min(640px, calc(100vw - 32px));
    max-height: calc(100vh - 80px);
    overflow: hidden;
    background: var(--bg-secondary);
    border: 1px solid var(--border-secondary);
    border-radius: 12px;
    box-shadow: var(--shadow-lg);
    display: flex;
    flex-direction: column;
    animation: modalPopIn 200ms cubic-bezier(0.16, 1, 0.3, 1);
  }

  :global(.modal-open .sidebar-resizer) {
    pointer-events: none;
    cursor: default;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px 0 20px;
  }

  .form-box {
    background: var(--bg-secondary);
    padding: 16px 20px 20px 20px;
    border-radius: 0;
    border: none;
    margin-bottom: 0;
    box-shadow: none;
    overflow-y: auto;
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

  .hint-text.error-text {
    color: var(--accent-error);
  }

  .password-field {
    position: relative;
  }

  .with-toggle {
    padding-right: 36px;
  }

  .toggle-visibility {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    width: 24px;
    height: 24px;
    padding: 0;
    border-radius: 4px;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
  }

  .toggle-visibility:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .toggle-visibility:disabled {
    opacity: 0.5;
    cursor: default;
  }

  @keyframes modalFadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes modalPopIn {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
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
