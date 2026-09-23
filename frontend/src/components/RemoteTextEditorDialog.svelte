<script>
  import { onDestroy, onMount, tick } from 'svelte';
  import { EditorState, Compartment } from '@codemirror/state';
  import { EditorView, keymap } from '@codemirror/view';
  import { basicSetup } from 'codemirror';
  import { indentWithTab, redo, redoDepth, undo, undoDepth } from '@codemirror/commands';
  import { StreamLanguage } from '@codemirror/language';
  import { json } from '@codemirror/lang-json';
  import { yaml } from '@codemirror/lang-yaml';
  import { xml } from '@codemirror/lang-xml';
  import { html } from '@codemirror/lang-html';
  import { css } from '@codemirror/lang-css';
  import { javascript } from '@codemirror/lang-javascript';
  import { markdown } from '@codemirror/lang-markdown';
  import { sql } from '@codemirror/lang-sql';
  import { python } from '@codemirror/lang-python';
  import { shell } from '@codemirror/legacy-modes/mode/shell';
  import { properties } from '@codemirror/legacy-modes/mode/properties';
  import { toml } from '@codemirror/legacy-modes/mode/toml';
  import { nginx } from '@codemirror/legacy-modes/mode/nginx';
  import { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
  import { oneDark } from '@codemirror/theme-one-dark';
  import Dialog from './ui/Dialog.svelte';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import { ReadRemoteTextFile, SaveRemoteTextFile } from '../../wailsjs/go/main/App.js';
  import { formatRemoteText, resolveTextLanguage } from '../lib/remoteTextFormat.js';
  import { isMacPlatform } from '../lib/fileManagerContextMenu.js';

  export let sessionId = '';
  export let file = null;
  export let onClose = () => {};
  export let onSaved = () => {};

  const isMac = typeof navigator !== 'undefined' && isMacPlatform(navigator.userAgent || navigator.platform);
  const saveShortcut = isMac ? '⌘S' : 'Ctrl+S';
  const languageCompartment = new Compartment();

  let host;
  let view;
  let loading = true;
  let loadError = '';
  let status = '';
  let statusError = false;
  let saving = false;
  let dirty = false;
  let canUndo = false;
  let canRedo = false;
  let languageLabel = '文本';
  let confirmDiscard = false;
  let savedText = '';
  let newline = '\n';

  function errorText(err) {
    if (!err) return '操作失败';
    if (typeof err === 'string') return err;
    return err.message || String(err);
  }

  function detectNewline(text) {
    const crlf = (String(text || '').match(/\r\n/g) || []).length;
    const lf = (String(text || '').match(/(?<!\r)\n/g) || []).length;
    return crlf > lf ? '\r\n' : '\n';
  }

  function applyNewline(text, ending) {
    const normalized = String(text ?? '').replace(/\r\n/g, '\n').replace(/\r/g, '\n');
    if (ending === '\r\n') return normalized.replace(/\n/g, '\r\n');
    return normalized;
  }

  function languageExtension(id) {
    switch (id) {
      case 'json':
      case 'jsonc':
        return json();
      case 'yaml':
        return yaml();
      case 'xml':
        return xml();
      case 'html':
        return html();
      case 'css':
        return css();
      case 'javascript':
        return javascript();
      case 'markdown':
        return markdown();
      case 'sql':
        return sql();
      case 'python':
        return python();
      case 'shell':
        return StreamLanguage.define(shell);
      case 'properties':
        return StreamLanguage.define(properties);
      case 'toml':
        return StreamLanguage.define(toml);
      case 'nginx':
        return StreamLanguage.define(nginx);
      case 'dockerfile':
        return StreamLanguage.define(dockerFile);
      default:
        return [];
    }
  }

  function syncFromView() {
    if (!view) return;
    dirty = view.state.doc.toString() !== savedText;
    canUndo = undoDepth(view.state) > 0;
    canRedo = redoDepth(view.state) > 0;
  }

  function createEditor(content) {
    const dark = typeof document !== 'undefined' && document.documentElement.classList.contains('dark');
    const language = resolveTextLanguage(file?.name, content);
    languageLabel = language.label;
    view = new EditorView({
      parent: host,
      state: EditorState.create({
        doc: content,
        extensions: [
          basicSetup,
          keymap.of([indentWithTab]),
          languageCompartment.of(languageExtension(language.id)),
          dark ? oneDark : [],
          EditorView.theme({
            '&': {
              height: '100%',
              backgroundColor: 'var(--bg-input)',
              color: 'var(--text-primary)',
              fontSize: '13px',
            },
            '.cm-scroller': { fontFamily: 'var(--terminal-font-family)' },
            '.cm-content': { caretColor: 'var(--text-primary)' },
            '.cm-gutters': {
              backgroundColor: 'var(--bg-tertiary)',
              color: 'var(--text-tertiary)',
              borderRight: '1px solid var(--border-primary)',
            },
            '.cm-activeLine': { backgroundColor: 'var(--bg-hover)' },
            '&.cm-focused': { outline: 'none' },
          }, { dark }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged || update.transactions.length) syncFromView();
          }),
        ],
      }),
    });
    savedText = view.state.doc.toString();
    syncFromView();
    view.focus();
  }

  function handleWindowKeydown(event) {
    const meta = isMac ? event.metaKey : event.ctrlKey;
    if (!meta || event.altKey || event.shiftKey) return;
    if (String(event.key).toLowerCase() !== 's') return;
    event.preventDefault();
    save();
  }

  onMount(async () => {
    window.addEventListener('keydown', handleWindowKeydown);
    try {
      const result = await ReadRemoteTextFile(sessionId, file.path);
      const content = result?.content || '';
      newline = detectNewline(content);
      loading = false;
      await tick();
      createEditor(content);
    } catch (err) {
      loading = false;
      loadError = errorText(err);
    }
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleWindowKeydown);
    view?.destroy();
  });

  function requestClose() {
    if (confirmDiscard) return;
    if (dirty) {
      confirmDiscard = true;
      return;
    }
    onClose();
  }

  function formatDocument() {
    if (!view || saving) return;
    const current = view.state.doc.toString();
    const result = formatRemoteText(file?.name, current);
    if (!result.ok) {
      statusError = true;
      status = result.error;
      return;
    }
    languageLabel = result.label || languageLabel;
    if (!result.changed) {
      view.dispatch({
        effects: languageCompartment.reconfigure(languageExtension(result.language)),
      });
      statusError = false;
      status = '内容已经是格式化结果';
      return;
    }
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: result.text },
      effects: languageCompartment.reconfigure(languageExtension(result.language)),
    });
    statusError = false;
    status = '已格式化，可撤销';
  }

  function undoEdit() {
    if (view) undo(view);
  }

  function redoEdit() {
    if (view) redo(view);
  }

  async function save() {
    if (!view || saving || loadError) return;
    const current = view.state.doc.toString();
    if (current === savedText) {
      statusError = false;
      status = '没有需要保存的修改';
      return;
    }
    saving = true;
    statusError = false;
    status = '正在保存…';
    try {
      await SaveRemoteTextFile(sessionId, file.path, applyNewline(current, newline));
      savedText = current;
      dirty = false;
      statusError = false;
      status = '已保存到服务器';
      onSaved();
    } catch (err) {
      statusError = true;
      status = errorText(err);
    } finally {
      saving = false;
    }
  }
