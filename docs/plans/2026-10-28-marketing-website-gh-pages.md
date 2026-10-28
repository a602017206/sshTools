# AHaSSHTools 宣传站与 GitHub Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在仓库新增纯静态宣传站 `website/`（首页 / 功能 / 下载，中英双语，四主题，three.js 粒子与鼠标跟随），并用 GitHub Actions 部署到 Pages。

**Architecture:** 零构建静态站点；CSS 变量驱动主题；`theme.js` / `i18n.js` 为可 Node 单测的纯逻辑 + DOM 应用层；`scene.js` 经 CDN 加载固定版本 three.js，在 hero 后绘制节点连线粒子场；`.github/workflows/gh-pages.yml` 用官方 deploy-pages 上传整个 `website/`。

**Tech Stack:** HTML / CSS / ES modules、three.js `0.170.0`（jsDelivr）、`node --test`、GitHub Actions Pages

**规格:** `docs/designs/2026-10-28-marketing-website-gh-pages-design.md`（与 `docs/superpowers/specs/` 副本一致）

## Global Constraints

- 纯静态，无 npm/Vite 构建；three.js 不 vendor，CDN 固定 `0.170.0`
- 资源与页间链接一律**相对路径**（适配 `https://a602017206.github.io/sshTools/`）
- 主题 id 仅：`graphite` | `amber` | `aurora` | `paper`；默认 `graphite`；键 `sshtools-site-theme`
- 语言仅：`zh` | `en`；默认 `zh`；键 `sshtools-site-lang`
- canvas：`pointer-events: none`；`prefers-reduced-motion: reduce` 或 WebGL 失败则不启场景
- 不改 `release.yml` / `mobile.yml`；不改 Terminal 剪贴板相关
- 文档中文；流程变更记 `docs/changes/process/`；功能交付记同一 process 文档即可（宣传站属仓库/发布流程）
- 截图从 `docs/designs/assets/aether-glass/` 与 `build/appicon.png` **复制**进 `website/assets/img/`，站点不引用 `docs/`

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `website/index.html` | 首页：hero + 能力摘要 + 平台 + CTA |
| `website/features.html` | 功能详情（成熟度标注） |
| `website/download.html` | 下载表与安装说明 |
| `website/assets/css/site.css` | 布局、token、四主题、降级背景 |
| `website/assets/js/theme.js` | 主题解析 / 应用 / 切换；导出纯函数供测 |
| `website/assets/js/i18n.js` | 文案字典与 DOM 应用；导出纯函数供测 |
| `website/assets/js/nav.js` | 顶栏当前页高亮、主题/语言控件绑定入口 |
| `website/assets/js/scene.js` | three.js 粒子场 + 鼠标吸引 + `setPalette` |
| `website/assets/img/*` | 图标与界面截图副本 |
| `website/test/theme.test.js` | 主题纯逻辑单测 |
| `website/test/i18n.test.js` | i18n 纯逻辑单测 |
| `.github/workflows/gh-pages.yml` | Pages 部署 |
| `README.md` | 增加宣传站链接 |
| `docs/changes/process/2026-10-28-marketing-website-gh-pages.md` | 变更记录 |
| `docs/plans/2026-10-28-marketing-website-gh-pages.md` | 本计划副本（与 superpowers/plans 同步） |

### 关键接口（跨任务约定）

```js
// theme.js
export const THEME_IDS = ['graphite', 'amber', 'aurora', 'paper'];
export const THEME_STORAGE_KEY = 'sshtools-site-theme';
export const THEME_PALETTES = {
  graphite: { bg: '#0b1214', accent: '#2dd4bf', particle: '#5eead4', line: '#2dd4bf66' },
  amber:    { bg: '#12100c', accent: '#f59e0b', particle: '#fbbf24', line: '#f59e0b66' },
  aurora:   { bg: '#0a0f18', accent: '#38bdf8', particle: '#67e8f9', line: '#38bdf866' },
  paper:    { bg: '#f4f7f6', accent: '#0f766e', particle: '#0d9488', line: '#0f766e44' },
};
export function normalizeThemeId(id) // unknown → 'graphite'
export function readStoredTheme(storage = localStorage) // → theme id
export function applyTheme(id, { documentElement, storage, onPalette } = {})
// sets data-theme, storage, calls onPalette?.(THEME_PALETTES[id])

// i18n.js
export const LANG_STORAGE_KEY = 'sshtools-site-lang';
export function normalizeLang(lang) // → 'zh' | 'en'
export function t(dict, lang, key) // missing key → key string
export function applyI18n(root, dict, lang) // [data-i18n], [data-i18n-html]

// scene.js
export function shouldStartScene({ matchMedia, webglAvailable }) // boolean
export async function mountScene(canvas, options) // returns { setPalette, destroy, pause, resume }
// options: { particleCount, density: 'full'|'light', getPalette }
```

