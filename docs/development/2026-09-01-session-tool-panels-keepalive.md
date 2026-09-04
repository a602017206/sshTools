# 开发记录：会话工具文件与性能保持挂载

## 实现内容

- `SessionToolDock` 在已绑定会话时同时渲染 `FileManager` 和 `ServerMonitor`，用 `sshToolPanelHidden` + `hidden` 切换可见性，不再 `{#if}` 销毁未激活面板。
- 切标签时保留文件当前路径、目录跟踪订阅，以及性能页已采集曲线和轮询。

## 验证

- `cd frontend && node --test src/lib/workspaceTabs.test.js test/sessionToolDock.test.js`：9 项通过。

## 剩余风险

后台监控轮询会持续发 SSH 命令。隐藏面板初次挂载时若宽高为 0，布局依赖 CSS 百分比，当前监控视图未使用需测量画布的 Chart.js。
