import { writable } from 'svelte/store';
import { GetSettings, UpdateSettings } from '../../wailsjs/go/main/App.js';

function createFileManagerStore() {
  const { subscribe, set, update } = writable({
    collapsed: true,
    width: 400,
    loading: true,

    // Per-session state
    sessionStates: {},  // sessionID -> { currentPath, selectedFiles, files, sortBy, sortOrder }

    // Active transfers
    transfers: {},  // transferID -> TransferProgress

    // View preferences
    showHiddenFiles: false,
    sortBy: 'name',
    sortOrder: 'asc'
  });

  return {
    subscribe,

    async init() {
      try {
        const settings = await GetSettings();
        update(state => ({
          ...state,
          collapsed: settings.file_manager_collapsed ?? true,
          width: settings.file_manager_width || 400,
          showHiddenFiles: settings.file_manager_show_hidden || false,
          sortBy: settings.file_manager_sort_by || 'name',
          sortOrder: settings.file_manager_sort_order || 'asc',
          loading: false
        }));
      } catch (error) {
        console.error('Failed to load file manager settings:', error);
        update(state => ({ ...state, loading: false }));
      }
    },

    async setCollapsed(collapsed) {
      try {
        await UpdateSettings({ file_manager_collapsed: collapsed });
        update(state => ({ ...state, collapsed }));
      } catch (error) {
        console.error('Failed to save collapsed state:', error);
      }
    },

    async setWidth(width) {
      try {
        const clampedWidth = Math.max(350, Math.min(800, width));
        await UpdateSettings({ file_manager_width: clampedWidth });
        update(state => ({ ...state, width: clampedWidth }));
      } catch (error) {
        console.error('Failed to save width:', error);
      }
    },

    // Session-specific state management
    setSessionState(sessionID, stateUpdates) {
      update(state => ({
        ...state,
        sessionStates: {
          ...state.sessionStates,
          [sessionID]: {
            ...(state.sessionStates[sessionID] || {}),
            ...stateUpdates
          }
        }
      }));
    },

    getSessionState(sessionID) {
      let currentState = null;
      update(state => {
        currentState = state.sessionStates[sessionID] || null;
        return state;
      });
      return currentState;
    },

    clearSessionState(sessionID) {
      update(state => {
        const newStates = { ...state.sessionStates };
        delete newStates[sessionID];
        return { ...state, sessionStates: newStates };
      });
    },

    // Transfer management
    updateTransfer(transferID, progress) {
      update(state => ({
        ...state,
        transfers: {
          ...state.transfers,
          [transferID]: progress
        }
      }));
    },

    removeTransfer(transferID) {
      update(state => {
        const newTransfers = { ...state.transfers };
        delete newTransfers[transferID];
        return { ...state, transfers: newTransfers };
      });
    },

    // View preferences
    async setSortPreferences(sortBy, sortOrder) {
      try {
        await UpdateSettings({
          file_manager_sort_by: sortBy,
          file_manager_sort_order: sortOrder
        });
        update(state => ({ ...state, sortBy, sortOrder }));
      } catch (error) {
        console.error('Failed to save sort preferences:', error);
      }
    },

    async toggleShowHidden() {
      update(state => {
        const newValue = !state.showHiddenFiles;
        UpdateSettings({ file_manager_show_hidden: newValue }).catch(err => {
          console.error('Failed to save show hidden preference:', err);
        });
        return { ...state, showHiddenFiles: newValue };
      });
    }
  };
}

export const fileManagerStore = createFileManagerStore();
