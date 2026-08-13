# 修复用户名输入自动首字母大写

## 背景

部分系统输入法会把普通文本框按句首规则自动大写，导致输入 `root` 后变为 `Root`，进而造成 SSH 或数据库认证失败。

## 范围

仅调整新增和编辑连接弹窗中的用户名输入框，不改变已有连接配置或服务端保存逻辑。

## 修改文件

- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/connectionUsernameInput.test.js`
- `.github/workflows/release.yml`

## 验证

- 运行 `cd frontend && node --test test/connectionUsernameInput.test.js`，确认用户名输入框已关闭自动大写、自动更正和拼写检查。

## 剩余风险

不同操作系统的输入法实现存在差异；HTML 输入属性已明确要求保持原始大小写，应用仍会按用户实际输入的内容保存用户名。
