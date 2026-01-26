import { writable } from 'svelte/store';
import { GetSettings, UpdateSettings } from '../../wailsjs/go/main/App.js';

function createMonitorStore() {
  const { subscribe, set, update } = writable({
    collapsed: true,
    width: 350,
    refreshInterval: 2, // seconds
    loading: true
  });

  return {
    subscribe,

    async init() {
      try {
        const settings = await GetSettings();
        update(state => ({
          ...state,
          collapsed: settings.monitor_collapsed ?? true,
          width: settings.monitor_width || 350,
          refreshInterval: settings.monitor_refresh_interval || 2,
          loading: false
        }));
      } catch (error) {
        console.error('Failed to load monitor settings:', error);
        update(state => ({ ...state, loading: false }));
      }
    },

    async setCollapsed(collapsed) {
      try {
        await UpdateSettings({ monitor_collapsed: collapsed });
        update(state => ({ ...state, collapsed }));
      } catch (error) {
        console.error('Failed to save collapsed state:', error);
      }
    },

    async setWidth(width) {
      try {
        const clampedWidth = Math.max(300, Math.min(600, width));
        await UpdateSettings({ monitor_width: clampedWidth });
        update(state => ({ ...state, width: clampedWidth }));
      } catch (error) {
        console.error('Failed to save width:', error);
      }
    },

    async setRefreshInterval(interval) {
      try {
        const clampedInterval = Math.max(1, Math.min(10, interval));
        await UpdateSettings({ monitor_refresh_interval: clampedInterval });
        update(state => ({ ...state, refreshInterval: clampedInterval }));
      } catch (error) {
        console.error('Failed to save refresh interval:', error);
      }
    }
  };
}

export const monitorStore = createMonitorStore();
