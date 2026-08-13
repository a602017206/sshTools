# 连接克隆与新增表单重置实施计划

> **执行说明：** 按任务顺序实施；每个功能步骤先执行测试驱动开发流程。

**目标：** 防止新增连接表单残留编辑内容，并支持从右键菜单克隆连接。

**架构：** 在 `frontend/src/lib` 中提供可独立测试的表单初始化与克隆名称函数。`App.svelte` 维护新增、编辑、克隆三种对话框输入，`AssetList.svelte` 负责把右键动作上报，`AddAssetDialog.svelte` 根据输入填充表单。

**技术栈：** Svelte 4、Node 内置测试运行器、Wails 前端绑定。

---

### 任务 1：表单初始化与克隆数据规则

**文件：**
- 新建：`frontend/src/lib/connectionFormData.js`
- 新建：`frontend/test/connectionFormData.test.js`

1. 编写失败测试：验证默认表单为空白且 SSH 默认端口为 `22`。
2. 运行 `node --test test/connectionFormData.test.js`，确认因模块不存在而失败。
3. 实现最小默认表单和克隆预填函数；克隆名称为“原名称 copy YYYYMMDDHHmmss”，且不复制密码、口令和 ID。
4. 重跑测试，确认通过。

### 任务 2：对话框状态与克隆入口

**文件：**
- 修改：`frontend/src/App.svelte`
- 修改：`frontend/src/components/AddAssetDialog.svelte`
- 修改：`frontend/src/components/AssetList.svelte`

1. 新增连接时显式清空对话框输入状态。
2. 为每个连接的右键菜单增加“克隆连接”，上报源连接到 `App`。
3. 以克隆预填数据打开表单，保存时创建新连接，且名称允许编辑。
4. 确保编辑、关闭、取消、重复打开后均不泄漏上一条表单状态。

### 任务 3：验证与记录

**文件：**
- 新建：`docs/changes/features/2026-08-13-connection-clone-and-form-reset.md`

1. 运行 `node --test test/connectionFormData.test.js`。
2. 运行 `npm run build`。
3. 记录修改范围、验证结果和剩余风险。
