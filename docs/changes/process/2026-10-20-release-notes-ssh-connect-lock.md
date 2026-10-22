# 变更：发布说明补充 SSH 连接卡死修复

## 背景

SSH 连接异常服务器时会卡在连接中并拖住其他连接与已有终端的问题已修复（见 `docs/changes/bugs/2026-10-19-ssh-connect-manager-lock.md`）。自动发布工作流的更新说明需同步这项修复，便于用户从 Release 页面了解。

## 范围

- 仅更新 `.github/workflows/release.yml` 中 Release 正文「修复问题」列表，新增一条用户可见说明
- 不改构建矩阵、签名、打包与附件路径，不改发布触发条件

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-10-20-release-notes-ssh-connect-lock.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 使用 Ruby `YAML.load_file` 解析 `.github/workflows/release.yml`
- 执行 `git diff --check -- .github/workflows/release.yml docs/changes/process/2026-10-20-release-notes-ssh-connect-lock.md`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
