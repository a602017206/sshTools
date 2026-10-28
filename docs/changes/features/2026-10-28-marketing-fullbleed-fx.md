# 变更：宣传站全屏 3D 与区块动效

## 背景

首页 three.js 效果只挤在左上角一小块，能力说明卡片也偏静态，整体不够炫。

## 范围

- 粒子场铺满视口：更多节点、远景星点、多核心枢纽、更大拓扑环
- 页面增加氛围光斑与网格蒙版
- 能力 / 功能 / 下载卡片：玻璃拟态、边框高光、鼠标倾斜光斑、滚动显现
- 平台标签与下载表增加微动效
- 新增 `fx.js`（尊重 `prefers-reduced-motion`）

## 修改文件

- `website/assets/js/scene.js`
- `website/assets/js/fx.js`
- `website/assets/css/site.css`
- `website/index.html`
- `website/features.html`
- `website/download.html`
- `website/test/fx.test.js`
- `docs/changes/features/2026-10-28-marketing-fullbleed-fx.md`

## 验证

```bash
cd website && node --test test/*.test.js
```

本地静态服务打开首页：确认粒子铺满、卡片悬停倾斜与滚动显现。

## 剩余风险

- 全屏 `density: full` 在低端机略增 GPU 占用；可按需再回调内页密度
- 卡片 `backdrop-filter` 在极旧浏览器上效果会降级为普通半透明底