---

### Task 1: 主题模块 + 单测

**Files:**
- Create: `website/assets/js/theme.js`
- Create: `website/test/theme.test.js`

**Interfaces:**
- Produces: `normalizeThemeId`, `readStoredTheme`, `applyTheme`, `THEME_IDS`, `THEME_PALETTES`, `THEME_STORAGE_KEY`

- [ ] **Step 1: 写失败单测**

```js
// website/test/theme.test.js
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  normalizeThemeId,
  readStoredTheme,
  applyTheme,
  THEME_STORAGE_KEY,
  THEME_PALETTES,
} from '../assets/js/theme.js';

describe('normalizeThemeId', () => {
  it('passes through known ids', () => {
    assert.equal(normalizeThemeId('amber'), 'amber');
  });
  it('falls back to graphite', () => {
    assert.equal(normalizeThemeId('nope'), 'graphite');
    assert.equal(normalizeThemeId(''), 'graphite');
  });
});

describe('readStoredTheme', () => {
  it('reads storage', () => {
    const storage = { getItem: () => 'paper' };
    assert.equal(readStoredTheme(storage), 'paper');
  });
  it('falls back when missing', () => {
    const storage = { getItem: () => null };
    assert.equal(readStoredTheme(storage), 'graphite');
  });
});

describe('applyTheme', () => {
  it('sets data-theme, storage, and palette callback', () => {
    const el = { setAttribute: (k, v) => { el[k] = v; } };
    const store = { setItem() {}, values: {} };
    store.setItem = (k, v) => { store.values[k] = v; };
    let palette;
    applyTheme('aurora', {
      documentElement: el,
      storage: store,
      onPalette: (p) => { palette = p; },
    });
    assert.equal(el['data-theme'], 'aurora');
    assert.equal(store.values[THEME_STORAGE_KEY], 'aurora');
    assert.deepEqual(palette, THEME_PALETTES.aurora);
  });
});
```

- [ ] **Step 2: 跑测确认失败**

Run: `cd website && node --test test/theme.test.js`  
Expected: FAIL（模块不存在或导出缺失）

- [ ] **Step 3: 实现 `theme.js`**

```js
// website/assets/js/theme.js
export const THEME_IDS = ['graphite', 'amber', 'aurora', 'paper'];
export const THEME_STORAGE_KEY = 'sshtools-site-theme';

export const THEME_PALETTES = {
  graphite: { bg: '#0b1214', accent: '#2dd4bf', particle: '#5eead4', line: '#2dd4bf66' },
  amber: { bg: '#12100c', accent: '#f59e0b', particle: '#fbbf24', line: '#f59e0b66' },
  aurora: { bg: '#0a0f18', accent: '#38bdf8', particle: '#67e8f9', line: '#38bdf866' },
  paper: { bg: '#f4f7f6', accent: '#0f766e', particle: '#0d9488', line: '#0f766e44' },
};

export function normalizeThemeId(id) {
  return THEME_IDS.includes(id) ? id : 'graphite';
}

export function readStoredTheme(storage = globalThis.localStorage) {
  try {
    return normalizeThemeId(storage?.getItem?.(THEME_STORAGE_KEY));
  } catch {
    return 'graphite';
  }
}

export function applyTheme(id, { documentElement, storage, onPalette } = {}) {
  const theme = normalizeThemeId(id);
  const root = documentElement || globalThis.document?.documentElement;
  root?.setAttribute?.('data-theme', theme);
  try {
    (storage || globalThis.localStorage)?.setItem?.(THEME_STORAGE_KEY, theme);
  } catch { /* ignore */ }
  onPalette?.(THEME_PALETTES[theme]);
  return theme;
}
```

