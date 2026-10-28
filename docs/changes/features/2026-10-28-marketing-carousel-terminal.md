# 变更：宣传站轮播、终端打字与跑马灯

## 背景

在全屏 3D 与卡片动效基础上，继续提升炫酷感；用户希望增加轮播等可感知的内容特效。不使用概念截图，改为文案驱动的场景轮播。

## 范围

- Hero 增加终端风格打字机预演
- 首页增加能力场景自动轮播（悬停暂停）
- 技术栈横向跑马灯
- 页面跟随光标的柔光
- 中英文字典补充轮播文案

## 修改文件

- `website/index.html`
- `website/assets/css/site.css`
- `website/assets/js/carousel.js`
- `website/assets/js/typewriter.js`
- `website/assets/js/fx.js`
- `website/assets/js/i18n.js`
- `website/test/carousel.test.js`
- `website/test/typewriter.test.js`
- `docs/changes/features/2026-10-28-marketing-carousel-terminal.md`

## 验证

```bash
cd website && node --test test/*.test.js
```

本地打开首页：确认终端打字、轮播切换、跑马灯与光标柔光。

## 剩余风险

- 打字机文案在语言切换后需刷新页面才重建（当前首屏按存储语言初始化）
- 跑马灯为装饰性，读屏器用 `aria-hidden`
