# Docker 容器详情侧栏（概览 + 日志跟流）

## 背景
Docker 二期要求在容器列表右侧提供详情侧栏，展示 Inspect 概览，并支持日志查看与实时跟流。

## 范围
- 新增详情侧栏组件，含概览与日志两个 Tab。
- 日志支持尾部行数选择、跟流、清空、复制、节流拼接、自动滚底与上滚暂停。
- 关闭侧栏、切换容器、组件销毁时停止日志跟流。
- 不涉及后端改动，不涉及终端剪贴板受保护代码。

## 修改文件
- `frontend/src/components/DockerContainerDetail.svelte`（新增）
- `frontend/src/components/DockerWorkspace.svelte`
- `frontend/src/lib/dockerLogs.js`（新增）
- `frontend/test/dockerLogs.test.js`（新增）

## 验证
- `node --test test/dockerLogs.test.js test/dockerWorkspace.test.js` 与终端剪贴板契约测试通过。
- `npx vite build` 成功。

## 剩余风险
- 侧栏宽度暂不可拖拽；未在真实 Docker 引擎上手工验证跟流。
