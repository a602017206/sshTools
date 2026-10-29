import { afterEach, describe, it } from 'node:test';
import assert from 'node:assert/strict';

const originalDocument = globalThis.document;
const originalStorage = globalThis.localStorage;

afterEach(() => {
  globalThis.document = originalDocument;
  globalThis.localStorage = originalStorage;
});

function button(attrValue) {
  const handlers = {};
  const attrs = {};
  return {
    getAttribute: (name) => (name ? attrValue : attrValue),
    setAttribute: (key, value) => { attrs[key] = value; },
    classList: { toggle: () => {} },
    addEventListener: (event, handler) => { handlers[event] = handler; },
    click: () => handlers.click?.(),
    attrs,
  };
}

describe('initChrome', () => {
  it('applies saved preferences, marks the current page, and updates the scene palette', async () => {
    const themeButton = button('amber');
    const langButton = button('en');
    const nav = {
      getAttribute: () => 'features',
      setAttribute: (key, value) => { nav[key] = value; },
    };
    const translated = { getAttribute: () => 'nav.home', textContent: '' };
    const root = {
      setAttribute: (key, value) => { root[key] = value; },
      style: { setProperty: () => {} },
    };
    const listeners = [];
    globalThis.localStorage = {
      getItem: (key) => {
        if (key === 'sshtools-site-theme') return 'graphite';
        if (key === 'sshtools-site-lang') return 'zh';
        return null;
      },
      setItem: () => {},
    };
    globalThis.document = {
      body: { dataset: { page: 'features' } },
      documentElement: root,
      addEventListener: (...args) => { listeners.push(args); },
      querySelectorAll: (selector) => ({
        '.theme-dots': [],
        '[data-set-theme]': [themeButton],
        '[data-set-lang]': [langButton],
        '[data-set-bg]': [],
        '[data-set-fg]': [],
        '[data-set-accent]': [],
        '[data-set-cursor]': [],
        '[data-theme-mode]': [],
        '[data-theme-pane]': [],
        '[data-theme-toggle]': [],
        '.theme-panel__pop': [],
        '[data-nav]': [nav],
        '[data-i18n]': [translated],
        '[data-i18n-html]': [],
        '[data-i18n="theme.presets"]': [],
        '[data-i18n="theme.bg"]': [],
        '[data-i18n="theme.fg"]': [],
        '[data-i18n="theme.accent"]': [],
        '[data-i18n="theme.cursor"]': [],
      }[selector] || []),
    };
    const palettes = [];
    const { initChrome } = await import('../assets/js/nav.js?theme-panel');

    const result = initChrome({ sceneApi: { setPalette: (palette) => palettes.push(palette) } });
    themeButton.click();
    langButton.click();

    assert.equal(result.theme, 'graphite');
    assert.equal(result.appearance.mode, 'preset');
    assert.equal(result.lang, 'zh');
    assert.equal(root['data-theme'], 'amber');
    assert.equal(root['data-theme-mode'], 'preset');
    assert.equal(root.lang, 'en');
    assert.equal(nav['aria-current'], 'page');
    assert.equal(translated.textContent, 'Home');
    assert.equal(palettes.length, 2);
    assert.equal(listeners.length, 1);
  });
});
