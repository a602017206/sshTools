# 变更：Redis CLI 命令文本域布局修复

## 背景

CLI 文本域曾被挤成竖条；修复后仍与左侧键列表并排，输入区过高，错误提示挤在左侧窄栏。AI 产出 `{"command":["SCAN",...]}` 时被 `String(array)` 变成带逗号命令，执行失败。

## 范围

- CLI 页签改为全宽控制台：隐藏键列表，紧凑命令框 + 下方结果区
- 状态/错误提示移到 CLI 内容区
- 规范化：`command` 数组、逗号拼接残留、SCAN 对象字段

## 验证

- `node --test test/nativeCopilotApply.test.js`

## 剩余风险

- 其它结构化 Redis 命令若字段命名不一致，仍可能需手工改命令行
