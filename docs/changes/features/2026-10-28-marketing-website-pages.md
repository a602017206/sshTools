# 营销网站导航与页面

## 背景

营销网站已有主题、国际化和粒子场景基础模块，需要补齐可部署的导航脚本及首页、功能页和下载页。

## 范围

- 新增统一导航初始化，保存语言与主题选择、标记当前页面，并将调色板传递给场景。
- 新增三张静态页面，使用相对资源路径、双语文案和四个主题选择点。
- 下载页直接指向 GitHub 最新发布包，覆盖 README 中列出的桌面端与移动端文件。

## 修改文件

- `website/assets/js/nav.js`
- `website/index.html`
- `website/features.html`
- `website/download.html`
- `website/test/nav.test.js`
- `.superpowers/sdd/task-5-report.md`

## 验证

- `cd website && node --test test/*.test.js`
- 使用本地静态服务器检查首页、功能页、下载页、样式及导航脚本均返回成功。

## 剩余风险

- Node 测试覆盖导航状态和既有场景判定，未在无头环境验证 WebGL 粒子随鼠标移动的实际渲染效果。
