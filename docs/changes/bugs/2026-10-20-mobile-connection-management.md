# 移动端连接管理缺陷修复（删不掉、无密码框、改不了用户名、连接鉴权失败）

## 背景

Android 真机试用反馈：新增的连接删不掉、改不了用户名；新增表单没有密码框、密码不保存；点连接报 `ssh: unable to authenticate, attempted methods [none password]`。逐项溯源后确认是同一组契约缺陷叠加：

1. Flutter 新增连接时 `id` 传空字符串，Go 侧 `AddConnection` 也不生成 ID，连接以空 ID 落盘；删除/更新都走 `/connections/:id`，空 ID 无法命中路由，表现为「删不掉、改不了」。
2. Flutter 表单没有密码字段、模型无密码概念；服务端虽有 `SavePassword`（AES-GCM 加密存储），但 API 的新增/更新接口不接密码，密码从未保存。
3. 因密码从未保存，测试连接每次只能弹窗手输且不回存；用户在新增阶段无处填密码，最终以无有效凭据发起连接，SSH 握手在 password 阶段被拒。
4. 用户名等字段未关移动键盘的自动大写/纠错，`root` 可能被改成 `Root`，加剧「改不了/登不上」的感知。

## 范围

包含：连接的新增、编辑（改用户名等）、删除、密码保存与测试连接回退，及移动端连接列表/表单的可用性修正。不包含：SSH 终端页面、私钥文件在移动沙盒内的选取方式、后台保活。

## 修改文件

Go 侧：

- `internal/api/handlers/connection.go`：新增/更新请求体增加可选 `password`（不落 config.json，经 CredentialStore 加密保存）；新增时 ID 为空由服务端用 UUID 生成；响应增加 `has_password`（只回标记、永不回密码）；测试连接新增 `connection_id`，`auth_value` 为空时回退已保存密码/私钥路径，临时密码测试成功后回存；`auth_value` 不再是必填。
- `internal/service/connection_service.go`：新增 `EnsureConnectionIDs` 回填历史空 ID 并落盘；`RemoveConnection` 连带删除已保存密码（失败不阻断删除）。
- `mobile/mobile.go`、`cmd/apiserver/main.go`：服务启动时调用 `EnsureConnectionIDs`，已装机用户的存量空 ID 连接（如试用时加的那条）升级后自动获救。
- `internal/api/handlers/connection_test.go`（新增）：覆盖空 ID 新增自动生成 ID、密码保存且响应不含明文、无密码更新不丢密码、测试回退已存密码、删除连带删密码、历史数据回填后可删除。

Flutter 侧：

- `flutter_ui/lib/data/models/connection_model.dart`、`.g.dart`：新增 `hasPassword`（服务端下发、只读）与临时 `password`（只上行、不序列化回读，手改生成文件并注明）。
- `flutter_ui/lib/data/repositories/connection_repository.dart`：新增/更新时把表单密码拼进请求体；测试连接支持只带 `connection_id` 走服务端回退。
- `flutter_ui/lib/presentation/screens/connections/widgets/connection_form_page.dart`（新增，替代删除的 `connection_form_dialog.dart`）：全屏表单页，中文字段；密码框（显隐切换、新增必填、编辑时「已保存则留空不改」）；私钥路径；用户名/主机/密码关闭自动大写与纠错；保存按钮大触控区。
- `flutter_ui/lib/presentation/screens/connections/connection_list_screen.dart`：中文界面；卡片直接给出「测试 / 编辑 / 删除」按钮（不再全埋在右上角菜单）；卡片显示「密码已保存 / 未保存密码」状态；已存密码点测试直接测、不再反复弹窗；失败时展示服务端错误详情；删除确认文案注明密码一并删除。

## 验证

- 本机：`go build ./...`、`go vet ./internal/api/...` 通过；`go test ./internal/api/... ./internal/service/` 通过（新增 handler 测试全绿）；`cd mobile && go test ./...` 通过。
- GitHub 流水线第 5 轮（[39380246368](https://github.com/a602017206/sshTools/actions/runs/39380246368)）：Android APK 与 iOS 无签名构建双双成功，产物已更新。
- 仍需真机复测确认：旧版已保存的那条空 ID 连接在新版启动后应能正常编辑/删除；重新填入密码保存后，点「测试」应直接成功且不再弹窗。

## 剩余风险

- 私钥认证在移动端目前仅支持手填路径，沙盒文件访问受限，该方式在真机上可用性有限，后续应改为私钥内容导入并存入安全存储。
- 密码加密仍沿用机器特征派生密钥（未迁 Android Keystore），与桌面端实现一致，上架前应迁移。
- 终端页面未实现前，点卡片的动作为「测试连接」；SSH 会话终端仍是下一步工作，不要在用户预期上当作已完成。
