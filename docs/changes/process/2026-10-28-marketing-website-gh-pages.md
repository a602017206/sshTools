# 变更：营销宣传站 GitHub Pages 部署

## 背景

为 sshTools（AHaSSHTools）提供独立于桌面/移动安装包的产品介绍站点，托管于 GitHub Pages，便于在 README 与对外宣传中引用固定 URL。

## 范围

- 新增 `.github/workflows/gh-pages.yml`，使用官方 `configure-pages` / `upload-pages-artifact` / `deploy-pages` 流程部署 `website/` 静态目录。
- 在根目录 `README.md` 增加宣传站链接。
- 同步本特性相关设计、计划与实现说明文档（`docs/designs/`、`docs/plans/`、`docs/superpowers/`、`docs/changes/process/`）。
- **未修改** `release.yml`、`mobile.yml` 及现有发布/移动端 CI。

## 修改文件

| 路径 | 说明 |
| --- | --- |
| `.github/workflows/gh-pages.yml` | 新增：push `master` 且 `website/**` 或本 workflow 变更时部署；支持 `workflow_dispatch` |
| `README.md` | 新增「宣传网站」小节与 Pages URL |
| `docs/changes/process/2026-10-28-marketing-website-gh-pages.md` | 本变更记录 |
| `docs/changes/process/2026-10-28-marketing-website-gh-pages-plan.md` | 计划流程变更记录 |
| `docs/designs/2026-10-28-marketing-website-gh-pages-design.md` | 设计文档 |
| `docs/plans/2026-10-28-marketing-website-gh-pages.md` | 实现计划（与 superpowers 副本同步） |
| `docs/superpowers/plans/2026-10-28-marketing-website-gh-pages.md` | Superpowers 计划源 |
| `docs/superpowers/specs/2026-10-28-marketing-website-gh-pages-design.md` | Superpowers 规格 |

## 验证

- 本地：`cd website && node --test test/*.test.js` — 全部通过。
- 合并至 `master` 后：在仓库 **Settings → Pages → Build and deployment → Source** 选择 **GitHub Actions**（若尚未配置，首次部署前需人工确认此项）。
- 推送 `website/**` 或 workflow 变更后，检查 Actions 中「Deploy GitHub Pages」工作流成功，并访问 https://a602017206.github.io/sshTools/ 确认三页与资源加载正常。

## 剩余风险

- Pages 源未设为 GitHub Actions 时，workflow 可能无法发布或 URL 不可用。
- 站点使用仓库子路径 `/sshTools/`；若 Pages 自定义域名或 base path 变更，需同步调整 `website` 内相对路径与链接。
- `master` 分支命名与默认分支不一致时，需调整 workflow 的 `branches` 触发条件。