- [ ] **Step 4: 跑测确认通过**

Run: `cd website && node --test test/theme.test.js`  
Expected: PASS

- [ ] **Step 5: Commit**（若用户要求提交时再执行；否则跳过）

```bash
git add website/assets/js/theme.js website/test/theme.test.js
git commit -m "$(cat <<'EOF'
feat(website): add theme module with unit tests

EOF
)"
```

---

### Task 2: i18n 模块 + 单测

**Files:**
- Create: `website/assets/js/i18n.js`
- Create: `website/test/i18n.test.js`

**Interfaces:**
- Produces: `normalizeLang`, `t`, `applyI18n`, `MESSAGES`, `LANG_STORAGE_KEY`, `readStoredLang`, `setStoredLang`
- Consumes: 无

- [ ] **Step 1: 写失败单测**

```js
// website/test/i18n.test.js
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { normalizeLang, t, applyI18n, MESSAGES } from '../assets/js/i18n.js';

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

describe('applyI18n', () => {
  it('fills data-i18n text', () => {
    const el = {
      getAttribute: () => 'nav.home',
      textContent: '',
      hasAttribute: (n) => n === 'data-i18n',
      querySelectorAll: () => [el],
    };
    // Minimal stub: implement applyI18n to accept array via root.querySelectorAll
    const root = {
      querySelectorAll: (sel) => (sel.includes('data-i18n-html') ? [] : [el]),
    };
    applyI18n(root, MESSAGES, 'zh');
    assert.ok(el.textContent.length > 0);
  });
});
```

- [ ] **Step 2: 跑测确认失败**

Run: `cd website && node --test test/i18n.test.js`  
Expected: FAIL

- [ ] **Step 3: 实现 `i18n.js`（含完整首期字典）**

字典至少包含键（中英都要有真实文案，对齐 README 语气）：

- `brand`, `tagline`, `nav.home`, `nav.features`, `nav.download`, `nav.github`, `nav.theme`, `nav.lang`
- `home.ctaFeatures`, `home.ctaDownload`, `home.capabilitiesTitle`, 以及 6 个能力短标题/描述：`cap.ssh.*`, `cap.sftp.*`, `cap.monitor.*`, `cap.data.*`, `cap.ai.*`, `cap.mobile.*`
- `home.platformsTitle`, `features.*` 各功能域标题与正文、`maturity.full|ready|early`
- `download.title`, `download.noteMac`, `download.noteAndroid`, `download.noteIos`, 各平台行标签

```js
// website/assets/js/i18n.js（结构示意；实现时写满中英字符串）
export const LANG_STORAGE_KEY = 'sshtools-site-lang';

export const MESSAGES = {
  zh: {
    brand: 'AHaSSHTools',
    tagline: '跨平台运维客户端：SSH、文件、监控、数据与 AI',
    'nav.home': '首页',
    'nav.features': '功能',
    'nav.download': '下载',
    // ... 其余键按上表补全
  },
  en: {
    brand: 'AHaSSHTools',
    tagline: 'Cross-platform ops client: SSH, files, monitoring, data, and AI',
    'nav.home': 'Home',
    'nav.features': 'Features',
    'nav.download': 'Download',
    // ...
  },
};

export function normalizeLang(lang) {
  return lang === 'en' ? 'en' : 'zh';
}

export function readStoredLang(storage = globalThis.localStorage) {
  try {
    return normalizeLang(storage?.getItem?.(LANG_STORAGE_KEY));
  } catch {
    return 'zh';
  }
}

export function setStoredLang(lang, storage = globalThis.localStorage) {
  const v = normalizeLang(lang);
  try {
    storage?.setItem?.(LANG_STORAGE_KEY, v);
  } catch { /* ignore */ }
  return v;
}

export function t(dict, lang, key) {
  const table = dict[normalizeLang(lang)] || dict.zh || {};
  return table[key] ?? key;
}

export function applyI18n(root, dict, lang) {
  const L = normalizeLang(lang);
  root.querySelectorAll('[data-i18n]').forEach((node) => {
    const key = node.getAttribute('data-i18n');
    if (key) node.textContent = t(dict, L, key);
  });
  root.querySelectorAll('[data-i18n-html]').forEach((node) => {
    const key = node.getAttribute('data-i18n-html');
    if (key) node.innerHTML = t(dict, L, key);
  });
  root.documentElement?.setAttribute?.('lang', L === 'zh' ? 'zh-CN' : 'en');
  // 若 root 是 document，上面 documentElement 可用；若是元素，调用方另设 html lang
}
```

