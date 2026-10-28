# 营销网站场景颜色与可见性修复

## 背景

营销网站场景使用的主题连线色为 `#RRGGBBAA`。`three@0.170.0` 的 `Color` 不解析这类包含 alpha 的十六进制颜色，导致连线颜色不能按预期显示。此外，页面在隐藏状态挂载时仍会请求动画帧。

## 范围

- 为场景颜色增加纯解析函数，将八位十六进制颜色拆分为 RGB 颜色值和不透明度。
- 在连线材质的初始创建与主题切换时应用拆分后的颜色和透明度。
- 页面隐藏时不启动 `requestAnimationFrame`，在重新可见时恢复。
- 为颜色解析增加不依赖 WebGL 的单元测试。

## 修改文件

- `website/assets/js/scene.js`
- `website/test/scene.test.js`
- `.superpowers/sdd/task-4-report.md`

## 验证

执行 `cd website && node --test test/scene.test.js test/theme.test.js test/i18n.test.js`，18 个测试全部通过。

## 剩余风险

未在 Node 环境测试实际 WebGL 材质构造；渲染结果仍依赖浏览器 WebGL、GPU 驱动及 CDN 可用性。
