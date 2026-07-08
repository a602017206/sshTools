<script>
  import { onMount } from 'svelte';
  import Dialog from './ui/Dialog.svelte';

  export let isOpen = false;
  export let onClose = () => {};
  export let themeStore = { subscribe: () => () => {} };

  let theme = 'light';
  let version = 'loading...';

  const unsubscribe = themeStore.subscribe(t => {
    theme = t;
  });

  onMount(async () => {
    try {
      const { GetVersion } = await import('../../wailsjs/go/main/App.js');
      version = await GetVersion();
    } catch (error) {
      console.error('Failed to get version:', error);
      version = 'unknown';
    }

    return () => {
      unsubscribe();
    };
  });

  async function openGitHub() {
    try {
      const runtime = await import('../../wailsjs/runtime/runtime.js');
      await runtime.BrowserOpenURL('https://github.com/a602017206/sshTools');
    } catch (error) {
      console.error('Failed to open GitHub:', error);
      window.open('https://github.com/a602017206/sshTools', '_blank');
    }
  }
</script>

<Dialog bind:isOpen={isOpen} onClose={onClose} title="关于 AHaSSHTools" size="sm">
  <div class="flex flex-col items-center gap-5 py-4">
    <!-- Logo -->
    <div class="w-20 h-20 bg-gradient-to-br from-purple-600 to-blue-600 rounded-2xl flex items-center justify-center font-bold text-3xl text-white shadow-lg">
      哈
    </div>

    <!-- App Info -->
    <div class="text-center space-y-1">
      <h3 class="text-xl font-bold {theme === 'dark' ? 'text-white' : 'text-gray-900'}">AHaSSHTools</h3>
      <p class="text-sm {theme === 'dark' ? 'text-gray-400' : 'text-gray-500'}">啊哈 SSH 连接工具</p>
      <p class="text-xs {theme === 'dark' ? 'text-gray-500' : 'text-gray-400'} mt-2">版本 {version}</p>
    </div>

    <!-- Description -->
    <p class="text-sm text-center max-w-xs {theme === 'dark' ? 'text-gray-400' : 'text-gray-600'}">
      一个功能完整的跨平台SSH桌面客户端工具，集成终端、文件管理、系统监控于一体。
    </p>

    <!-- GitHub Link -->
    <button
      on:click={openGitHub}
      class="flex items-center gap-2 px-5 py-2.5 rounded-lg font-medium transition-all shadow-sm {theme === 'dark' ? 'bg-gray-700 hover:bg-gray-600 text-white' : 'bg-gray-100 hover:bg-gray-200 text-gray-700'}"
    >
      <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
      </svg>
      <span>访问 GitHub</span>
    </button>

    <!-- Copyright -->
    <p class="text-xs {theme === 'dark' ? 'text-gray-500' : 'text-gray-400'} pt-2">
      © 2026 AHaSSHTools. All rights reserved.
    </p>
  </div>
</Dialog>
