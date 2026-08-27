# 原生数据平台只读详情实现记录

## 实现概述

服务层通过 `NativeResourceDetails` 统一表达资源详情，其中 `Content` 为 JSON 字符串，避免原生协议的数据结构泄漏至 Wails 与 Svelte UI。`DescribeNativeDatabaseResource` 复用现有会话，并使用十秒超时。

## 行为限制

- Redis：字符串值会限制到 4096 字节；其他键类型不读取完整值。
- Elasticsearch：`_search` 请求固定 `size=20`，仅用作预览。
- Kafka：只执行 Metadata 请求；不创建消费者、不拉取消息也不生产消息。

## 兼容性

其它原生 provider 暂未实现详情接口时会返回明确的“暂不支持资源详情”错误，既有资源树功能保持不变。
