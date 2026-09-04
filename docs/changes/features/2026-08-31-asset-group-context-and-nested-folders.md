# 资产树文件夹右键新建与多级分组

## 背景

资产树文件夹原先只能点击展开，新建连接时分组要手工填写。分组字段也只按整段字符串渲染一层，无法表示环境/区域这类多级目录。

设计文档：`docs/designs/2026-08-31-asset-group-context-and-nested-folders.md`。

## 范围

- 文件夹右键提供「新建连接」，并把该文件夹完整路径预填到分组字段
- 分组支持 `/` 分隔的多级目录，资产树按路径嵌套展示与展开
- 顶部加号、工作区空态等原有新建入口保持空白分组

不包含空文件夹实体、拖拽调整分组、文件夹重命名。

## 修改文件

- `frontend/src/lib/assetGroupTree.js`（新建）
- `frontend/test/assetGroupTree.test.js`（新建）
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/App.svelte`
- `docs/designs/2026-08-31-asset-group-context-and-nested-folders.md`
- `docs/changes/features/2026-08-31-asset-group-context-and-nested-folders.md`（本文）

## 验证

- `cd frontend && node --test test/assetGroupTree.test.js test/connectionDialogRequest.test.js`
- `cd frontend && npx vite build`
- 手工：在文件夹上右键「新建连接」，分组字段应带入该文件夹路径；填写 `生产/华东` 后资产树出现嵌套文件夹

## 剩余风险

- 既有分组名若包含 `/`，会被解释成多级目录
- 桌面端右键菜单位置未做视口夹紧，靠近窗口边缘时可能被裁切
- 本次无法在 Wails 窗口内自动点选验证，右键与预填仍需本地跑一次应用确认
