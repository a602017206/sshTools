# JDBC 驱动管理设计文档

## 背景

数据库模块需要一份设计文档，用于规划通过全量 JDBC agent 和驱动管理系统扩展数据库支持。brainstorming 流程会把确认后的设计保存到 `docs/superpowers/specs/`，该目录此前没有记录在文档地图中。

用户明确要求后续文档不要再出现全英文内容，因此本次同步增加仓库级文档语言约束。

## 范围

- 新增 JDBC 驱动管理设计文档。
- 在 `docs/README.md` 中记录 `docs/superpowers/specs/` 文档类别。
- 忽略本地 `.superpowers/` 视觉伴随运行文件。
- 在 `AGENTS.md` 增加文档语言要求：新增和修改的文档正文必须使用中文。
- 将 `AGENTS.md` 当前正文同步改为中文，仅保留必要技术标识原文。
- 将本次新增和触碰的文档正文改为中文。

## 修改文件

- `.gitignore`
- `AGENTS.md`
- `docs/README.md`
- `docs/superpowers/specs/2026-07-09-jdbc-driver-management-design.md`
- `docs/changes/process/2026-07-09-jdbc-driver-management-spec.md`

## 验证

- 检查设计文档没有待填占位、互相矛盾的选择、范围漂移和明显歧义。
- 确认设计文档只描述架构和验收，不进入实现代码。
- 确认文档语言要求已经写入 `AGENTS.md`。

## 剩余风险

- 这是设计文档，不包含实现；Java agent 打包、gRPC 生成、厂商驱动 license 等细节仍需要后续技术验证。