实现时：`applyI18n` 若传入 `document`，应对 `document.documentElement` 设 `lang`；单测用假 root 即可。

- [ ] **Step 4: 跑测确认通过**

Run: `cd website && node --test test/i18n.test.js test/theme.test.js`  
Expected: 全部 PASS

- [ ] **Step 5: Commit**（用户要求时）

```bash
git add website/assets/js/i18n.js website/test/i18n.test.js
git commit -m "$(cat <<'EOF'
feat(website): add i18n dictionary and helpers

EOF
)"
```

---

### Task 3: 站点 CSS 与素材

**Files:**
- Create: `website/assets/css/site.css`
- Create: `website/assets/img/`（复制素材）

**Interfaces:**
- Consumes: `data-theme` 属性名与 Task 1 一致
- Produces: 页面可用的布局 class：`.site-header`, `.hero`, `.cap-grid`, `.shot-strip`, `#bg-canvas`, `.theme-dots` 等

- [ ] **Step 1: 复制并压缩素材**

```bash
mkdir -p website/assets/img
cp build/appicon.png website/assets/img/appicon.png
cp docs/designs/assets/aether-glass/aether-glass-dark-ssh.png website/assets/img/shot-ssh-dark.png
cp docs/designs/assets/aether-glass/aether-glass-dark-database.png website/assets/img/shot-db-dark.png
cp docs/designs/assets/aether-glass/aether-glass-light-ssh.png website/assets/img/shot-ssh-light.png
# 若本机有 sips/pngquant 可压缩；无则保留原图并在变更文档记风险
```

- [ ] **Step 2: 编写 `site.css`**

要求：
- `@import` 或 `<link>` 字体在 HTML 中引入；CSS 内定义四主题：

```css
:root, [data-theme="graphite"] {
  --bg: #0b1214;
  --fg: #e8eef0;
  --muted: #9db0b5;
  --accent: #2dd4bf;
  --surface: #121a1c;
  --border: #1e2a2e;
}
[data-theme="amber"] { /* 对齐 THEME_PALETTES */ }
[data-theme="aurora"] { /* ... */ }
[data-theme="paper"] {
  --bg: #f4f7f6;
  --fg: #12201e;
  --muted: #5b6f6c;
  --accent: #0f766e;
  --surface: #ffffff;
  --border: #d5e0dd;
}
```

- `#bg-canvas`：`position: fixed; inset: 0; z-index: 0; pointer-events: none;`
- `.page`：`position: relative; z-index: 1;`
- `.hero` 全宽；截图条 `.shot-strip img` 最大宽度限制，圆角克制（≤8px）
- `@media (prefers-reduced-motion: reduce)`：可隐藏 canvas
- 响应式：窄屏单列能力摘要

- [ ] **Step 3: 目视确认文件存在**

Run: `ls -la website/assets/css/site.css website/assets/img/`  
Expected: css 与至少 appicon + 2 张 shot 存在

- [ ] **Step 4: Commit**（用户要求时）

---

### Task 4: three.js 场景模块

**Files:**
- Create: `website/assets/js/scene.js`
- Create: `website/test/scene.test.js`（仅测 `shouldStartScene`）

**Interfaces:**
- Consumes: `THEME_PALETTES` 形状 `{ bg, accent, particle, line }`
- Produces: `shouldStartScene`, `mountScene` → `{ setPalette, destroy, pause, resume }`

- [ ] **Step 1: 写 `shouldStartScene` 单测**

```js
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { shouldStartScene } from '../assets/js/scene.js';

describe('shouldStartScene', () => {
  it('false when reduced motion', () => {
    assert.equal(
      shouldStartScene({
        matchMedia: (q) => ({ matches: q.includes('prefers-reduced-motion') }),
        webglAvailable: true,
      }),
      false,
    );
  });
  it('false when no webgl', () => {
    assert.equal(
      shouldStartScene({
        matchMedia: () => ({ matches: false }),
        webglAvailable: false,
      }),
      false,
    );
  });
  it('true otherwise', () => {
    assert.equal(
      shouldStartScene({
        matchMedia: () => ({ matches: false }),
        webglAvailable: true,
      }),
      true,
    );
  });
});
```

