# 营销网站 three.js 场景模块

## 背景

营销网站需要一个可随主题切换、并尊重用户无障碍偏好的轻量粒子背景场景。

## 范围

- 新增 `shouldStartScene`，在用户启用减少动态效果或 WebGL 不可用时不启动场景。
- 新增 `mountScene`，通过固定版本的 three.js CDN 创建粒子、近邻连线和鼠标吸引效果。
- 场景支持主题调色板更新、页面可见性暂停/恢复及资源销毁；`light` 密度使用较少粒子。
- 新增仅覆盖 `shouldStartScene` 的 Node 单元测试，不在 Node 环境运行 WebGL。

## 修改文件

- `website/assets/js/scene.js`
- `website/test/scene.test.js`
- `.superpowers/sdd/task-4-report.md`

## 验证

- `cd website && node --test test/scene.test.js test/theme.test.js test/i18n.test.js`

## 剩余风险

- WebGL 渲染依赖浏览器 GPU 和远程 CDN；Node 单测仅验证启动判定，未执行浏览器 WebGL 渲染路径。
