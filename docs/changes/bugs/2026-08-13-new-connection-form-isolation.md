# 修复新建连接复用编辑表单的问题

## 背景

编辑连接后再次点击“新建连接”时，弹窗会保留上一份编辑表单的数据。新建连接必须始终从空白默认表单开始，不能依赖弹窗是否已经完成关闭动画或状态更新顺序。

## 范围

仅调整连接弹窗的新建、编辑和克隆请求状态隔离：每次请求都重建弹窗实例。既有编辑异步加载、克隆预填和密码保存行为保持不变。

## 修改文件

- `frontend/src/App.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/lib/connectionDialogRequest.js`
- `frontend/test/connectionDialogRequest.test.js`
- `.github/workflows/release.yml`

## 验证

- 运行连接弹窗请求、新建表单、编辑加载、克隆和用户名输入相关测试。
- 运行前端生产构建。

## 剩余风险

该修复通过请求序号隔离相同弹窗中的不同操作。仍建议在桌面端手动依次验证“编辑 A → 关闭 → 新建”“编辑 A → 编辑 B”“克隆 A → 新建”。