- [ ] **Step 2: 跑测失败 → 实现 `shouldStartScene` + `mountScene`**

`mountScene` 要点（实现须完整写入文件，此处为约束清单）：
- 动态 `import('https://cdn.jsdelivr.net/npm/three@0.170.0/build/three.module.js')`
- `PerspectiveCamera` + `WebGLRenderer({ canvas, alpha: true, antialias: true })`
- 粒子：`BufferGeometry` + `Points`；桌面默认 count≈120，`density: 'light'` 时 ≈50
- 近邻连线：每帧或节流重建淡线（限制最大边数，如 200）
- 鼠标：`pointermove` 在 `window` 上听，归一化到 [-1,1]，对粒子施加强度有限的吸引
- `setPalette(p)` 更新 Points/Line 颜色与 clear 色
- `visibilitychange` → pause/resume
- 失败时 `destroy` 并 rethrow 或返回 null（调用方 catch）

```js
export function shouldStartScene({ matchMedia, webglAvailable }) {
  const reduced = matchMedia?.('(prefers-reduced-motion: reduce)')?.matches;
  return Boolean(webglAvailable) && !reduced;
}

function detectWebGL() {
  try {
    const c = document.createElement('canvas');
    return !!(c.getContext('webgl') || c.getContext('experimental-webgl'));
  } catch {
    return false;
  }
}
```

- [ ] **Step 3: 跑测**

Run: `cd website && node --test test/scene.test.js`  
Expected: PASS（`mountScene` 不在 Node 里跑 WebGL）

- [ ] **Step 4: Commit**（用户要求时）

---

### Task 5: 导航脚本 + 三页 HTML

**Files:**
- Create: `website/assets/js/nav.js`
- Create: `website/index.html`
- Create: `website/features.html`
- Create: `website/download.html`

**Interfaces:**
- Consumes: `applyTheme`, `readStoredTheme`, `THEME_IDS`, `applyI18n`, `readStoredLang`, `setStoredLang`, `MESSAGES`, `mountScene`, `shouldStartScene`

- [ ] **Step 1: 实现 `nav.js`**

```js
import { applyTheme, readStoredTheme, THEME_IDS } from './theme.js';
import { applyI18n, readStoredLang, setStoredLang, MESSAGES } from './i18n.js';

export function initChrome({ sceneApi } = {}) {
  const theme = applyTheme(readStoredTheme(), {
    onPalette: (p) => sceneApi?.setPalette?.(p),
  });
  let lang = readStoredLang();
  applyI18n(document, MESSAGES, lang);
  document.documentElement.lang = lang === 'zh' ? 'zh-CN' : 'en';

  document.querySelectorAll('[data-set-theme]').forEach((btn) => {
    btn.addEventListener('click', () => {
      const id = btn.getAttribute('data-set-theme');
      applyTheme(id, { onPalette: (p) => sceneApi?.setPalette?.(p) });
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
  document.querySelectorAll('[data-nav]').forEach((a) => {
    if (a.getAttribute('data-nav') === page) a.setAttribute('aria-current', 'page');
  });

  return { theme, lang, THEME_IDS };
}
```

- [ ] **Step 2: 写 `index.html` 骨架**

必备：
- fonts link（Google Fonts 或 fontsource CDN：IBM Plex Mono + Source Sans 3）
- `<canvas id="bg-canvas" aria-hidden="true"></canvas>`
- header 导航相对链接：`index.html` / `features.html` / `download.html`
- GitHub：`https://github.com/a602017206/sshTools`
- hero：`data-i18n` 品牌与 tagline；CTA 链到功能/下载
- 能力 6 格；平台一行；截图条使用 `assets/img/shot-*.png`
- 底部 module：

