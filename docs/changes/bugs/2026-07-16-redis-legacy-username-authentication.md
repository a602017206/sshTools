# Redis 历史用户名导致认证错误

## 背景

Redis 常见部署只启用 `requirepass`，不需要 ACL 用户名。旧版连接配置可能保存了默认用户名，例如 `Root`。原生连接入口继续转发该值时，Redis 客户端会改用带用户名的 ACL 认证，与仅密码服务端不兼容，并可能表现为连接测试超时。

## 范围

Redis 原生请求统一忽略用户名，仅传递密码；表单不再为 Redis 显示用户名字段。其他数据库的用户名认证行为不变。

## 修改文件

- `app.go`
- `app_native_database_test.go`
- `frontend/src/components/AddAssetDialog.svelte`
- 本变更记录。

## 验证

执行 `go test . -run TestNativeDatabaseRequestClearsLegacyRedisUsername -v`，验证旧 Redis 连接中的用户名会在请求构造时清空；执行前端构建验证表单编译。

## 剩余风险

本次按仅密码认证处理 Redis，不支持 Redis 6 及以上的自定义 ACL 用户名。此类实例需要后续单独增加 ACL 用户名配置和认证模式选择。
