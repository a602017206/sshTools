import { writable } from 'svelte/store';
import { GetSettings, UpdateSettings } from '../../wailsjs/go/main/App.js';

function createThemeStore() {
  const { subscribe, set, update } = writable({
    theme: 'dark',
    sidebarWidth: 300,
    loading: true
  });

  return {
    subscribe,

    async init() {
      try {
        const settings = await GetSettings();
        update(state => ({
          ...state,
          theme: settings.theme || 'dark',
          sidebarWidth: settings.sidebar_width || 300,
          loading: false
        }));

        document.documentElement.setAttribute('data-theme', settings.theme || 'dark');
      } catch (error) {
        console.error('Failed to load settings:', error);
        update(state => ({ ...state, loading: false }));
      }
    },

    async setTheme(newTheme) {
      try {
        await UpdateSettings({ theme: newTheme });
        update(state => ({ ...state, theme: newTheme }));
        document.documentElement.setAttribute('data-theme', newTheme);
      } catch (error) {
        console.error('Failed to save theme:', error);
        throw error;
      }
    },

    async setSidebarWidth(width) {
      try {
        const clampedWidth = Math.max(200, Math.min(600, width));
        await UpdateSettings({ sidebar_width: clampedWidth });
        update(state => ({ ...state, sidebarWidth: clampedWidth }));
      } catch (error) {
        console.error('Failed to save sidebar width:', error);
        throw error;
      }
    }
  };
}

export const themeStore = createThemeStore();