```html
<script type="module">
  import { initChrome } from './assets/js/nav.js';
  import { shouldStartScene, mountScene } from './assets/js/scene.js';
  import { THEME_PALETTES, readStoredTheme } from './assets/js/theme.js';

  let sceneApi;
  const canvas = document.getElementById('bg-canvas');
  const webglAvailable = (() => {
    try {
      const c = document.createElement('canvas');
      return !!(c.getContext('webgl') || c.getContext('experimental-webgl'));
    } catch { return false; }
  })();

  if (canvas && shouldStartScene({ matchMedia: window.matchMedia.bind(window), webglAvailable })) {
    try {
      sceneApi = await mountScene(canvas, {
        density: 'full',
        getPalette: () => THEME_PALETTES[readStoredTheme()],
      });
    } catch (e) {
      console.warn('scene disabled', e);
      sceneApi = null;
    }
  }
  initChrome({ sceneApi });
</script>
```

HTML 注释标明：`three@0.170.0 via jsDelivr inside scene.js`

- [ ] **Step 3: 写 `features.html` / `download.html`**

- 功能页：按 README 域分段；成熟度用 `data-i18n`；内页 `density: 'light'` 或可省略 scene（规格允许轻量；**实现选 light 挂载**）
- 下载页：表格链到  
  `https://github.com/a602017206/sshTools/releases/latest/download/AHaSSHTools.exe`  
  以及 macos arm64/amd64 zip、android 三 ABI apk、ios unsigned ipa（文件名与 README 一致）  
  + macOS `xattr`、Android ABI、iOS 自签说明

- [ ] **Step 4: 本地静态服务验收**

Run:

```bash
cd website && python3 -m http.server 8765
```

浏览器打开 `http://127.0.0.1:8765/`，检查清单：
- [ ] 三页可切换
- [ ] 中英切换文案变化且刷新保持
- [ ] 四主题切换，粒子色随变
- [ ] 鼠标移动粒子有响应
- [ ] 点击 CTA / 下载链不被 canvas 挡住

- [ ] **Step 5: Commit**（用户要求时）

---

### Task 6: GitHub Pages 工作流 + README + 变更文档

**Files:**
- Create: `.github/workflows/gh-pages.yml`
- Modify: `README.md`（下载表之前或文档表增加宣传站一行）
- Create: `docs/changes/process/2026-10-28-marketing-website-gh-pages.md`
- Create: `docs/plans/2026-10-28-marketing-website-gh-pages.md`（从 superpowers/plans 复制）

- [ ] **Step 1: 写 workflow**

```yaml
name: Deploy GitHub Pages

on:
  push:
    branches: [master]
    paths:
      - 'website/**'
      - '.github/workflows/gh-pages.yml'
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: true

jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Verify website entry
        run: test -f website/index.html

      - uses: actions/configure-pages@v5

      - uses: actions/upload-pages-artifact@v3
        with:
          path: website

      - id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: 更新 README**

在合适位置增加（中文）：

```markdown
## 宣传网站

产品介绍站（GitHub Pages）：https://a602017206.github.io/sshTools/
```

- [ ] **Step 3: 写变更文档**

`docs/changes/process/2026-10-28-marketing-website-gh-pages.md` 必须含：背景、范围、修改文件、验证、剩余风险；并写明人工检查：**Settings → Pages → Source = GitHub Actions**。

- [ ] **Step 4: 同步计划副本**

```bash
cp docs/superpowers/plans/2026-10-28-marketing-website-gh-pages.md \
   docs/plans/2026-10-28-marketing-website-gh-pages.md
```

- [ ] **Step 5: 跑全部 website 单测**

Run: `cd website && node --test test/*.test.js`  
Expected: PASS

- [ ] **Step 6: Commit**（用户要求时）

---

## 规格覆盖自检

| 规格项 | 任务 |
|--------|------|
| 三页信息架构 | Task 5 |
| 中英 i18n | Task 2, 5 |
| 四主题 + localStorage | Task 1, 3, 5 |
| three.js 粒子 + 鼠标跟随 + 降级 | Task 4, 5 |
| 素材拷贝 | Task 3 |
| gh-pages.yml + 官方 Actions | Task 6 |
| README 链接 + 变更文档 | Task 6 |
| 相对路径 / 子路径 | Task 5 Global Constraints |
| 不改 release/mobile | Task 6（仅新增 workflow） |

## Placeholder 扫描

计划内无 TBD；i18n 字典在 Task 2 Step 3 要求「写满」真实中英，实现者不得留空键。
