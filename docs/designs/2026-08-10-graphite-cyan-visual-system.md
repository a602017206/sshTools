# 设计：Graphite Cyan 视觉系统（Dark First）

## 背景

一体化运维工作台（SSH + 数据库）功能已完备，但界面存在层级坍塌、色彩失序、状态依赖弹窗等问题，整体停留在“能用但不想用”。需要对齐 Warp / TablePlus / DataGrip 一类专业工具审美。

## 目标

- **Dark First**：默认深色主题，亮色为同一语义 Token 的映射。
- **视觉降噪**：区域靠表面抬升与微对比，少用粗边框。
- **状态信号化**：连接/执行状态用左侧色条与呼吸灯传达，避免弹窗轰炸。
- **高信息密度但有呼吸感**：侧栏默认 260px，主区编辑器/结果默认 58/42。

## 决策摘要

| 维度 | 决策 |
|---|---|
| 品牌色 | 低饱和极客青（Teal / Graphite Cyan） |
| 默认主题 | `theme=dark`，`theme_mode=dark` |
| UI 字体 | SF Pro Text / Inter + 中文系统字体 |
| 等宽字体 | JetBrains Mono（终端 / SQL / 结果单元格） |
| 侧栏 | 默认 260，最小 200，最大 420；双击分割线复位 |
| 上下分割 | 默认 58% 编辑器 / 42% 结果；1px + 微阴影分割线 |
| 活跃会话 | 背景渐变 + 左侧 3px 发光竖条 |
| 连接状态 | 名称前呼吸灯：连接中呼吸 / 已连接常亮 / 报错快闪 |
| 终端氛围 | 深色模式下极轻点阵网格（可关） |
| 品牌水印 | 侧栏左下：菱形节点标 + `AHa vX` |

## 非目标

- 不改后端连接/查询逻辑。
- 不一次性清扫所有历史 Tailwind 硬编码色（分阶段迁移）。
- 不引入新的 UI 框架。

## 落地入口

- Token：`frontend/src/styles/app.css`
- 默认外观：`frontend/src/settings/appearance.js`、`internal/config/config.go`
- 组件：`AssetList`、`DatabasePanel`、`DatabaseTablePanel`、`Terminal`、`App`
