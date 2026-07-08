import { writable, derived } from 'svelte/store';

const initialState = {
  transfers: [],
  maxHistory: 50,
  isPanelOpen: false,
  activeTab: 'active',
};

const { subscribe, set, update } = writable(initialState);

const storeInstance = {
  subscribe,
  set,

  // Add a new transfer
  addTransfer: (transfer) =>
    update((state) => {
      const completed = state.transfers.filter(t => t.status !== 'running');
      let transfers = state.transfers;

      if (completed.length >= state.maxHistory) {
        const toRemove = completed.length - state.maxHistory + 1;
        const completedIds = completed
          .slice(0, toRemove)
          .map(t => t.id);
        transfers = state.transfers.filter(t => !completedIds.includes(t.id));
      }

      return {
        ...state,
        transfers: [transfer, ...transfers],
      };
    }),

  // Update an existing transfer
  updateTransfer: (id, progress) =>
    update((state) => ({
      ...state,
      transfers: state.transfers.map((t) =>
        t.id === id ? { ...t, ...progress } : t
      ),
    })),

  // Cancel a running transfer
  cancelTransfer: (id) =>
    update((state) => ({
      ...state,
      transfers: state.transfers.map((t) =>
        t.id === id ? { ...t, status: 'cancelled' } : t
      ),
    })),

  // Remove a transfer from history
  removeTransfer: (id) =>
    update((state) => ({
      ...state,
      transfers: state.transfers.filter((t) => t.id !== id),
    })),

  // Clear all completed transfers
  clearCompleted: () =>
    update((state) => ({
      ...state,
      transfers: state.transfers.filter(t => t.status === 'running'),
    })),

  // Clear all transfers
  clearAll: () =>
    update((state) => ({
      ...state,
      transfers: [],
    })),

  // Toggle panel open/closed
  togglePanel: () =>
    update((state) => ({
      ...state,
      isPanelOpen: !state.isPanelOpen,
    })),

  // Set panel open state
  setPanelOpen: (isOpen) =>
    update((state) => ({
      ...state,
      isPanelOpen: isOpen,
    })),

  // Set active tab
  setActiveTab: (tab) =>
    update((state) => ({
      ...state,
      activeTab: tab,
    })),
};

// Helper function to format file size
export function formatFileSize(size) {
  if (!size) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024;
    i++;
  }
  return `${size.toFixed(1)} ${units[i]}`;
}

// Helper function to format speed
export function formatSpeed(speed) {
  if (!speed) return '0 B/s';
  return `${formatFileSize(speed)}/s`;
}

// Helper function to get transfer percentage
export function getTransferPercentage(transfer) {
  if (!transfer) return 0;
  if (transfer.totalBytes) {
    return (transfer.bytesSent / transfer.totalBytes) * 100;
  }
  return transfer.percentage || 0;
}

// Derived stores
export const activeTransfers = derived(
  storeInstance,
  ($store) => ($store?.transfers || []).filter(t => t.status === 'running')
);

export const completedTransfers = derived(
  storeInstance,
  ($store) => ($store?.transfers || []).filter(t => t.status !== 'running')
);

export const overallProgress = derived(
  storeInstance,
  ($store) => {
    const running = ($store?.transfers || []).filter(t => t.status === 'running');
    if (running.length === 0) return 0;

    let totalBytes = 0;
    let sentBytes = 0;
    running.forEach((transfer) => {
      const total = transfer.totalBytes || 0;
      const sent = transfer.bytesSent || 0;
      totalBytes += total;
      sentBytes += sent;
    });

    if (totalBytes > 0) {
      return (sentBytes / totalBytes) * 100;
    }

    const sumPercentage = running.reduce(
      (sum, transfer) => {
        if (transfer.totalBytes) {
          return sum + (transfer.bytesSent / transfer.totalBytes) * 100;
        }
        return sum + (transfer.percentage || 0);
      },
      0
    );
    return sumPercentage / running.length;
  }
);

// Export the store instance
export const uploadStore = storeInstance;
