# Redis 仅密码认证连接失败

## 背景

Redis 常见部署只配置 `requirepass`，不配置 ACL 用户名。原生 Redis provider 已支持空用户名，但前端数据库表单把所有密码认证都要求填写用户名，导致 Redis 在保存或测试连接前被错误拦截。

## 范围

将 Redis 定义为仅密码认证类型，隐藏用户名输入，并在测试与保存时跳过用户名校验。其他原生数据库保持原有用户名校验。

## 修改文件

- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- 本变更记录。

## 验证

执行 `node --test test/nativeDatabaseTypes.test.js` 验证 Redis 不要求用户名；执行前端构建验证表单编译。

## 剩余风险

首版 Redis 表单不暴露 Redis 6 ACL 用户名。需要 ACL 用户名的 Redis 实例暂需使用后续扩展配置。
