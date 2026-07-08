# 修复添加/编辑连接弹窗背景不跟随主题

## 类型

Bug 修复

## 问题

编辑连接信息弹窗中，表单输入框、认证方式切换、连接类型切换、分组下拉和底部按钮仍使用固定 Tailwind 颜色，导致主题切换后局部背景和文字颜色与整体界面不一致。

## 修复

- 新增弹窗表单相关主题工具类，包括输入框、只读输入、选择项、下拉菜单、弱按钮、提示文字和测试结果状态。
- 将 `AddAssetDialog.svelte` 中的硬编码灰色/紫色/深色背景替换为主题 token 驱动样式。
- 顺手补齐该弹窗内输入框 label 关联，并将分组下拉项改为可聚焦按钮，减少当前文件的构建警告。
- 保留现有添加、编辑、测试连接和默认端口逻辑。

## 验证

- 已执行 `cd frontend && npm run build`，通过；`AddAssetDialog.svelte` 相关 a11y 警告已清理，项目其他文件仍存在既有 Svelte a11y 和 chunk size 警告。
- 已执行 `GOCACHE="$(pwd)/.cache/go-build" go test ./...`，通过。
- 已执行 `GOCACHE="$(pwd)/.cache/go-build" go vet ./...`，通过。
