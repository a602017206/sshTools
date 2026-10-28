import { afterEach, describe, it } from 'node:test';
import assert from 'node:assert/strict';

const originalDocument = globalThis.document;
const originalStorage = globalThis.localStorage;

afterEach(() => {
  globalThis.document = originalDocument;
  globalThis.localStorage = originalStorage;
});

function button(value) {
  const handlers = {};
  return {
    getAttribute: () => value,
    addEventListener: (event, handler) => { handlers[event] = handler; },
    click: () => handlers.click(),
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
    };
    globalThis.localStorage = {
      getItem: (key) => key === 'sshtools-site-theme' ? 'graphite' : 'zh',
      setItem: () => {},
    };
    globalThis.document = {
      body: { dataset: { page: 'features' } },
      documentElement: root,
      querySelectorAll: (selector) => ({
        '[data-set-theme]': [themeButton],
        '[data-set-lang]': [langButton],
        '[data-nav]': [nav],
        '[data-i18n]': [translated],
        '[data-i18n-html]': [],
      }[selector] || []),
    };
    const palettes = [];
    const { initChrome } = await import('../assets/js/nav.js');

    const result = initChrome({ sceneApi: { setPalette: (palette) => palettes.push(palette) } });
    themeButton.click();
    langButton.click();

    assert.equal(result.theme, 'graphite');
    assert.equal(result.lang, 'zh');
    assert.equal(root['data-theme'], 'amber');
    assert.equal(root.lang, 'en');
    assert.equal(nav['aria-current'], 'page');
    assert.equal(translated.textContent, 'Home');
    assert.equal(palettes.length, 2);
  });
});
