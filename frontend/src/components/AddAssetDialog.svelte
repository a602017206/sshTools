<script>
  import Dialog from './ui/Dialog.svelte';
  import { assetsStore } from '../stores.js';

  export let isOpen = false;
  export let onAdd = () => {};
  export let onUpdate = () => {};
  export let editingAsset = null;

  let assetType = 'ssh';
  let authType = 'password';
  let testingConnection = false;
  let testResult = '';
  let savingConnection = false;
  let showPassword = false;
  let showPassphrase = false;
  let editingAssetLoaded = false; // Flag to prevent reloading data
  let wasOpen = false; // Track previous open state
  let lastEditingAssetId = null;
  const labelClass = 'block text-xs font-medium ops-field-label mb-1';
  const inputClass = 'w-full px-3 py-1.5 ops-input border rounded-md text-xs focus:outline-none focus:ring-2 transition-all';
  const passwordInputClass = `${inputClass} pr-10`;
  const choiceClass = (active) => `p-2 rounded-lg border-2 transition-all ops-choice ${active ? 'ops-choice-active' : ''}`;
  const authChoiceClass = (active) => `flex-1 px-3 py-1.5 rounded-md border-2 transition-all ops-choice ${active ? 'ops-choice-active' : ''}`;
  const databaseTypes = [
    { value: 'mysql', label: 'MySQL', port: '3306' },
    { value: 'postgresql', label: 'PostgreSQL', port: '5432' },
    { value: 'sqlite', label: 'SQLite', port: '0' },
    { value: 'oracle', label: 'Oracle', port: '1521' },
    { value: 'sqlserver', label: 'SQL Server', port: '1433' },
    { value: 'dm', label: '达梦 DM', port: '5236' },
    { value: 'kingbase', label: '人大金仓 Kingbase', port: '54321' },
    { value: 'opengauss', label: 'openGauss', port: '5432' }
  ];

  // Group selector state
  let showGroupDropdown = false;
  let groupSearchTerm = '';
  let selectedGroupIndex = -1;
  let jdbcDrivers = [];
  let jdbcDriversLoaded = false;

  let formData = {
    id: '',
    name: '',
    host: '',
    port: '',
    username: '',
    password: '',
    savePassword: false,
    keyPath: '',
    passphrase: '',
    group: '',
    dbType: 'mysql',
    driverProfileID: '',
    database: '',
  };

  // Extract all unique groups from existing assets
  $: allGroups = $assetsStore.reduce((groups, asset) => {
    if (asset.group && !groups.includes(asset.group)) {
      groups.push(asset.group);
    }
    return groups;
  }, []);

  // Filter groups based on search term
  $: filteredGroups = allGroups.filter(group =>
    group.toLowerCase().includes(groupSearchTerm.toLowerCase())
  );

  // Add custom input as an option if it doesn't match existing groups
  $: availableGroups = formData.group && !filteredGroups.includes(formData.group)
    ? [...filteredGroups, formData.group]
    : filteredGroups;
  $: isSQLiteDatabase = assetType === 'database' && formData.dbType === 'sqlite';
  $: selectedJDBCDriver = jdbcDrivers.find(driver => driver.id === formData.dbType);
  $: jdbcProfiles = selectedJDBCDriver?.profiles || [];
  $: selectedJDBCProfile =
    jdbcProfiles.find(profile => profile.id === formData.driverProfileID) ||
    jdbcProfiles.find(profile => profile.version === selectedJDBCDriver?.recommendedVersion) ||
    jdbcProfiles[0] ||
    null;
  $: selectedJDBCProfileMissing = assetType === 'database' && selectedJDBCProfile && !selectedJDBCProfile.installed;

  async function handleTestConnection() {
    if (!window.wailsBindings) {
      testResult = 'Wails 绑定未加载';
      return;
    }

    if (!isSQLiteDatabase && !formData.host) {
      testResult = '请填写主机地址';
      return;
    }

    if (selectedJDBCProfileMissing) {
	  testResult = `请先在全局设置中安装 ${selectedJDBCDriver.name} ${selectedJDBCProfile.version} JDBC 驱动`;
      return;
    }

    if (authType === 'password' && !isSQLiteDatabase && !formData.password) {
      testResult = '请输入密码以测试连接';
      return;
    }

    if (authType === 'password' && !isSQLiteDatabase && !formData.username) {
      testResult = '请填写用户名';
      return;
    }


    if (authType === 'key' && !formData.keyPath) {
      testResult = '请选择 SSH 密钥文件';
      return;
    }

    testingConnection = true;
    testResult = '';

    try {
      if (assetType === 'database') {
        await window.wailsBindings.TestDatabaseConnection(
          isSQLiteDatabase ? '' : formData.host,
          parseInt(formData.port),
          isSQLiteDatabase ? '' : formData.username,
          isSQLiteDatabase ? '' : formData.password,
          formData.dbType,
          formData.database
        );
      } else {
        const authValue = authType === 'key' ? formData.keyPath : formData.password;
        await window.wailsBindings.TestConnection(
          formData.host,
          parseInt(formData.port),
          formData.username,
          authType,
          authValue,
          formData.passphrase || ''
        );
      }
      testResult = '✓ 连接成功';
    } catch (error) {
      console.error('Connection test failed:', error);
      testResult = '✗ 连接失败: ' + error;
    } finally {
      testingConnection = false;
    }
  }

  async function handleSelectKeyFile() {
    if (!window.wailsBindings) {
      alert('Wails 绑定未加载');
      return;
    }

    try {
      const filePath = await window.wailsBindings.SelectSSHKeyFile();
      if (filePath) {
        formData.keyPath = filePath;
      }
    } catch (error) {
      console.error('Failed to select key file:', error);
      testResult = '选择密钥文件失败: ' + error;
    }
  }

  async function handleSubmit() {
    if (!window.wailsBindings) {
      alert('Wails 绑定未加载');
      return;
    }

    if (!formData.name || (!isSQLiteDatabase && !formData.host)) {
      alert('请填写必填字段（连接名称、主机地址）');
      return;
    }

    if (isSQLiteDatabase && !formData.database) {
      alert('请填写 SQLite 数据库文件路径或 JDBC URL');
      return;
    }

    if (selectedJDBCProfileMissing) {
	  alert(`请先在全局设置中安装 ${selectedJDBCDriver.name} ${selectedJDBCProfile.version} JDBC 驱动`);
      return;
    }

    if (authType === 'password' && !isSQLiteDatabase && !formData.username) {
      alert('密码认证需要填写用户名');
      return;
    }

    if (authType === 'password' && !isSQLiteDatabase && !formData.password) {
      alert('密码认证需要输入密码');
      return;
    }

    if (authType === 'key' && !formData.keyPath) {
      alert('密钥认证需要选择密钥文件');
      return;
    }


    savingConnection = true;

    try {
      const isEdit = !!editingAsset?.id;
      const connectionData = {
        id: isEdit ? editingAsset.id : `conn_${Date.now()}`,
        name: formData.name,
        host: isSQLiteDatabase ? '' : formData.host,
        port: parseInt(formData.port),
        user: isSQLiteDatabase ? '' : formData.username,
        auth_type: authType,
        key_path: authType === 'key' ? formData.keyPath : '',
        tags: [formData.group || '默认分组'],
        type: assetType,
        metadata: {
          database: formData.database || undefined,
          db_type: formData.dbType,
          driver_profile_id: formData.driverProfileID || undefined
        }
      };

      if (isEdit) {
        // Update existing connection
        await window.wailsBindings.UpdateConnection(connectionData);

        // Save password if checkbox is checked
        if (authType === 'password' && formData.password && formData.savePassword) {
          await window.wailsBindings.SavePassword(connectionData.id, formData.password);
        } else if (authType === 'password' && !formData.savePassword) {
          // Remove saved password if checkbox is unchecked
          await window.wailsBindings.DeletePassword(connectionData.id);
        }

        onUpdate(connectionData);
      } else {
        // Add new connection
        onAdd(connectionData);

        // Save password if checkbox is checked
        if (authType === 'password' && formData.password && formData.savePassword) {
          await window.wailsBindings.SavePassword(connectionData.id, formData.password);
        }
      }

      editingAsset = null;
      isOpen = false;
    } catch (error) {
      console.error('Failed to save connection:', error);
      alert('保存连接失败: ' + error);
    } finally {
      savingConnection = false;
    }
  }

  function resetForm() {
    assetType = 'ssh'; // Reset asset type before computing defaults
    authType = 'password'; // Reset auth type
    formData = {
      id: '',
      name: '',
      host: '',
      port: getDefaultPortFor('ssh', 'mysql'),
      username: '',
      password: '',
      savePassword: false,
      keyPath: '',
      passphrase: '',
      group: '',
      dbType: 'mysql',
      driverProfileID: '',
      database: '',
    };
    testResult = '';
    showPassword = false;
    showPassphrase = false;
    showGroupDropdown = false;
    groupSearchTerm = '';
    selectedGroupIndex = -1;
    jdbcDriversLoaded = false;
    editingAssetLoaded = false; // Reset the loaded flag
  }

  $: if (wasOpen && !isOpen) {
    // Dialog was just closed, reset the form
    resetForm();
  }
  $: wasOpen = isOpen;

  $: if (isOpen && !editingAsset && !formData.id) {
    // New connection: only set port if form is empty
    if (!formData.port) {
      formData.port = getDefaultPort();
    }
  }

  $: if (isOpen) {
    const currentEditingAssetId = editingAsset?.id || null;
    if (currentEditingAssetId !== lastEditingAssetId) {
      if (!currentEditingAssetId) {
        resetForm();
      } else {
        editingAssetLoaded = false;
      }
      lastEditingAssetId = currentEditingAssetId;
    }
  }

  function getDefaultPortFor(type, dbType) {
    switch (type) {
      case 'ssh': return '22';
      case 'docker': return '2375';
      case 'database':
        return databaseTypes.find(type => type.value === dbType)?.port || '';
      default: return '';
    }
  }

  $: if (formData.port === '') {
    formData.port = getDefaultPort();
  }

  function getDefaultPort() {
    return getDefaultPortFor(assetType, formData.dbType);
  }

  function selectAssetType(type) {
    assetType = type;
    testResult = '';
    if (!editingAsset) {
      formData.port = getDefaultPortFor(type, formData.dbType);
    }
  }

  function handleDatabaseTypeChange(event) {
	const driver = jdbcDrivers.find(item => item.id === event.currentTarget.value);
	formData.driverProfileID = driver?.profiles?.find(profile => profile.version === driver.recommendedVersion)?.id || driver?.profiles?.[0]?.id || '';
    if (!editingAsset && assetType === 'database') {
      formData.port = getDefaultPortFor('database', event.currentTarget.value);
    }
    testResult = '';
  }

  $: if (assetType === 'ssh') {
    if (authType !== 'password' && authType !== 'key') {
      authType = 'password';
    }
  } else {
    // 数据库和 Docker 只支持密码认证
    authType = 'password';
  }

  // Load connection data when editing
  async function loadConnectionData() {
    if (!editingAsset || !isOpen || editingAssetLoaded) return;

    try {
      const conn = await window.wailsBindings.GetConnection(editingAsset.id);
      if (conn) {
        formData = {
          id: conn.id,
          name: conn.name || '',
          host: conn.host || '',
          port: conn.port?.toString() || '',
          username: conn.user || '',
          password: '', // Will load separately
          savePassword: false, // Will check separately
          keyPath: conn.key_path || '',
          passphrase: '',
          group: conn.tags?.[0] || '',
          dbType: conn.metadata?.db_type || 'mysql',
		  driverProfileID: conn.metadata?.driver_profile_id || '',
          database: conn.metadata?.database || '',
        };
        assetType = conn.type || 'ssh';
        authType = conn.auth_type || 'password';
        editingAssetLoaded = true; // Mark as loaded

        // Load password if saved
        try {
          const hasPassword = await window.wailsBindings.HasPassword(conn.id);
          if (hasPassword) {
            formData.savePassword = true;
            const password = await window.wailsBindings.GetPassword(conn.id);
            formData.password = password || '';
          }
        } catch (error) {
          console.warn('Failed to load password:', error);
        }
      }
    } catch (error) {
      console.error('Failed to load connection:', error);
    }
  }

  $: if (isOpen && editingAsset) {
    loadConnectionData();
  }

  $: if (isOpen && assetType === 'database' && !jdbcDriversLoaded) {
    loadJDBCDrivers();
  }

  // Reset the loaded flag when editingAsset changes or dialog closes
  $: if (editingAsset && formData.id !== editingAsset.id) {
    // editingAsset has changed to a different one, reload data
    editingAssetLoaded = false;
  }

  $: if (!wasOpen && isOpen && editingAsset) {
    // Dialog just opened with editingAsset, ensure we load data
    editingAssetLoaded = false;
  }

  // Group selector handlers
  function toggleGroupDropdown() {
    showGroupDropdown = !showGroupDropdown;
    if (showGroupDropdown) {
      groupSearchTerm = formData.group || '';
      // Find index of selected group
      selectedGroupIndex = availableGroups.indexOf(formData.group);
    }
  }

  function handleGroupSearchInput(e) {
    groupSearchTerm = e.target.value;
    formData.group = e.target.value;
    showGroupDropdown = true;
    selectedGroupIndex = -1;
    editingAssetLoaded = false; // Reset the loaded flag
    assetType = 'ssh'; // Reset asset type
    authType = 'password'; // Reset auth type
  }

  function selectGroup(group) {
    formData.group = group;
    groupSearchTerm = group;
    showGroupDropdown = false;
    selectedGroupIndex = availableGroups.indexOf(group);
  }

  function handleGroupKeydown(e) {
    const groups = availableGroups;
    if (!showGroupDropdown) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter') {
        e.preventDefault();
        showGroupDropdown = true;
        groupSearchTerm = formData.group || '';
        selectedGroupIndex = groups.indexOf(formData.group);
      }
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        selectedGroupIndex = Math.min(selectedGroupIndex + 1, groups.length - 1);
        break;
      case 'ArrowUp':
        e.preventDefault();
        selectedGroupIndex = Math.max(selectedGroupIndex - 1, 0);
        break;
      case 'Enter':
        e.preventDefault();
        if (selectedGroupIndex >= 0 && selectedGroupIndex < groups.length) {
          selectGroup(groups[selectedGroupIndex]);
        } else {
          showGroupDropdown = false;
        }
        break;
      case 'Escape':
        e.preventDefault();
        showGroupDropdown = false;
        groupSearchTerm = formData.group || '';
        break;
      case 'Tab':
        e.preventDefault();
        showGroupDropdown = false;
        break;
    }
  }

  // Close dropdown when clicking outside
  function handleGroupBlur() {
    setTimeout(() => {
      showGroupDropdown = false;
    }, 150);
  }

  async function loadJDBCDrivers() {
    if (!window.wailsBindings || typeof window.wailsBindings.ListJDBCDrivers !== 'function') {
      return;
    }
    try {
      jdbcDrivers = await window.wailsBindings.ListJDBCDrivers();
	  if (!formData.driverProfileID) {
		const driver = jdbcDrivers.find(item => item.id === formData.dbType);
		formData.driverProfileID = driver?.profiles?.find(profile => profile.version === driver.recommendedVersion)?.id || driver?.profiles?.[0]?.id || '';
	  }
      jdbcDriversLoaded = true;
    } catch (error) {
      console.warn('Failed to load JDBC driver status:', error);
      jdbcDriversLoaded = true;
    }
  }
