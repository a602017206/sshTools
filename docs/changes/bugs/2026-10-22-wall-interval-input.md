# 修复状态墙探活间隔只能点步进、无法顺手输入

## 背景

全局设置「多资产状态墙」的探活间隔使用裸 `type=number` 输入框，用户反馈在应用内手输不方便，只能点右侧步进箭头逐秒增减（15 点到 18 需要多次点击）。

## 范围

- `GlobalSettingsDialog.svelte` 该项改三段式控件：顶部小数字框（可直接输入，失焦按 5–300 钳制）、中部滑杆（5–300 秒、步进 5，可拖动）、一排常用档位（5/15/30/60/120/300 秒一键选，当前档位高亮）。
- 保存口径不变：`asset_wall_probe_interval_seconds` 字段、默认 15、双侧钳制均沿用。

## 修改文件

- `frontend/src/components/GlobalSettingsDialog.svelte`
- `docs/changes/bugs/2026-10-22-wall-interval-input.md`（本文）

## 验证

- `vite build` 编译通过；前端单测无新增失败（改动仅模板控件，不动数据逻辑）。

## 剩余风险

- 无 GUI 环境未肉眼确认滑杆与档位渲染，需人工打开设置页看一眼。
