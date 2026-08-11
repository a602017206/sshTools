# 开发记录：运维工作区 Glassmorphism 落地

## 目标

在已完成的 SSH / 数据库双模式壳上，按玻璃拟态设计统一视觉与微动效，不破坏会话保活与现有连接流程。

## 实现要点

1. **Token 与壳层**  
   在 `app.css` 扩展 `--glass-*`、`--overlay-scrim`、`--shadow-glass`、`--trans-*`；`ops-shell` 使用径向光晕 + 渐变沉浸背景；顶栏/侧栏/面板/工具轨使用毛玻璃。

2. **可感知玻璃（二次强化）**  
   根因是背景过淡 + 面板 alpha 过高，半透明看起来像白板。已改为：  
   - `ops-shell::before` 鲜艳氛围色块（可缓慢漂移）  
   - 顶栏/侧栏/主面板 alpha 降到约 0.16–0.28  
   - 去掉 `isolation: isolate`，避免削弱 backdrop 采样  
   - 数据库空状态改为独立毛玻璃卡片  
   不使用贴图；CSS 半透明 + `backdrop-filter` 即可。

3. **组件对齐**  
   - `WorkspaceNavigation`：胶囊分段切换，激活态信号色。  
   - `SessionToolDock` / `DatabaseWorkspaceEmpty`：透明底 + 玻璃按钮。  
   - `TerminalPanel`：标签栏与工具条半透明玻璃，终端内容区仍为深色视口。  
   - `Dialog`：毛玻璃遮罩 + `glass-rise` 弹入；点击遮罩关闭。  
   - 上传飞出层、资源树/表对象右键菜单：`ops-flyout` 风格。

4. **交互**  
   图标按钮、软按钮、模式 pill 增加短时过渡与轻微 `scale` 反馈；尊重 `prefers-reduced-motion`。

## 未做事项

- 未恢复概览顶栏，未卸载 `TerminalPanel`。  
- JDBC 驱动管理等次要对话框内部仍有部分非 glass 局部样式，可后续迭代。  
- 高保真中的快速命令输入条、底栏延迟/加密信息尚未落地（本期聚焦终端深色视口与文件管理玻璃化）。

## 验证说明

见同日变更文档中的命令与手工检查项。
