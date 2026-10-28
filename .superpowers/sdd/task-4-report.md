# Task 4：three.js 场景模块报告

## 完成内容

- 新增 `website/assets/js/scene.js`，固定使用 `three@0.170.0` 的 jsDelivr ESM 地址。
- 提供 `shouldStartScene`：在 `prefers-reduced-motion` 或 WebGL 不可用时返回 `false`。
- 提供 `mountScene`：创建粒子场、近邻连线、鼠标吸引效果，并支持主题色更新、可见性暂停/恢复和资源释放。
- 新增 Node 单元测试，仅覆盖 `shouldStartScene`，没有在 Node 中执行 WebGL。

## 验证

执行：

```bash
cd website && node --test test/scene.test.js test/theme.test.js test/i18n.test.js
```

结果：16 个测试全部通过。

## 风险

- 渲染路径依赖浏览器 WebGL 支持、GPU 驱动和 jsDelivr CDN 可用性；本任务按约束未在 Node 环境执行 `mountScene`。
# Task 4 报告：DockerService 镜像/网络/卷方法

- 新增 `internal/service/docker_resources.go`：List/Inspect（镜像、网络、卷）、`RemoveDockerImage(sessionID, ref, force)`、`RemoveDockerNetwork`、`RemoveDockerVolume`（force 固定 false）。
- 通用泛型 helper `dockerResourceCall`：`lookupSession` → API（context 超时）或 CLI（`cliSession` 超时）。list/inspect 用 `dockerOpTimeout`，remove 用 `dockerLifecycleTimeout`。CLI inspect 补 `Via:"cli"`（list 由 parser 设置）。
- 校验：镜像允许 `: / @`，网络/卷非空、不以 `-` 开头、仅 `[A-Za-z0-9_.-]`；非法输入不会到达 CLI。未知会话先报错。
- 测试 `docker_resources_test.go`：API httptest mock（含路径 `registry.io/team/app:1.0`、`?force=1`、卷删除无 force、409 错误）、fake CLI 命令断言、校验、未知会话。
- 验证：`go test ./internal/service -run TestDocker` PASS；gofmt/vet 干净。
- 文档：`docs/changes/features/2026-10-28-docker-resources-service.md`。
- 未做：`app.go` 绑定、前端。

## 场景模块审查修复（2026-10-28）

### 修复内容

- `website/assets/js/scene.js` 新增纯函数 `parseCssColor`，将 `#RRGGBBAA` 拆分为 three@0.170.0 可识别的 `#RRGGBB` 和 0–1 的不透明度。
- 初始创建和 `setPalette()` 更新连线材质时均应用拆分后的颜色、不透明度与 `transparent`，避免 three@0.170.0 将 8 位十六进制颜色解析失败。
- 场景挂载时若 `document.visibilityState === 'hidden'`，保持暂停状态且不申请 RAF；页面重新可见后由 `visibilitychange` 恢复。
- `website/test/scene.test.js` 新增 8 位和 6 位十六进制颜色解析的纯单元测试，无需 WebGL。

### 验证

执行：

```bash
cd website && node --test test/scene.test.js test/theme.test.js test/i18n.test.js
```

结果：18 个测试全部通过。
