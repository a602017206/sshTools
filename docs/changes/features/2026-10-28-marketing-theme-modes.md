# 营销站主题：背景/字体/三种组合模型

## 背景

官网主题原先只有 4 个完整预设色点。需要把主题拆成背景色与字体色，并提供三种组合模型，同时扩充可选主题数量。

## 范围

- 组合模型：`preset`（命名预设）、`duo`（背景 + 字体）、`trio`（背景 + 字体 + 强调色）
- 预设主题扩充至 10 个：graphite / amber / aurora / paper / ember / ocean / forest / orchid / sand / noir
- 导航改为可展开主题面板；CSS 变量由 `theme.js` 运行时写入
- 中英文文案与单测同步更新

## 修改文件

- `website/assets/js/theme.js`
- `website/assets/js/nav.js`
- `website/assets/js/i18n.js`
- `website/assets/css/site.css`
- `website/index.html` / `website/features.html` / `website/download.html`
- `website/test/theme.test.js` / `website/test/nav.test.js`
- `docs/changes/features/2026-10-28-marketing-theme-modes.md`

## 验证

```bash
cd website && node --test test/theme.test.js test/nav.test.js test/i18n.test.js
```

## 剩余风险

- 旧 localStorage 仅存 `sshtools-site-theme` 时会按 preset 模式读取；duo/trio 新键缺失时回落默认 bg/fg/accent
- 浅色方案下鼠标光晕改为依赖 `data-scheme="light"`，需在浏览器中目视确认