</script>

<Dialog
  isOpen={true}
  onClose={requestClose}
  title={`编辑 ${file?.name || '文件'}`}
  size="xl"
>
  <div class="remote-text-editor">
    {#if loading}
      <p class="remote-text-editor__status">正在读取文件…</p>
    {:else if loadError}
      <p class="remote-text-editor__status is-error">{loadError}</p>
    {:else}
      <div class="remote-text-editor__toolbar">
        <div class="remote-text-editor__meta">
          <span class="remote-text-editor__lang">{languageLabel}</span>
          <span class="remote-text-editor__path" title={file?.path || ''}>{file?.path || ''}</span>
          <span class="remote-text-editor__dirty">{dirty ? '未保存' : '已保存'}</span>
        </div>
        <div class="remote-text-editor__actions">
          <button type="button" class="ops-btn-glass" on:click={formatDocument} disabled={saving}>格式化</button>
          <button type="button" class="ops-btn-glass" on:click={undoEdit} disabled={!canUndo || saving}>撤销</button>
          <button type="button" class="ops-btn-glass" on:click={redoEdit} disabled={!canRedo || saving}>重做</button>
          <button type="button" class="remote-text-editor__save" on:click={save} disabled={saving || !dirty}>
            {saving ? '保存中…' : `保存 ${saveShortcut}`}
          </button>
        </div>
      </div>
      <div class="remote-text-editor__host" bind:this={host}></div>
      {#if status}
        <p class="remote-text-editor__status" class:is-error={statusError}>{status}</p>
      {/if}
    {/if}
  </div>
</Dialog>

<ConfirmDialog
  bind:isOpen={confirmDiscard}
  title="放弃修改"
  message="文件有未保存的修改，关闭后不会写回服务器。"
  type="warning"
  confirmText="放弃"
  cancelText="继续编辑"
  onConfirm={() => { confirmDiscard = false; onClose(); }}
  onCancel={() => { confirmDiscard = false; }}
/>

<style>
  .remote-text-editor {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    height: min(72vh, 760px);
    min-height: 360px;
  }

  .remote-text-editor__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .remote-text-editor__meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    flex: 1;
  }

  .remote-text-editor__lang,
  .remote-text-editor__dirty {
    flex-shrink: 0;
    font-size: 12px;
    line-height: 1.4;
    padding: 0.15rem 0.45rem;
    border-radius: 999px;
    background: var(--accent-subtle);
    color: var(--accent-primary);
  }

  .remote-text-editor__path {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .remote-text-editor__actions {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
  }

  .remote-text-editor__actions button {
    padding: 0.35rem 0.7rem;
    border-radius: 0.6rem;
    font-size: 12px;
  }

  .remote-text-editor__actions button:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .remote-text-editor__save {
    background: var(--accent-primary);
    color: white;
  }

  .remote-text-editor__save:hover:not(:disabled) {
    background: var(--accent-hover);
  }

  .remote-text-editor__host {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    border: 1px solid var(--border-primary);
    border-radius: 12px;
    background: var(--bg-input);
  }

  :global(.remote-text-editor__host .cm-editor) {
    height: 100%;
  }

  .remote-text-editor__status {
    margin: 0;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .remote-text-editor__status.is-error {
    color: var(--ops-alert);
  }
</style>
