# 原生 NoSQL 计划顺序调整

## 背景

需求从三类原生服务扩展为常用 NoSQL 和 Kafka。原计划把最终端到端验证放在 provider 扩展之前，会导致完成记录早于实际范围完成。

## 范围

将最终验证和开发记录移动到内置 provider 注册表及全部首批 provider 之后，其他任务内容不改变。

## 修改文件

- `docs/plans/2026-07-13-native-nosql-and-jdbc-driver-removal-implementation.md`
- 本变更记录。

## 验证

已检查调整后的顺序为：界面接入、注册表、全部 provider、最终验证。

## 剩余风险

各 provider 的 SDK 和服务端兼容性仍需要在对应实现任务中逐项验证。
