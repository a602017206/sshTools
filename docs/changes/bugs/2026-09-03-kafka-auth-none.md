# 变更：Kafka 支持无认证连接

## 背景

Kafka Broker 常无密码；原表单把密码标为必填，且连接流程不允许空密码，导致无法测试/保存/打开。参考 DataGrip 的 Authentication: None。

## 范围

- Kafka 增加认证「无 / 密码」；无认证时不展示也不校验用户名密码
- 连接打开时对 Kafka / `auth_type=none` 允许空密码，不再弹必填密码框
- `databaseTypeRequiresPassword` / `connectionAllowsEmptyPassword` 辅助判断

## 验证

- `node --test test/nativeDatabaseTypes.test.js`

## 剩余风险

- 当前后端 Kafka 客户端仍未接 SASL；选「密码」只会把凭据传到现有连接参数，尚未真正启用 SASL 握手
