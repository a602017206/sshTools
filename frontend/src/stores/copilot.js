import { writable } from 'svelte/store';

const initialState = {
  open: false,
  width: 360,
  messagesBySession: {}
};

function createCopilotStore() {
  const { subscribe, update, set } = writable(initialState);

  return {
    subscribe,
    set,
    toggle() {
      update((state) => ({ ...state, open: !state.open }));
    },
    setOpen(open) {
      update((state) => ({ ...state, open: Boolean(open) }));
    },
    setWidth(width) {
      const next = Math.max(280, Math.min(520, Number(width) || 360));
      update((state) => ({ ...state, width: next }));
    },
    appendMessage(sessionId, message) {
      if (!sessionId || !message) return;
      update((state) => {
        const current = state.messagesBySession[sessionId] || [];
        return {
          ...state,
          messagesBySession: {
            ...state.messagesBySession,
            [sessionId]: [...current, message]
          }
        };
      });
    },
    setMessages(sessionId, messages) {
      if (!sessionId) return;
      update((state) => ({
        ...state,
        messagesBySession: {
          ...state.messagesBySession,
          [sessionId]: Array.isArray(messages) ? messages : []
        }
      }));
    },
    clearSession(sessionId) {
      if (!sessionId) return;
      update((state) => {
        if (!state.messagesBySession[sessionId]) return state;
        const next = { ...state.messagesBySession };
        delete next[sessionId];
        return { ...state, messagesBySession: next };
      });
    }
  };
}

export const copilotStore = createCopilotStore();
