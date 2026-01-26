<script>
  import Dialog from './ui/Dialog.svelte';

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
    database: '',
  };

  async function handleTestConnection() {
    if (!window.wailsBindings) {
      testResult = 'Wails 绑定未加载';
      return;
    }

    if (!formData.host) {
      testResult = '请填写主机地址';
      return;
    }

    if (authType === 'password' && !formData.password) {
      testResult = '请输入密码以测试连接';
      return;
    }

    if (authType === 'password' && !formData.username) {
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
      const authValue = authType === 'key' ? formData.keyPath : formData.password;
      await window.wailsBindings.TestConnection(
        formData.host,
        parseInt(formData.port),
        formData.username,
        authType,
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

    if (!formData.name || !formData.host) {
      alert('请填写必填字段（连接名称、主机地址）');
      return;
    }

    if (authType === 'password' && !formData.username) {
      alert('密码认证需要填写用户名');
      return;
    }

    if (authType === 'password' && !formData.password) {
      alert('密码认证需要输入密码');
      return;
    }

    if (authType === 'key' && !formData.keyPath) {
      alert('密钥认证需要选择密钥文件');
      return;
    }

    savingConnection = true;

    try {
      const isEdit = !!formData.id;
      const connectionData = {
        id: isEdit ? formData.id : `conn_${Date.now()}`,
        name: formData.name,
        host: formData.host,
        port: parseInt(formData.port),
        user: formData.username,
        auth_type: authType,
        key_path: authType === 'key' ? formData.keyPath : '',
        tags: [formData.group || '默认分组'],
        type: assetType,
        metadata: {
          database: formData.database || undefined
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

      isOpen = false;
    } catch (error) {
      console.error('Failed to save connection:', error);
      alert('保存连接失败: ' + error);
    } finally {
      savingConnection = false;
    }
  }

  function resetForm() {
    formData = {
      id: '',
      name: '',
      host: '',
      port: getDefaultPort(),
      username: '',
      password: '',
      savePassword: false,
      keyPath: '',
      passphrase: '',
      group: '',
      dbType: 'mysql',
      database: '',
    };
    testResult = '';
    showPassword = false;
    showPassphrase = false;
  }

  $: if (!isOpen) {
    resetForm();
  }

  $: if (isOpen && !editingAsset) {
    // If opened without editingAsset, ensure form is reset
    resetForm();
    formData.port = getDefaultPort();
  }

  function getDefaultPort() {
    switch (assetType) {
      case 'ssh': return '22';
      case 'docker': return '2375';
      case 'database':
        switch (formData.dbType) {
          case 'mysql': return '3306';
          case 'postgresql': return '5432';
          case 'mongodb': return '27017';
          case 'redis': return '6379';
          default: return '';
        }
      default: return '';
    }
  }

  $: if (formData.port === '') {
    formData.port = getDefaultPort();
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
    if (!editingAsset || !isOpen) return;

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
          dbType: 'mysql',
          database: conn.metadata?.database || '',
        };
        assetType = conn.type || 'ssh';
        authType = conn.auth_type || 'password';

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
    <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">连接类型</label>
    <div class="grid grid-cols-3 gap-2">
      <button
        type="button"
        on:click={() => assetType = 'ssh'}
        class={`p-2 rounded-lg border-2 transition-all ${
          assetType === 'ssh'
            ? 'border-purple-600 bg-purple-50 dark:bg-purple-900/20'
            : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
        }`}
      >
        <div class="flex items-center justify-center gap-1">
          <svg class={`w-4 h-4 ${
            assetType === 'ssh' ? 'text-purple-600' : 'text-gray-600 dark:text-gray-400'
          }`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          <span class={`text-xs font-medium ${
            assetType === 'ssh' ? 'text-purple-900 dark:text-purple-200' : 'text-gray-700 dark:text-gray-300'
          }`}>SSH</span>
        </div>
      </button>

      <button
        type="button"
        on:click={() => assetType = 'database'}
        class={`p-2 rounded-lg border-2 transition-all ${
          assetType === 'database'
            ? 'border-purple-600 bg-purple-50 dark:bg-purple-900/20'
            : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
        }`}
      >
        <div class="flex items-center justify-center gap-1">
          <svg class={`w-4 h-4 ${
            assetType === 'database' ? 'text-purple-600' : 'text-gray-600 dark:text-gray-400'
          }`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
          </svg>
          <span class={`text-xs font-medium ${
            assetType === 'database' ? 'text-purple-900 dark:text-purple-200' : 'text-gray-700 dark:text-gray-300'
          }`}>数据库</span>
        </div>
      </button>

      <button
        type="button"
        on:click={() => assetType = 'docker'}
        class={`p-2 rounded-lg border-2 transition-all ${
          assetType === 'docker'
            ? 'border-purple-600 bg-purple-50 dark:bg-purple-900/20'
            : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
        }`}
      >
        <div class="flex items-center justify-center gap-1">
          <svg class={`w-4 h-4 ${
            assetType === 'docker' ? 'text-purple-600' : 'text-gray-600 dark:text-gray-400'
          }`} fill="currentColor" viewBox="0 0 24 24">
            <path d="M13.983 11.078h2.119a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.119a.185.185 0 00-.185.185v1.888c0 .102.083.185.185.185m-2.954-3.333h2.118a.186.186 0 00.186-.186V5.671a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.185m-2.954 3.333h2.118a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186H8.075a.186.186 0 00-.186.186v1.888c0 .102.083.185.186.185m-2.954-3.333h2.119a.186.186 0 00.185-.186V5.671a.185.185 0 00-.185-.185H5.12a.186.186 0 00-.186.185v1.888c0 .102.084.185.186.185m-2.93 3.333h2.12a.185.185 0 00.185-.185V9.006a.185.185 0 00-.186-.186h-2.12a.185.185 0 00-.184.186v1.888c0 .102.083.185.185.185M20.69 6.662c.057.16.09.331.09.51v7.9c0 3.058-2.724 4.928-8.78 4.928-6.055 0-8.779-1.87-8.779-4.928v-7.9c0-.179.033-.35.09-.51C1.536 7.396 0 9.522 0 12.072v3.639c0 4.072 3.608 6.789 12 6.789 8.391 0 12-2.717 12-6.79v-3.638c0-2.55-1.536-4.677-4.31-6.41" />
          </svg>
          <span class={`text-xs font-medium ${
            assetType === 'docker' ? 'text-purple-900 dark:text-purple-200' : 'text-gray-700 dark:text-gray-300'
          }`}>Docker</span>
        </div>
      </button>
    </div>
   </div>

   <form on:submit|preventDefault={handleSubmit} class="space-y-3">
      <div>
       <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
         连接名称 <span class="text-red-500">*</span>
       </label>
       <input
         type="text"
         required
         bind:value={formData.name}
         placeholder="例如：生产服务器-01"
         class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
       />
     </div>

     {#if assetType === 'ssh'}
       <div>
         <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
           认证方式
         </label>
         <div class="flex gap-2">
           <button
             type="button"
             on:click={() => { authType = 'password'; testResult = ''; }}
             class={`flex-1 px-3 py-1.5 rounded-md border-2 transition-all ${
               authType === 'password'
                 ? 'border-purple-600 bg-purple-50 dark:bg-purple-900/20'
                 : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
             }`}
           >
             <div class={`text-xs font-medium ${
               authType === 'password' ? 'text-purple-700 dark:text-purple-200' : 'text-gray-700 dark:text-gray-300'
             }`}>密码</div>
           </button>
           <button
             type="button"
             on:click={() => { authType = 'key'; testResult = ''; }}
             class={`flex-1 px-3 py-1.5 rounded-md border-2 transition-all ${
               authType === 'key'
                 ? 'border-purple-600 bg-purple-50 dark:bg-purple-900/20'
                 : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
             }`}
           >
             <div class={`text-xs font-medium ${
               authType === 'key' ? 'text-purple-700 dark:text-purple-200' : 'text-gray-700 dark:text-gray-300'
             }`}>SSH 密钥</div>
           </button>
         </div>
       </div>
     {/if}
      {#if authType === 'key'}
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            SSH 私钥文件
          </label>
          <div class="flex gap-1.5">
            <input
              type="text"
              bind:value={formData.keyPath}
              placeholder="点击选择密钥文件 (例如: ~/.ssh/id_rsa)"
              readonly
              class="flex-1 px-3 py-1.5 bg-gray-100 dark:bg-gray-600 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none transition-all"
            />
            <button
              type="button"
              on:click={handleSelectKeyFile}
              class="px-3 py-1.5 bg-purple-600 hover:bg-purple-700 text-white rounded-md text-xs font-medium transition-colors whitespace-nowrap"
            >
              选择文件
            </button>
          </div>
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            Passphrase（可选）
          </label>
          <div class="relative">
            {#if showPassphrase}
              <input
                type="text"
                bind:value={formData.passphrase}
                placeholder="如果密钥已加密，请输入 passphrase"
                class="w-full px-3 py-1.5 pr-10 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            {:else}
              <input
                type="password"
                bind:value={formData.passphrase}
                placeholder="如果密钥已加密，请输入 passphrase"
                class="w-full px-3 py-1.5 pr-10 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            {/if}
            <button
              type="button"
              on:click={() => showPassphrase = !showPassphrase}
              class="absolute right-2 top-1/2 transform -translate-y-1/2 p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
            >
              {#if showPassphrase}
                <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              {:else}
                <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              {/if}
            </button>
          </div>
          <p class="mt-0.5 text-[10px] text-gray-500 dark:text-gray-400">
            如果您的 SSH 密钥文件已加密，请输入 passphrase。否则留空即可。
          </p>
        </div>
      {/if}

     {#if testResult}
       <div class={`p-2 rounded-md text-xs ${
         testResult.includes('成功') ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300' : 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300'
       }`}>
         {testResult}
       </div>
      {/if}

     {#if assetType === 'database'}
       <div>
         <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
           数据库类型 <span class="text-red-500">*</span>
         </label>
         <select
           bind:value={formData.dbType}
           class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
         >
           <option value="mysql">MySQL</option>
           <option value="postgresql">PostgreSQL</option>
           <option value="mongodb">MongoDB</option>
           <option value="redis">Redis</option>
         </select>
       </div>
     {/if}

     <div class="grid grid-cols-3 gap-2">
       <div class="col-span-2">
         <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
           主机地址 <span class="text-red-500">*</span>
         </label>
         <input
           type="text"
           required
           bind:value={formData.host}
           placeholder="192.168.1.10 或 example.com"
           class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
         />
       </div>
       <div>
         <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
           端口 <span class="text-red-500">*</span>
         </label>
         <input
           type="text"
           required
           bind:value={formData.port}
           placeholder={getDefaultPort()}
           class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
         />
       </div>
     </div>

      {#if authType === 'password' || assetType === 'database' || assetType === 'docker'}
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            用户名 <span class="text-red-500">*</span>
          </label>
          <input
            type="text"
            required
            bind:value={formData.username}
            placeholder="root"
            class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
          />
        </div>
      {/if}

      {#if authType === 'password'}
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            密码 <span class="text-red-500">*</span>
          </label>
          <div class="relative">
            {#if showPassword}
              <input
                type="text"
                bind:value={formData.password}
                placeholder="输入密码"
                class="w-full px-3 py-1.5 pr-10 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            {:else}
              <input
                type="password"
                bind:value={formData.password}
                placeholder="输入密码"
                class="w-full px-3 py-1.5 pr-10 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            {/if}
            <button
              type="button"
              on:click={() => showPassword = !showPassword}
              class="absolute right-2 top-1/2 transform -translate-y-1/2 p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
            >
              {#if showPassword}
                <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              {:else}
                <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              {/if}
            </button>
          </div>
        </div>

        <div class="flex items-center gap-1.5 cursor-pointer">
          <input type="checkbox" bind:checked={formData.savePassword} class="w-3.5 h-3.5 rounded border-gray-300 text-purple-600 focus:ring-purple-500" />
          <span class="text-xs text-gray-700 dark:text-gray-300">保存密码</span>
        </div>
      {/if}

      {#if assetType === 'database'}
       <div>
         <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
           数据库名
         </label>
         <input
           type="text"
           bind:value={formData.database}
           placeholder="例如：production_db"
           class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
         />
       </div>
      {/if}

     <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          分组
        </label>
        <input
          type="text"
          bind:value={formData.group}
          placeholder="例如：生产环境"
          class="w-full px-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-xs text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
        />
     </div>

     <div class="flex gap-2 pt-3">
       <button
         type="button"
         on:click={() => isOpen = false}
         class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-md text-xs font-medium transition-colors"
       >
         取消
       </button>
       <button
         type="button"
         on:click={handleTestConnection}
         disabled={testingConnection}
         class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-md text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
       >
         {testingConnection ? '测试中...' : '测试连接'}
       </button>
        <button
          type="submit"
          disabled={savingConnection}
          class="flex-1 px-3 py-1.5 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white rounded-md text-xs font-medium transition-all shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {savingConnection ? '保存中...' : (editingAsset ? '更新连接' : '添加连接')}
        </button>
     </div>
   </form>
</Dialog>
