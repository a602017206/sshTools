# 缺陷：会话工具文件与性能切换丢失状态

## 背景

「文件」和「性能」用 `{#if}` 互斥渲染。切到性能后文件组件被销毁，目录跟踪和当前路径丢失，再回去会落到用户家目录；切回性能时监控定时器和曲线也要从头拉，体感很慢。

## 范围

- 已绑定 SSH 会话时同时挂载两个面板，仅用 `hidden` 切换可见性
- 补充 `sshToolPanelHidden` 与源码契约测试
- 发布说明增加本条修复

## 修改文件

- `docs/designs/2026-09-01-session-tool-panels-keepalive.md`
- `docs/development/2026-09-01-session-tool-panels-keepalive.md`
- `frontend/src/lib/workspaceTabs.js`
- `frontend/src/lib/workspaceTabs.test.js`
- `frontend/src/components/SessionToolDock.svelte`
- `frontend/test/sessionToolDock.test.js`
- `.github/workflows/release.yml`

## 验证

```bash
cd frontend && node --test src/lib/workspaceTabs.test.js test/sessionToolDock.test.js
```

9 项通过。未在桌面应用内对真实 SSH 会话做手工切换验证。

## 剩余风险

性能页在后台继续 2 秒轮询，会占用一条 SSH 命令通道。断开会话时两个面板仍会一起卸载。