</script>

<Dialog
   bind:isOpen={isOpen}
   onClose={() => {
     isOpen = false;
   }}
   title={editingAsset ? "编辑连接" : "添加连接"}
   size="sm"
  >
   <div class="mb-4">
    <div class="block text-xs font-medium ops-field-label mb-2">连接类型</div>
    <div class="grid grid-cols-3 gap-2">
      <button
        type="button"
        on:click={() => selectAssetType('ssh')}
        class={choiceClass(assetType === 'ssh')}
      >
        <div class="flex items-center justify-center gap-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          <span class="text-xs font-medium">SSH</span>
        </div>
      </button>

      <button
        type="button"
        on:click={() => selectAssetType('database')}
        class={choiceClass(assetType === 'database')}
      >
        <div class="flex items-center justify-center gap-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
          </svg>
          <span class="text-xs font-medium">数据库</span>
        </div>
      </button>

      <button
        type="button"
        on:click={() => selectAssetType('docker')}
        class={choiceClass(assetType === 'docker')}
      >
        <div class="flex items-center justify-center gap-1">
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M13.983 11.078h2.119a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.119a.185.185 0 00-.185.185v1.888c0 .102.083.185.185.185m-2.954-3.333h2.118a.186.186 0 00.186-.186V5.671a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.185m-2.954 3.333h2.118a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186H8.075a.186.186 0 00-.186.186v1.888c0 .102.083.185.186.185m-2.954-3.333h2.119a.186.186 0 00.185-.186V5.671a.185.185 0 00-.185-.185H5.12a.186.186 0 00-.186.185v1.888c0 .102.084.185.186.185m-2.93 3.333h2.12a.185.185 0 00.185-.185V9.006a.185.185 0 00-.186-.186h-2.12a.185.185 0 00-.184.186v1.888c0 .102.083.185.185.185M20.69 6.662c.057.16.09.331.09.51v7.9c0 3.058-2.724 4.928-8.78 4.928-6.055 0-8.779-1.87-8.779-4.928v-7.9c0-.179.033-.35.09-.51C1.536 7.396 0 9.522 0 12.072v3.639c0 4.072 3.608 6.789 12 6.789 8.391 0 12-2.717 12-6.79v-3.638c0-2.55-1.536-4.677-4.31-6.41" />
          </svg>
          <span class="text-xs font-medium">Docker</span>
        </div>
      </button>
    </div>
   </div>

   <form on:submit|preventDefault={handleSubmit} class="space-y-3">
      <div>
       <label class={labelClass} for="connection-name">
         连接名称 <span class="ops-required">*</span>
       </label>
       <input
         type="text"
         id="connection-name"
         required
         bind:value={formData.name}
         placeholder="例如：生产服务器-01"
         class={inputClass}
       />
     </div>

     {#if assetType === 'ssh'}
       <div>
         <div class={labelClass}>
           认证方式
         </div>
         <div class="flex gap-2">
           <button
             type="button"
             on:click={() => { authType = 'password'; testResult = ''; }}
             class={authChoiceClass(authType === 'password')}
           >
             <div class="text-xs font-medium">密码</div>
           </button>
           <button
             type="button"
             on:click={() => { authType = 'key'; testResult = ''; }}
             class={authChoiceClass(authType === 'key')}
           >
             <div class="text-xs font-medium">SSH 密钥</div>
           </button>
         </div>
       </div>
     {/if}
      {#if authType === 'key'}
        <div>
          <label class={labelClass} for="ssh-key-path">
            SSH 私钥文件
          </label>
          <div class="flex gap-1.5">
            <input
              type="text"
              id="ssh-key-path"
              bind:value={formData.keyPath}
              placeholder="点击选择密钥文件 (例如: ~/.ssh/id_rsa)"
              readonly
              class="flex-1 px-3 py-1.5 ops-input ops-input-readonly border rounded-md text-xs focus:outline-none transition-all"
            />
            <button
              type="button"
              on:click={handleSelectKeyFile}
              class="px-3 py-1.5 accent-gradient rounded-md text-xs font-medium transition-all whitespace-nowrap"
            >
              选择文件
            </button>
          </div>
        </div>

        <div>
          <label class={labelClass} for="ssh-key-passphrase">
            Passphrase（可选）
          </label>
          <div class="relative">
            {#if showPassphrase}
              <input
                type="text"
                id="ssh-key-passphrase"
                bind:value={formData.passphrase}
                placeholder="如果密钥已加密，请输入 passphrase"
                class={passwordInputClass}
              />
            {:else}
              <input
                type="password"
                id="ssh-key-passphrase"
                bind:value={formData.passphrase}
                placeholder="如果密钥已加密，请输入 passphrase"
                class={passwordInputClass}
              />
            {/if}
            <button
              type="button"
              on:click={() => showPassphrase = !showPassphrase}
              class="absolute right-2 top-1/2 transform -translate-y-1/2 p-1 ops-icon-inline rounded transition-colors"
            >
              {#if showPassphrase}
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              {:else}
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              {/if}
            </button>
          </div>
          <p class="mt-0.5 text-[10px] ops-help">
            如果您的 SSH 密钥文件已加密，请输入 passphrase。否则留空即可。
          </p>
        </div>
      {/if}

     {#if testResult}
       <div class={`p-2 rounded-md text-xs ${
         testResult.includes('成功') ? 'ops-status-success' : 'ops-status-error'
       }`}>
         {testResult}
       </div>
      {/if}

     {#if assetType === 'database'}
       <div>
         <label class={labelClass} for="database-type">
           数据库类型 <span class="ops-required">*</span>
         </label>
        <select
          id="database-type"
          bind:value={formData.dbType}
          on:change={handleDatabaseTypeChange}
          class={inputClass}
        >
            {#each databaseTypes as db}
              <option value={db.value}>{db.label}</option>
            {/each}
         </select>
       </div>
	   {#if selectedJDBCProfileMissing}
         <div class="p-2 rounded-md text-xs ops-status-error">
           请先在全局设置中安装 {selectedJDBCDriver.name} {selectedJDBCProfile.version} JDBC 驱动。
         </div>
       {/if}
	   {#if jdbcProfiles.length > 0}
		 <div>
		   <label class={labelClass} for="database-driver-profile">JDBC 驱动版本</label>
		   <select id="database-driver-profile" bind:value={formData.driverProfileID} class={inputClass}>
			 {#each jdbcProfiles as profile}
			   <option value={profile.id}>{profile.version}{profile.installed ? '' : '（未安装）'}</option>
			 {/each}
		   </select>
		 </div>
	   {/if}
     {/if}

     {#if !isSQLiteDatabase}
       <div class="grid grid-cols-3 gap-2">
         <div class="col-span-2">
           <label class={labelClass} for="connection-host">
             主机地址 <span class="ops-required">*</span>
           </label>
           <input
             type="text"
             id="connection-host"
             required
             bind:value={formData.host}
             placeholder="192.168.1.10 或 example.com"
             class={inputClass}
           />
         </div>
         <div>
           <label class={labelClass} for="connection-port">
             端口 <span class="ops-required">*</span>
           </label>
           <input
             type="text"
             id="connection-port"
             required
             bind:value={formData.port}
             placeholder={getDefaultPort()}
             class={inputClass}
           />
         </div>
       </div>
     {/if}

      {#if (authType === 'password' || assetType === 'database' || assetType === 'docker') && !isSQLiteDatabase}
        <div>
          <label class={labelClass} for="connection-username">
            用户名 <span class="ops-required">*</span>
          </label>
          <input
            type="text"
            id="connection-username"
            required
            bind:value={formData.username}
            placeholder="root"
            class={inputClass}
          />
        </div>
      {/if}

      {#if authType === 'password' && !isSQLiteDatabase}
        <div>
          <label class={labelClass} for="connection-password">
            密码 <span class="ops-required">*</span>
          </label>
          <div class="relative">
            {#if showPassword}
              <input
                type="text"
                id="connection-password"
                bind:value={formData.password}
                placeholder="输入密码"
                class={passwordInputClass}
              />
            {:else}
              <input
                type="password"
                id="connection-password"
                bind:value={formData.password}
                placeholder="输入密码"
                class={passwordInputClass}
              />
            {/if}
            <button
              type="button"
              on:click={() => showPassword = !showPassword}
              class="absolute right-2 top-1/2 transform -translate-y-1/2 p-1 ops-icon-inline rounded transition-colors"
            >
              {#if showPassword}
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              {:else}
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              {/if}
            </button>
          </div>
        </div>

        <div class="flex items-center gap-1.5 cursor-pointer">
          <input type="checkbox" bind:checked={formData.savePassword} class="w-3.5 h-3.5 rounded accent-control" />
          <span class="text-xs ops-field-label">保存密码</span>
        </div>
      {/if}

      {#if assetType === 'database'}
       <div>
         <label class={labelClass} for="database-name">
           {isSQLiteDatabase ? '数据库文件或 JDBC URL' : '数据库名'}
         </label>
         <input
           type="text"
           id="database-name"
           bind:value={formData.database}
           placeholder={isSQLiteDatabase ? '例如：/data/app.db 或 jdbc:sqlite:/data/app.db' : '例如：production_db'}
           class={inputClass}
         />
       </div>
      {/if}

      <div>
         <label class={labelClass} for="connection-group">
           分组
         </label>
         <div class="relative">
           <input
             type="text"
             id="connection-group"
             bind:value={groupSearchTerm}
             on:focus={toggleGroupDropdown}
             on:input={handleGroupSearchInput}
             on:keydown={handleGroupKeydown}
             on:blur={handleGroupBlur}
             placeholder="例如：生产环境"
             class={inputClass}
           />
           {#if showGroupDropdown && filteredGroups.length > 0}
             <div class="absolute z-50 mt-1 w-full ops-dropdown border rounded-md max-h-48 overflow-y-auto">
               {#each availableGroups as group, index (group)}
                 <button
                   type="button"
                   on:click={() => selectGroup(group)}
                   on:mouseover={() => selectedGroupIndex = index}
                   on:focus={() => selectedGroupIndex = index}
                   class={`w-full px-3 py-2 text-left text-xs cursor-pointer transition-colors ${
                     selectedGroupIndex === index
                       ? 'ops-dropdown-item-active'
                       : 'ops-dropdown-item'
                   }`}
                 >
                   {group}
                 </button>
               {/each}
               {#if availableGroups.length === 0}
                 <div class="px-3 py-2 text-xs ops-help">
                   没有匹配的分组
                 </div>
               {/if}
             </div>
           {/if}
         </div>
      </div>

     <div class="flex gap-2 pt-3">
       <button
         type="button"
         on:click={() => isOpen = false}
         class="px-3 py-1.5 ops-soft-button rounded-md text-xs font-medium transition-colors"
       >
         取消
       </button>
       <button
         type="button"
         on:click={handleTestConnection}
         disabled={testingConnection}
         class="px-3 py-1.5 ops-soft-button rounded-md text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
       >
         {testingConnection ? '测试中...' : '测试连接'}
       </button>
        <button
          type="submit"
          disabled={savingConnection}
          class="flex-1 px-3 py-1.5 accent-gradient rounded-md text-xs font-medium transition-all shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {savingConnection ? '保存中...' : (editingAsset ? '更新连接' : '添加连接')}
        </button>
     </div>
   </form>
</Dialog>
