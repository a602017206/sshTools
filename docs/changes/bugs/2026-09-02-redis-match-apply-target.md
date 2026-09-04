# 变更：Redis MATCH 智能填入与左侧键列表恢复

## 背景

CLI 页签曾隐藏左侧键列表；AI「填入」一律写入 CLI。用户查询「开头为 mini」一类需求应写入顶部 MATCH，并保持原始大小写；MATCH 输入还受系统自动首字母大写干扰。

## 范围

- CLI / 键详情均保留左侧键列表
- SCAN / 通配模式默认填入 MATCH 并触发扫描；GET 等命令仍填 CLI
- MATCH / CLI 输入关闭 autocapitalize / autocorrect / spellcheck

## 验证

- `node --test test/nativeCopilotApply.test.js`

## 剩余风险

- 模型若把浏览类 SCAN 标成其它命令名，仍可能落到 CLI；可用 `target: "match"|"cli"` 强制
