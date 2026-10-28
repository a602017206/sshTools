# 营销站鼠标造型改为描线发光 SVG

## 背景

用户不希望使用 emoji 表情作为光标；需要线稿动物/特效造型，并带主题色光效。

## 范围

- 全部造型改为 32×32 描线 SVG（无 emoji、无填充色块）
- 主题面板预览、主光标、拖尾与点击爆开统一使用线稿 + `drop-shadow` 光效
- 颜色跟随 `--accent` / `--glow`，切换主题即时变色

## 修改文件

- `website/assets/js/cursor.js`
- `website/assets/js/fx.js`
- `website/assets/js/nav.js`
- `website/assets/css/site.css`
- `website/test/cursor.test.js`
- `docs/changes/features/2026-10-28-marketing-cursor-lineart.md`

## 验证

```bash
cd website && node --test test/cursor.test.js test/fx.test.js test/nav.test.js
```

## 剩余风险

- 生肖线稿为简化识别稿，细节有限
- 旧 localStorage 若存了无效 id 会回落到小猫
