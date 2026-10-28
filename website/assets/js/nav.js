import { applyTheme, readStoredTheme, THEME_IDS } from './theme.js';
import { applyI18n, readStoredLang, setStoredLang, MESSAGES } from './i18n.js';

export function initChrome({ sceneApi } = {}) {
  const theme = applyTheme(readStoredTheme(), {
    onPalette: (palette) => sceneApi?.setPalette?.(palette),
  });
  let lang = readStoredLang();
  applyI18n(document, MESSAGES, lang);
  document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';

  document.querySelectorAll('[data-set-theme]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const id = btn.getAttribute('data-set-theme');
      applyTheme(id, { onPalette: (palette) => sceneApi?.setPalette?.(palette) });
    });
  });

  document.querySelectorAll('[data-set-lang]').forEach((btn) => {
    btn.addEventListener('click', () => {
      lang = setStoredLang(btn.getAttribute('data-set-lang'));
      applyI18n(document, MESSAGES, lang);
      document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';
    });
  });

  const page = document.body.dataset.page;
  document.querySelectorAll('[data-nav]').forEach((link) => {
    if (link.getAttribute('data-nav') === page) link.setAttribute('aria-current', 'page');
  });

  return { theme, lang, THEME_IDS };
}
