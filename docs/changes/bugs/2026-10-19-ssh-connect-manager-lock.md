# SSH 连接卡住拖死其他会话与整应用

## 背景

连接异常服务器时，SSH 握手可能长时间阻塞。用户此时再连其他服务器会一直停在「连接中」，已有终端也无法输入，严重时整应用像卡死。

根因是 `SessionManager` 在持有全局互斥锁期间执行网络 I/O：`CreateSession` 在 `Lock` 内调用 `client.Connect()`，`closeSession` 在 `Lock` 内调用 Session/Client/SFTP 的 `Close()`。`WriteToSession` / `ListSessions` / 另一次 `CreateSession` 都依赖同一把锁，因此被一并堵住。

## 范围

- 将 dial、握手、SFTP 创建、本地 shell 创建、资源 Close 移出全局锁。
- 锁内仅做会话 map 的查重、登记与摘除。
- 补充并发回归测试，用可注入 dial / close hook 验证锁不跨网络 I/O。

不在此次范围：连接过程中的取消（关闭标签打断进行中的 dial）、前端超时提示文案。

## 修改文件

- `internal/ssh/manager.go`：`CreateSession`、`closeSession`、`GetOrCreateSFTPClient`、`CloseSFTPClient`、`CreateLocalSession` 调整锁范围；抽出 `ensureSessionAvailable` / `addSession` / `closeManagedResources`；`ManagedSession` 增加测试用 `resourceCloseHook`。
- `internal/ssh/client.go`：`sshDial` 可替换，便于模拟卡住的 dial。
- `internal/ssh/manager_connect_lock_test.go`：并发锁回归测试。
- `docs/changes/bugs/2026-10-19-ssh-connect-manager-lock.md`（本文）。

## 验证

```bash
go test ./internal/ssh/ -count=1 -timeout 60s
```

结果：通过。

## 剩余风险

- 进行中的 `ConnectSSH` 在登记进 map 之前，前端关闭标签无法打断 dial；仍会等到 dial 失败或超时（默认约 30s），但不再阻塞其他会话。
- `ssh.ClientConfig.Timeout` 只覆盖 TCP dial，TCP 已建立但握手无响应时仍可能较久；需要时可后续改为带 deadline 的自定义 dialer。
