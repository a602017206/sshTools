# 变更：宣传站首页移除概念图

## 背景

首页底部三张图来自 `aether-glass` 设计概念图，并非产品实际截图，不宜作为宣传证据展示。

## 范围

- 删除首页截图条带
- 删除对应 PNG 资源与无用 CSS
- 保留真实应用图标 `appicon.png`

## 修改文件

- `website/index.html`
- `website/assets/css/site.css`
- `website/assets/img/shot-ssh-dark.png`（删除）
- `website/assets/img/shot-ssh-light.png`（删除）
- `website/assets/img/shot-db-dark.png`（删除）
- `docs/changes/bugs/2026-10-28-marketing-remove-concept-shots.md`（本文件）

## 验证

- 首页不再引用 `shot-*.png`
- `website/assets/img/` 仅保留 `appicon.png`

## 剩余风险

- 首页暂无真实产品截图；待有实机截图后再补展示区
