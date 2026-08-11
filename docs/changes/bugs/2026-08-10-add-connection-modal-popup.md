# 变更：恢复添加连接为真正弹窗

## 背景

玻璃拟态改造后，「添加连接」虽仍走 `Dialog`，但遮罩过淡，且弹层挂在 `ops-shell` 内，观感像嵌在主工作区的卡片，而不像居中弹窗。

## 范围

- `Dialog` 通过 portal 挂到 `document.body`，全屏强遮罩 + 毛玻璃面板。
- 亮色遮罩加深；点击遮罩/ESC 关闭。
- 「添加连接」弹窗宽度改为 `md`。
- 不改表单字段与保存/测试连接逻辑。

## 修改文件

- `frontend/src/components/ui/Dialog.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/styles/app.css`
- `docs/changes/bugs/2026-08-10-add-connection-modal-popup.md`（本文件）

## 验证

- 建议手工：`wails dev` 点侧栏「+」，确认整屏变暗、表单居中浮起；点遮罩或取消可关闭。

## 剩余风险

- 其他基于 `Dialog` 的确认框会一并获得更强遮罩，这是预期统一行为。
