# 变更：宣传站 three.js 氛围层增强

## 背景

首页粒子场偏平面、动效偏弱，希望更有立体感与互动反馈，同时保持运维节点隐喻与主题色同步。

## 范围

- 加深粒子景深（含 Z 轴运动与边界夹紧）
- 发光点纹理、加性连线、中央线框核心节点
- 旋转拓扑环、底部透视网格
- 鼠标视差相机与更强的三维吸引
- 更新边界 / 指针映射相关单测

## 修改文件

- `website/assets/js/scene.js`
- `website/test/scene.test.js`
- `docs/changes/features/2026-10-28-marketing-scene-richer-3d.md`

## 验证

```bash
cd website && node --test test/*.test.js
```

本地用静态服务打开首页，移动鼠标观察视差、核心旋转与连线密度。

## 剩余风险

- 低端 GPU 上首页粒子与环略增负载；内页仍用 `density: 'light'`
- 仍无后处理 bloom；若再加强需评估额外 CDN 依赖
