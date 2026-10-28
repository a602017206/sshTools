# 变更：宣传站与 gh-pages 实现计划

## 背景

已确认宣传站设计（纯静态多页、中英、四主题、three.js 氛围层、Actions 部署 Pages）。进入实现前需落盘分步计划，供执行与评审。

## 范围

- 新增实现计划与设计规格（此前已写入 designs/specs）的计划副本
- 不包含站点代码与 workflow 落地（后续按计划执行）

## 修改文件

- `docs/superpowers/plans/2026-10-28-marketing-website-gh-pages.md`（新增）
- `docs/plans/2026-10-28-marketing-website-gh-pages.md`（同步副本）
- `docs/designs/2026-10-28-marketing-website-gh-pages-design.md`（设计，含 three.js / 多主题）
- `docs/superpowers/specs/2026-10-28-marketing-website-gh-pages-design.md`（规格副本）
- `docs/changes/process/2026-10-28-marketing-website-gh-pages-plan.md`（本文件）

## 验证

- 计划任务覆盖规格中的页面、i18n、主题、scene、CI、README、变更文档
- 计划内主题/i18n/scene 门控函数含 `node --test` 步骤

## 剩余风险

- 计划尚未执行；Pages Source 需在实现合并后人工配置
- 大图未压缩风险记在实现任务 Task 3
