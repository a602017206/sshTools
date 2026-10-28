import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  LANG_STORAGE_KEY,
  MESSAGES,
  applyI18n,
  normalizeLang,
  readStoredLang,
  setStoredLang,
  t,
} from '../assets/js/i18n.js';

describe('normalizeLang', () => {
  it('accepts zh and en', () => {
    assert.equal(normalizeLang('en'), 'en');
    assert.equal(normalizeLang('zh'), 'zh');
  });

  it('defaults to zh', () => {
    assert.equal(normalizeLang('fr'), 'zh');
  });
});

describe('t', () => {
  it('returns zh string', () => {
    assert.equal(t(MESSAGES, 'zh', 'brand'), 'AHaSSHTools');
  });

  it('returns en tagline', () => {
    assert.match(t(MESSAGES, 'en', 'tagline'), /ops|SSH|client/i);
  });

  it('falls back to key', () => {
    assert.equal(t(MESSAGES, 'zh', 'missing.key'), 'missing.key');
  });
});

describe('language storage', () => {
  it('reads and normalizes the saved language', () => {
    assert.equal(readStoredLang({ getItem: () => 'en' }), 'en');
    assert.equal(readStoredLang({ getItem: () => 'fr' }), 'zh');
  });

  it('normalizes and persists the selected language', () => {
    const values = {};
    const storage = { setItem: (key, value) => { values[key] = value; } };

    assert.equal(setStoredLang('en', storage), 'en');
    assert.equal(values[LANG_STORAGE_KEY], 'en');
    assert.equal(setStoredLang('fr', storage), 'zh');
    assert.equal(values[LANG_STORAGE_KEY], 'zh');
  });
});

describe('applyI18n', () => {
  it('fills text and HTML translations, then sets the document language', () => {
    const el = {
      getAttribute: () => 'nav.home',
      textContent: '',
    };
    const htmlEl = {
      getAttribute: () => 'download.noteMac',
      innerHTML: '',
    };
    const documentElement = {
      setAttribute: (key, value) => { documentElement[key] = value; },
    };
    const root = {
      documentElement,
      querySelectorAll: (selector) => (selector.includes('data-i18n-html') ? [htmlEl] : [el]),
    };

    applyI18n(root, MESSAGES, 'zh');

    assert.ok(el.textContent.length > 0);
    assert.ok(htmlEl.innerHTML.length > 0);
    assert.equal(documentElement.lang, 'zh-CN');
  });
});
