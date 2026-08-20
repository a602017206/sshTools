import { writable } from 'svelte/store';
import { appendTerminalTail } from '../lib/copilotContext.js';

const initialState = {
  open: false,
  width: 360,
  messagesBySession: {},
  terminalTailsBySession: {}
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
    appendTerminalOutput(sessionId, output) {
      if (!sessionId || !output) return;
      update((state) => ({
        ...state,
        terminalTailsBySession: appendTerminalTail(state.terminalTailsBySession, sessionId, output)
      }));
    },
    clearSession(sessionId) {
      if (!sessionId) return;
      update((state) => {
        const messagesBySession = { ...state.messagesBySession };
        const terminalTailsBySession = { ...state.terminalTailsBySession };
        delete messagesBySession[sessionId];
        delete terminalTailsBySession[sessionId];
        return { ...state, messagesBySession, terminalTailsBySession };
      });
    }
  };
}

export const copilotStore = createCopilotStore();
