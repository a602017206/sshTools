# JDBC 驱动管理 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 把现有 Go `database/sql` 数据库模块迁移为全部 JDBC agent 架构，并提供 JRE/驱动安装、离线导入、连接、查询、元数据浏览的首版闭环。

**架构：** Go/Wails 应用保留现有前端 API 入口，新增 `JdbcGatewayService`、驱动目录、运行时管理和 agent 进程管理。Java agent 作为本地子进程运行，通过绑定到 `127.0.0.1` 的 gRPC 接收连接、查询、元数据和关闭请求。驱动以 profile + jar 文件形式存储在 `~/.sshtools/drivers`，每个 profile 在 Java agent 内使用隔离 classloader 加载。

**技术栈：** Go、Wails、Svelte、Java 21、Gradle、gRPC、Protocol Buffers、JDBC、H2/SQLite 测试驱动、Go `testing`。

---

## 实施约束

- 每个代码、配置、构建或文档变更都要有对应 `docs/changes/` 记录。
- 新增或修改的文档正文必须使用中文；技术标识、命令、API 名、文件路径可保留原文。
- 优先保持现有 Wails 数据库 API 名称稳定：`ConnectDatabase`、`ExecuteDatabaseQuery`、`ListDatabaseTables`、`ListDatabaseDatabases`、`GetDatabaseTableSchema`、`CloseDatabase`。
- 首版只实现连接、查询、列表、字段元数据和驱动/JRE 管理；不做数据编辑、导入导出、ER 图。
- 先用 H2 JDBC 作为集成测试驱动，避免依赖外部数据库服务。

## 任务 1：建立 JDBC 元数据模型和本地目录约定

**文件：**
- 创建：`internal/config/jdbc.go`
- 创建：`internal/service/jdbc_paths.go`
- 创建：`internal/service/jdbc_catalog.go`
- 创建：`internal/service/jdbc_catalog_test.go`
- 修改：`internal/config/database.go`
- 创建：`docs/changes/features/2026-07-09-jdbc-metadata-models.md`

**步骤 1：写失败测试**

在 `internal/service/jdbc_catalog_test.go` 写：

```go
package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDriverCatalogLoadsManifestAndSelectsRecommendedProfile(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "manifest.json")
	err := os.WriteFile(manifestPath, []byte(`{
	  "version": 1,
	  "drivers": [{
	    "id": "oracle",
	    "name": "Oracle",
	    "recommendedVersion": "23.5",
	    "profiles": [{
	      "id": "oracle-23.5",
	      "version": "23.5",
	      "driverClass": "oracle.jdbc.OracleDriver",
	      "urlTemplate": "jdbc:oracle:thin:@//{host}:{port}/{database}",
	      "defaultPort": 1521,
	      "jre": ">=17",
	      "jars": [{"name": "ojdbc11.jar", "sha256": "abc"}]
	    }]
	  }]
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	catalog := NewDriverCatalogService(manifestPath, "")
	driver, profile, err := catalog.GetRecommendedProfile("oracle")
	if err != nil {
		t.Fatalf("expected profile, got error: %v", err)
	}
	if driver.Name != "Oracle" {
		t.Fatalf("expected Oracle, got %q", driver.Name)
	}
	if profile.DriverClass != "oracle.jdbc.OracleDriver" {
		t.Fatalf("unexpected driver class: %s", profile.DriverClass)
	}
	if profile.DefaultPort != 1521 {
		t.Fatalf("unexpected default port: %d", profile.DefaultPort)
	}
}
```

**步骤 2：运行测试确认失败**

运行：

```bash
go test ./internal/service -run TestDriverCatalogLoadsManifestAndSelectsRecommendedProfile -v
```

预期：失败，提示 `NewDriverCatalogService` 或相关类型未定义。

**步骤 3：实现最小模型**

在 `internal/config/jdbc.go` 写：

```go
package config

type JDBCManifest struct {
	Version int          `json:"version"`
	Drivers []JDBCDriver `json:"drivers"`
}

type JDBCDriver struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	RecommendedVersion string              `json:"recommendedVersion"`
	Profiles           []JDBCDriverProfile `json:"profiles"`
}

type JDBCDriverProfile struct {
	ID             string    `json:"id"`
	Version        string    `json:"version"`
	DriverClass    string    `json:"driverClass"`
	URLTemplate     string    `json:"urlTemplate"`
	DefaultPort     int       `json:"defaultPort"`
	JRERequirement  string    `json:"jre"`
	Jars           []JDBCJar  `json:"jars"`
	Properties     []JDBCProp `json:"properties,omitempty"`
	Source          string    `json:"source,omitempty"`
	Installed       bool      `json:"installed,omitempty"`
	InstallPath     string    `json:"installPath,omitempty"`
}

type JDBCJar struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	URL    string `json:"url,omitempty"`
}

type JDBCProp struct {
	Name         string `json:"name"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Required     bool   `json:"required,omitempty"`
}
```

在 `internal/service/jdbc_paths.go` 写路径常量和构造函数：

```go
package service

import "path/filepath"

type JDBCPaths struct {
	RootDir     string
	DriversDir  string
	RuntimesDir string
	AgentDir    string
	LogsDir     string
	Manifest    string
}

func NewJDBCPaths(root string) JDBCPaths {
	return JDBCPaths{
		RootDir:     root,
		DriversDir:  filepath.Join(root, "drivers"),
		RuntimesDir: filepath.Join(root, "runtimes"),
		AgentDir:    filepath.Join(root, "agent"),
		LogsDir:     filepath.Join(root, "logs"),
		Manifest:    filepath.Join(root, "drivers", "manifest.json"),
	}
}
```

在 `internal/service/jdbc_catalog.go` 写：

```go
package service

import (
	"encoding/json"
	"fmt"
	"os"

	"sshTools/internal/config"
)

type DriverCatalogService struct {
	manifestPath  string
	installedPath string
}

func NewDriverCatalogService(manifestPath, installedPath string) *DriverCatalogService {
	return &DriverCatalogService{manifestPath: manifestPath, installedPath: installedPath}
}

func (s *DriverCatalogService) LoadManifest() (*config.JDBCManifest, error) {
	data, err := os.ReadFile(s.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取 JDBC 驱动清单失败: %w", err)
	}
	var manifest config.JDBCManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("解析 JDBC 驱动清单失败: %w", err)
	}
	return &manifest, nil
}

func (s *DriverCatalogService) GetRecommendedProfile(driverID string) (*config.JDBCDriver, *config.JDBCDriverProfile, error) {
	manifest, err := s.LoadManifest()
	if err != nil {
		return nil, nil, err
	}
	for i := range manifest.Drivers {
		driver := &manifest.Drivers[i]
		if driver.ID != driverID {
			continue
		}
		for j := range driver.Profiles {
			profile := &driver.Profiles[j]
			if profile.Version == driver.RecommendedVersion || profile.ID == driver.RecommendedVersion {
				return driver, profile, nil
			}
		}
		if len(driver.Profiles) > 0 {
			return driver, &driver.Profiles[0], nil
		}
		return nil, nil, fmt.Errorf("数据库 %s 没有可用 JDBC profile", driverID)
	}
	return nil, nil, fmt.Errorf("未找到 JDBC 驱动: %s", driverID)
}
```

更新 `internal/config/database.go`：

```go
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBType          string
	Database        string
	Timeout         time.Duration
	DriverProfileID string
	Properties      map[string]string
}
```

**步骤 4：运行测试确认通过**

运行：

```bash
go test ./internal/service -run TestDriverCatalogLoadsManifestAndSelectsRecommendedProfile -v
```

预期：PASS。

**步骤 5：写变更记录**

创建 `docs/changes/features/2026-07-09-jdbc-metadata-models.md`，包含背景、范围、修改文件、验证、剩余风险。

**步骤 6：提交**

```bash
git add internal/config/jdbc.go internal/config/database.go internal/service/jdbc_paths.go internal/service/jdbc_catalog.go internal/service/jdbc_catalog_test.go docs/changes/features/2026-07-09-jdbc-metadata-models.md
git commit -m "feat: add jdbc driver metadata models"
```

## 任务 2：实现 checksum 校验和离线驱动包导入

**文件：**
- 创建：`internal/service/jdbc_install.go`
- 创建：`internal/service/jdbc_install_test.go`
- 创建：`docs/changes/features/2026-07-09-jdbc-offline-driver-import.md`

**步骤 1：写失败测试**

在 `internal/service/jdbc_install_test.go` 写：

```go
package service

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestDriverInstallImportsOfflinePackageAndValidatesChecksum(t *testing.T) {
	root := t.TempDir()
	jarBytes := []byte("fake-h2-driver")
	sum := sha256.Sum256(jarBytes)
	zipPath := filepath.Join(root, "driver-package.zip")
	createTestDriverPackage(t, zipPath, jarBytes, hex.EncodeToString(sum[:]))

	installer := NewDriverInstallService(NewJDBCPaths(filepath.Join(root, ".sshtools")))
	result, err := installer.ImportOfflinePackage(zipPath)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if result.ProfileID != "h2-2.2.224" {
		t.Fatalf("unexpected profile id: %s", result.ProfileID)
	}
	if _, err := os.Stat(filepath.Join(result.InstallPath, "jars", "h2.jar")); err != nil {
		t.Fatalf("jar not installed: %v", err)
	}
}

func createTestDriverPackage(t *testing.T, zipPath string, jarBytes []byte, sha string) {
	t.Helper()
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	files := map[string]string{
		"package.json": `{"id":"h2","name":"H2","version":"2.2.224","driverClass":"org.h2.Driver","urlTemplate":"jdbc:h2:mem:{database}","defaultPort":0,"jre":">=17","jars":["h2.jar"]}`,
		"checksums.sha256": sha + "  jars/h2.jar\n",
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	w, err := zw.Create("jars/h2.jar")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(jarBytes)
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestDriverInstallImportsOfflinePackageAndValidatesChecksum -v
```

预期：失败，提示 `NewDriverInstallService` 未定义。

**步骤 3：实现最小导入服务**

在 `internal/service/jdbc_install.go` 实现：

- 打开 zip。
- 读取 `package.json`。
- 读取 `checksums.sha256`。
- 校验每个 jar 的 sha256。
- 解压到 `paths.DriversDir/<id>/<version>/`。
- 写入 `driver.json`。
- 任一步失败时删除目标临时目录。

核心结构：

```go
type DriverInstallService struct {
	paths JDBCPaths
}

type DriverInstallResult struct {
	DriverID    string
	ProfileID   string
	Version     string
	InstallPath string
}
```

导入结果 `ProfileID` 规则：`<id>-<version>`。

**步骤 4：运行测试确认通过**

```bash
go test ./internal/service -run TestDriverInstallImportsOfflinePackageAndValidatesChecksum -v
```

预期：PASS。

**步骤 5：增加失败回滚测试**

新增测试：checksum 错误时返回错误，且目标目录不存在。

运行：

```bash
go test ./internal/service -run 'TestDriverInstallImportsOfflinePackageAndValidatesChecksum|TestDriverInstallRollsBackOnChecksumMismatch' -v
```

预期：PASS。

**步骤 6：写变更记录并提交**

```bash
git add internal/service/jdbc_install.go internal/service/jdbc_install_test.go docs/changes/features/2026-07-09-jdbc-offline-driver-import.md
git commit -m "feat: import offline jdbc driver packages"
```

## 任务 3：实现 JRE 运行时选择模型

**文件：**
- 创建：`internal/service/jdbc_runtime.go`
- 创建：`internal/service/jdbc_runtime_test.go`
- 创建：`docs/changes/features/2026-07-09-jdbc-runtime-selection.md`

**步骤 1：写失败测试**

```go
func TestRuntimeServicePrefersManagedRuntimeThenSystemRuntime(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	managedJava := filepath.Join(paths.RuntimesDir, "jre-21-test", "bin", "java")
	if err := os.MkdirAll(filepath.Dir(managedJava), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managedJava, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	runtimeSvc := NewRuntimeService(paths, "/usr/bin/java")
	selected, err := runtimeSvc.SelectRuntime()
	if err != nil {
		t.Fatalf("select runtime failed: %v", err)
	}
	if selected.Kind != RuntimeKindManaged {
		t.Fatalf("expected managed runtime, got %s", selected.Kind)
	}
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestRuntimeServicePrefersManagedRuntimeThenSystemRuntime -v
```

预期：失败，类型未定义。

**步骤 3：实现最小 RuntimeService**

实现：

- `RuntimeKindManaged`
- `RuntimeKindSystem`
- `RuntimeKindMissing`
- `RuntimeService.SelectRuntime()`
- `RuntimeService.ImportRuntimeArchive()` 先只定义接口，不实现下载。

选择规则：

1. 如果用户显式选择系统 Java 模式，并且路径有效，使用系统 Java。
2. 否则优先使用最新托管 JRE。
3. 如果两者都不可用，返回缺失状态。

这个规则满足“默认托管 JRE”，同时保留高级用户手动选择系统 Java 的能力。

**步骤 4：运行测试确认通过**

```bash
go test ./internal/service -run TestRuntimeService -v
```

预期：PASS。

**步骤 5：写变更记录并提交**

```bash
git add internal/service/jdbc_runtime.go internal/service/jdbc_runtime_test.go docs/changes/features/2026-07-09-jdbc-runtime-selection.md
git commit -m "feat: select jdbc runtime"
```

## 任务 4：建立 Java agent 工程和 gRPC 协议

**文件：**
- 创建：`jdbc-agent/settings.gradle`
- 创建：`jdbc-agent/build.gradle`
- 创建：`jdbc-agent/src/main/proto/jdbc_agent.proto`
- 创建：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/JdbcAgentApplication.java`
- 创建：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/HealthServiceImpl.java`
- 创建：`jdbc-agent/src/test/java/com/sshtools/jdbcagent/HealthServiceImplTest.java`
- 修改：`.gitignore`，忽略 `jdbc-agent/build/`
- 创建：`docs/changes/features/2026-07-09-jdbc-agent-bootstrap.md`

**步骤 1：写 proto**

`jdbc-agent/src/main/proto/jdbc_agent.proto`：

```proto
syntax = "proto3";

package sshtools.jdbc;

option java_multiple_files = true;
option java_package = "com.sshtools.jdbcagent.proto";
option java_outer_classname = "JdbcAgentProto";

service JdbcAgent {
  rpc Health(HealthRequest) returns (HealthResponse);
  rpc OpenSession(OpenSessionRequest) returns (OpenSessionResponse);
  rpc ExecuteQuery(ExecuteQueryRequest) returns (QueryResult);
  rpc ListTables(ListTablesRequest) returns (ListTablesResponse);
  rpc ListColumns(ListColumnsRequest) returns (ListColumnsResponse);
  rpc CloseSession(CloseSessionRequest) returns (CloseSessionResponse);
}

message HealthRequest {
  string token = 1;
}

message HealthResponse {
  string status = 1;
  string agent_version = 2;
  string java_version = 3;
}

message DriverProfile {
  string id = 1;
  string driver_class = 2;
  string url_template = 3;
  repeated string jar_paths = 4;
}

message OpenSessionRequest {
  string token = 1;
  string session_id = 2;
  DriverProfile profile = 3;
  string host = 4;
  int32 port = 5;
  string database = 6;
  string user = 7;
  string password = 8;
  map<string, string> properties = 9;
}

message OpenSessionResponse {
  string session_id = 1;
}

message ExecuteQueryRequest {
  string token = 1;
  string session_id = 2;
  string sql = 3;
  int32 timeout_seconds = 4;
}

message QueryResult {
  repeated string columns = 1;
  repeated Row rows = 2;
  int64 affected = 3;
}

message Row {
  repeated string values = 1;
}

message ListTablesRequest {
  string token = 1;
  string session_id = 2;
  string catalog = 3;
  string schema = 4;
}

message ListTablesResponse {
  repeated string tables = 1;
}

message ListColumnsRequest {
  string token = 1;
  string session_id = 2;
  string catalog = 3;
  string schema = 4;
  string table = 5;
}

message Column {
  string name = 1;
  string type = 2;
  bool nullable = 3;
  bool primary_key = 4;
}

message ListColumnsResponse {
  repeated Column columns = 1;
}

message CloseSessionRequest {
  string token = 1;
  string session_id = 2;
}

message CloseSessionResponse {
  bool closed = 1;
}
```

**步骤 2：写失败测试**

`HealthServiceImplTest` 验证 token 正确返回 `OK`，token 错误抛 `UNAUTHENTICATED`。

**步骤 3：运行测试确认失败**

```bash
cd jdbc-agent && ./gradlew test --tests '*HealthServiceImplTest'
```

预期：失败，服务实现未完成。

**步骤 4：实现最小 agent**

- `JdbcAgentApplication` 读取参数：`--port`、`--token`。
- 启动 gRPC server。
- `HealthServiceImpl` 校验 token 并返回 `OK`、agent version、Java version。

**步骤 5：运行测试确认通过**

```bash
cd jdbc-agent && ./gradlew test --tests '*HealthServiceImplTest'
```

预期：PASS。

**步骤 6：写变更记录并提交**

```bash
git add .gitignore jdbc-agent docs/changes/features/2026-07-09-jdbc-agent-bootstrap.md
git commit -m "feat: bootstrap jdbc agent"
```

## 任务 5：实现 Java agent 的 DriverLoader 和 H2 查询闭环

**文件：**
- 创建：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/DriverLoader.java`
- 创建：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/ConnectionRegistry.java`
- 创建：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/QueryServiceImpl.java`
- 创建：`jdbc-agent/src/test/java/com/sshtools/jdbcagent/QueryServiceImplTest.java`
- 修改：`jdbc-agent/build.gradle`
- 创建：`docs/changes/features/2026-07-09-jdbc-agent-query.md`

**步骤 1：写失败测试**

测试用 H2 JDBC：

```java
@Test
void executeQueryReturnsRows() {
    // 使用 H2 jar 的 driver class 和 jdbc:h2:mem:test URL
    // OpenSession 后执行 "select 1 as id, 'ok' as name"
    // 断言 columns = ["ID", "NAME"]，第一行 values = ["1", "ok"]
}
```

**步骤 2：运行测试确认失败**

```bash
cd jdbc-agent && ./gradlew test --tests '*QueryServiceImplTest'
```

预期：失败，`QueryServiceImpl` 未定义或未实现。

**步骤 3：实现最小查询**

- `DriverLoader` 用 `URLClassLoader` 加载 jar。
- `ConnectionRegistry` 保存 `Map<String, Connection>`。
- `OpenSession` 渲染 URL template，创建 connection。
- `ExecuteQuery` 对 `ResultSet` 返回 rows；对 update 返回 affected。
- 所有返回值先转为 string，避免 gRPC 动态类型复杂度。

**步骤 4：运行测试确认通过**

```bash
cd jdbc-agent && ./gradlew test --tests '*QueryServiceImplTest'
```

预期：PASS。

**步骤 5：提交**

```bash
git add jdbc-agent docs/changes/features/2026-07-09-jdbc-agent-query.md
git commit -m "feat: execute jdbc queries in agent"
```

## 任务 6：实现 Java agent 元数据和关闭会话

**文件：**
- 创建：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- 创建：`jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`
- 修改：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/ConnectionRegistry.java`
- 创建：`docs/changes/features/2026-07-09-jdbc-agent-metadata.md`

**步骤 1：写失败测试**

测试创建 H2 表：

```sql
create table users (id int primary key, name varchar(32))
```

断言：

- `ListTables` 返回 `USERS`。
- `ListColumns` 返回 `ID` 和 `NAME`。
- `CloseSession` 后再查询返回 NOT_FOUND。

**步骤 2：运行测试确认失败**

```bash
cd jdbc-agent && ./gradlew test --tests '*MetadataServiceImplTest'
```

预期：失败。

**步骤 3：实现元数据服务**

- 使用 `DatabaseMetaData.getTables`。
- 使用 `DatabaseMetaData.getColumns`。
- 主键用 `DatabaseMetaData.getPrimaryKeys` 合并。
- `CloseSession` 从 registry 删除并关闭连接。

**步骤 4：运行测试确认通过**

```bash
cd jdbc-agent && ./gradlew test --tests '*MetadataServiceImplTest'
```

预期：PASS。

**步骤 5：提交**

```bash
git add jdbc-agent docs/changes/features/2026-07-09-jdbc-agent-metadata.md
git commit -m "feat: expose jdbc metadata from agent"
```

## 任务 7：实现 Go AgentProcessManager 和 gRPC client

**文件：**
- 创建：`internal/service/jdbc_agent_process.go`
- 创建：`internal/service/jdbc_agent_process_test.go`
- 创建：`internal/service/jdbc_gateway.go`
- 创建：`internal/service/jdbc_gateway_test.go`
- 创建：`internal/service/jdbcproto/`（由 proto 生成）
- 修改：`go.mod`
- 创建：`docs/changes/features/2026-07-09-jdbc-agent-gateway.md`

**步骤 1：写失败测试**

`jdbc_agent_process_test.go` 用 fake command runner：

```go
func TestAgentProcessManagerStartsAgentWithLocalPortAndToken(t *testing.T) {
	runner := &fakeCommandRunner{}
	manager := NewAgentProcessManager(runner, AgentProcessConfig{JavaPath: "/bin/java", AgentJar: "/tmp/jdbc-agent.jar"})
	handle, err := manager.Start(context.Background())
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if handle.Port == 0 || handle.Token == "" {
		t.Fatalf("expected port and token")
	}
	if !strings.Contains(strings.Join(runner.args, " "), "--token") {
		t.Fatalf("token argument missing")
	}
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestAgentProcessManagerStartsAgentWithLocalPortAndToken -v
```

预期：失败。

**步骤 3：实现最小进程管理**

- 随机选择本地端口。
- 生成 32 字节 token。
- 用 `exec.CommandContext(java, "-jar", agentJar, "--port", port, "--token", token)` 启动。
- 暴露 `Start`、`Stop`、`Health`。
- 测试中通过接口替代真实 `exec.Command`。

**步骤 4：生成 Go gRPC client**

需要加入生成命令：

```bash
protoc --go_out=internal/service/jdbcproto --go-grpc_out=internal/service/jdbcproto jdbc-agent/src/main/proto/jdbc_agent.proto
```

如果仓库没有 protoc 工具链，先记录为构建前置，并在计划执行时补脚本 `scripts/generate-jdbc-proto.sh`。

**步骤 5：写 JdbcGatewayService 测试**

使用 fake gRPC client，断言：

- `ConnectDatabase` 会渲染 profile 并调用 `OpenSession`。
- `ExecuteQuery` 把 agent rows 转成现有 `QueryResult`。
- agent `DRIVER_MISSING` 映射为 Go 错误分类。

**步骤 6：运行测试确认通过**

```bash
go test ./internal/service -run 'TestAgentProcessManager|TestJdbcGateway' -v
```

预期：PASS。

**步骤 7：提交**

```bash
git add go.mod go.sum internal/service/jdbc_agent_process.go internal/service/jdbc_agent_process_test.go internal/service/jdbc_gateway.go internal/service/jdbc_gateway_test.go internal/service/jdbcproto docs/changes/features/2026-07-09-jdbc-agent-gateway.md
git commit -m "feat: add jdbc agent gateway"
```

## 任务 8：把现有 DatabaseService 切到 JdbcGatewayService

**文件：**
- 修改：`internal/service/database_service.go`
- 修改：`internal/service/database_service_test.go`
- 修改：`app.go`
- 修改：`internal/service/database_drivers.go`
- 创建：`docs/changes/features/2026-07-09-database-service-jdbc-gateway.md`

**步骤 1：写失败测试**

在 `database_service_test.go` 新增 fake gateway：

```go
func TestDatabaseServiceDelegatesConnectToJdbcGateway(t *testing.T) {
	gateway := &fakeJdbcGateway{}
	ds := NewDatabaseServiceWithGateway(nil, gateway)
	err := ds.ConnectDatabase("db-test", "localhost", 1521, "scott", "tiger", "oracle", "orcl")
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if gateway.lastDBType != "oracle" {
		t.Fatalf("expected oracle, got %s", gateway.lastDBType)
	}
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestDatabaseServiceDelegatesConnectToJdbcGateway -v
```

预期：失败，构造函数未定义。

**步骤 3：实现迁移适配**

- 给 `DatabaseService` 增加 `gateway DatabaseGateway` 字段。
- 定义接口：

```go
type DatabaseGateway interface {
	ConnectDatabase(ctx context.Context, sessionID string, cfg config.DatabaseConfig) error
	ExecuteQuery(ctx context.Context, sessionID string, query string) (*QueryResult, error)
	ListTables(ctx context.Context, sessionID, database string) ([]string, error)
	ListDatabases(ctx context.Context, sessionID string) ([]string, error)
	GetTableSchema(ctx context.Context, sessionID, table string) (*config.TableSchema, error)
	CloseDatabase(ctx context.Context, sessionID string) error
}
```

- 现有 public 方法保持不变，内部转发给 gateway。
- 移除直接依赖 MySQL/PostgreSQL 的 DSN 分支，或保留到 legacy 文件但不被调用。
- `app.go` 初始化时构建 `DriverCatalogService`、`RuntimeService`、`AgentProcessManager`、`JdbcGatewayService`，再传入 `DatabaseService`。

**步骤 4：运行测试确认通过**

```bash
go test ./internal/service -run 'TestDatabaseServiceDelegates|TestDatabaseService_CloseDatabase' -v
```

预期：PASS。

**步骤 5：全量后端测试**

```bash
go test ./...
```

预期：PASS；如有和 Java agent 尚未打包相关的测试，使用 fake gateway 隔离。

**步骤 6：提交**

```bash
git add app.go internal/service/database_service.go internal/service/database_service_test.go internal/service/database_drivers.go docs/changes/features/2026-07-09-database-service-jdbc-gateway.md
git commit -m "feat: route database service through jdbc gateway"
```

## 任务 9：新增驱动管理 Wails API

**文件：**
- 修改：`app.go`
- 创建：`internal/service/jdbc_api_models.go`
- 修改：`frontend/wailsjs/go/main/App.js`（由 Wails 生成）
- 创建：`docs/changes/features/2026-07-09-driver-manager-api.md`

**步骤 1：写失败测试**

在 Go 侧新增 service 测试，覆盖 API 方法背后的服务方法：

```go
func TestDriverManagerListsDriversWithInstallStatus(t *testing.T) {
	// 使用临时 manifest 和已安装目录
	// 断言返回 oracle installed=false，h2 installed=true
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestDriverManagerListsDriversWithInstallStatus -v
```

预期：失败。

**步骤 3：实现 App 导出方法**

新增 Wails 方法：

```go
func (a *App) ListJDBCDrivers() ([]service.DriverView, error)
func (a *App) InstallJDBCDriver(driverID, version string) error
func (a *App) ImportJDBCDriverPackage(path string) error
func (a *App) ValidateJDBCDriver(driverID, version string) error
func (a *App) RemoveJDBCDriver(driverID, version string) error
func (a *App) GetJDBCRuntimeStatus() (service.RuntimeStatus, error)
func (a *App) SetJDBCRuntimeMode(mode, path string) error
func (a *App) RestartJDBCAgent() error
```

首版如果文件选择器还未接入，`ImportJDBCDriverPackage(path)` 先接收路径，前端用 Wails runtime 文件对话框补齐。

**步骤 4：运行 Wails 生成**

```bash
wails generate module
```

如果当前 Wails 版本不支持该命令，运行：

```bash
wails dev
```

并停止 dev server 后确认 `frontend/wailsjs/go/main/App.js` 更新。

**步骤 5：运行测试**

```bash
go test ./internal/service -run TestDriverManager -v
go test ./...
```

预期：PASS。

**步骤 6：提交**

```bash
git add app.go internal/service/jdbc_api_models.go frontend/wailsjs docs/changes/features/2026-07-09-driver-manager-api.md
git commit -m "feat: expose jdbc driver management api"
```

## 任务 10：实现驱动管理前端页面

**文件：**
- 创建：`frontend/src/components/JDBCDriverManager.svelte`
- 修改：`frontend/src/components/GlobalSettingsDialog.svelte`
- 修改：`frontend/src/App.svelte`
- 修改：`frontend/src/styles/app.css` 或现有样式文件
- 创建：`docs/changes/features/2026-07-09-jdbc-driver-manager-ui.md`

**步骤 1：写手工验收清单**

在变更文档的验证段先写：

- 打开全局设置，能进入“数据库驱动”页。
- 左栏能搜索、过滤驱动。
- 右栏能展示驱动详情、版本、jar、URL template、高级配置。
- 未安装驱动显示“安装”和“导入离线包”。
- 已安装驱动显示“校验”“重新安装”“卸载”。
- 顶部显示 JRE 状态和 agent 状态。

**步骤 2：实现最小 UI**

`JDBCDriverManager.svelte`：

- `onMount` 调 `ListJDBCDrivers()` 和 `GetJDBCRuntimeStatus()`。
- 左栏列表状态由返回数据驱动。
- 选中项显示详情。
- 所有按钮先接真实 Wails API。
- 任务状态用本地 `isBusy` 和 `activeTaskMessage`；后续再做实时事件。

**步骤 3：集成入口**

在 `GlobalSettingsDialog.svelte` 增加“数据库驱动”设置项，渲染 `JDBCDriverManager`。

如现有设置弹窗不适合承载宽页面，可在 `App.svelte` 增加独立弹窗，尺寸使用 `xl`。

**步骤 4：运行前端构建**

```bash
cd frontend && npm run build
```

预期：PASS。

**步骤 5：手工运行**

```bash
wails dev
```

验证 UI 清单。记录截图或文字结果到变更文档。

**步骤 6：提交**

```bash
git add frontend/src/components/JDBCDriverManager.svelte frontend/src/components/GlobalSettingsDialog.svelte frontend/src/App.svelte frontend/src/styles/app.css docs/changes/features/2026-07-09-jdbc-driver-manager-ui.md
git commit -m "feat: add jdbc driver manager ui"
```

## 任务 11：扩展数据库连接表单支持首批 JDBC 类型

**文件：**
- 修改：`frontend/src/components/AddAssetDialog.svelte`
- 修改：`internal/config/database.go`
- 修改：`internal/service/jdbc_catalog.go`
- 创建：`docs/changes/features/2026-07-09-jdbc-connection-types.md`

**步骤 1：写失败测试**

`internal/service/jdbc_catalog_test.go`：

```go
func TestDriverCatalogReturnsDefaultPortsForInitialJDBCTypes(t *testing.T) {
	cases := map[string]int{
		"mysql": 3306,
		"postgresql": 5432,
		"sqlite": 0,
		"oracle": 1521,
		"sqlserver": 1433,
		"dm": 5236,
		"kingbase": 54321,
		"opengauss": 5432,
	}
	for dbType, want := range cases {
		if got := config.GetDefaultPort(dbType); got != want {
			t.Fatalf("%s default port: got %d want %d", dbType, got, want)
		}
	}
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestDriverCatalogReturnsDefaultPortsForInitialJDBCTypes -v
```

预期：失败，默认端口未覆盖。

**步骤 3：实现默认类型**

- `GetDefaultPort` 覆盖首批类型。
- `AddAssetDialog.svelte` 数据库类型下拉加入首批类型。
- 当所选类型未安装驱动时，连接按钮提示先安装驱动。
- SQLite 连接表单显示文件路径，隐藏 host/port；如果首版不做 SQLite 文件选择，先支持 JDBC URL 高级输入。

**步骤 4：运行测试和构建**

```bash
go test ./internal/service -run TestDriverCatalogReturnsDefaultPortsForInitialJDBCTypes -v
cd frontend && npm run build
```

预期：PASS。

**步骤 5：提交**

```bash
git add frontend/src/components/AddAssetDialog.svelte internal/config/database.go internal/service/jdbc_catalog.go internal/service/jdbc_catalog_test.go docs/changes/features/2026-07-09-jdbc-connection-types.md
git commit -m "feat: add initial jdbc database types"
```

## 任务 12：端到端 H2/SQLite 集成测试

**文件：**
- 创建：`internal/service/jdbc_integration_test.go`
- 创建：`testdata/jdbc/h2-driver-package.zip` 或测试内动态生成 zip
- 创建：`scripts/test-jdbc-agent.sh`
- 创建：`docs/changes/features/2026-07-09-jdbc-integration-test.md`

**步骤 1：写失败测试**

`jdbc_integration_test.go`：

- 构建或定位 `jdbc-agent.jar`。
- 导入 H2 driver package。
- 启动 agent。
- 建立 H2 memory session。
- 执行 `create table users...`。
- 执行 insert。
- 执行 select。
- 列出 tables 和 columns。
- 关闭 session。

如集成测试需要 Java/Gradle，使用 build tag：

```go
//go:build integration
```

**步骤 2：运行测试确认失败**

```bash
go test -tags=integration ./internal/service -run TestJDBCAgentH2EndToEnd -v
```

预期：初始失败，直到 agent jar 构建和路径配置完成。

**步骤 3：实现脚本**

`scripts/test-jdbc-agent.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
(cd jdbc-agent && ./gradlew clean test shadowJar)
go test -tags=integration ./internal/service -run TestJDBCAgentH2EndToEnd -v
```

**步骤 4：修正实现直到通过**

重点处理：

- agent jar 路径。
- 临时 `~/.sshtools` 根目录注入。
- 端口和 token。
- H2 URL template。
- 进程退出清理。

**步骤 5：提交**

```bash
git add internal/service/jdbc_integration_test.go scripts/test-jdbc-agent.sh docs/changes/features/2026-07-09-jdbc-integration-test.md
git commit -m "test: add jdbc agent integration coverage"
```

## 任务 13：错误分类、日志和前端行动按钮

**文件：**
- 创建：`internal/service/jdbc_errors.go`
- 修改：`internal/service/jdbc_gateway.go`
- 修改：`frontend/src/components/DatabasePanel.svelte`
- 修改：`frontend/src/components/JDBCDriverManager.svelte`
- 创建：`docs/changes/features/2026-07-09-jdbc-error-handling.md`

**步骤 1：写失败测试**

```go
func TestJDBCErrorMapsRuntimeMissingToActionableCode(t *testing.T) {
	err := MapJDBCAgentError("runtime not found")
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) {
		t.Fatalf("expected JDBCError")
	}
	if jdbcErr.Code != "RUNTIME_MISSING" {
		t.Fatalf("unexpected code: %s", jdbcErr.Code)
	}
}
```

**步骤 2：运行测试确认失败**

```bash
go test ./internal/service -run TestJDBCErrorMapsRuntimeMissingToActionableCode -v
```

预期：失败。

**步骤 3：实现错误分类**

错误代码：

- `RUNTIME_MISSING`
- `DRIVER_MISSING`
- `DRIVER_INVALID`
- `AGENT_UNAVAILABLE`
- `DB_CONNECT_FAILED`

日志路径：

- `~/.sshtools/logs/jdbc-agent.log`
- `~/.sshtools/logs/driver-install.log`
- `~/.sshtools/logs/runtime-install.log`

**步骤 4：前端行动按钮**

- `RUNTIME_MISSING`：安装 JRE / 导入 JRE / 选择系统 Java。
- `DRIVER_MISSING`：安装推荐驱动 / 导入离线包。
- `DRIVER_INVALID`：重新安装 / 查看文件 / 删除。
- `AGENT_UNAVAILABLE`：重启 agent / 查看日志。
- `DB_CONNECT_FAILED`：编辑连接 / 查看原始错误。

**步骤 5：运行验证**

```bash
go test ./internal/service -run TestJDBCError -v
cd frontend && npm run build
```

预期：PASS。

**步骤 6：提交**

```bash
git add internal/service/jdbc_errors.go internal/service/jdbc_gateway.go frontend/src/components/DatabasePanel.svelte frontend/src/components/JDBCDriverManager.svelte docs/changes/features/2026-07-09-jdbc-error-handling.md
git commit -m "feat: add actionable jdbc errors"
```

## 任务 14：最终验证和发布前整理

**文件：**
- 修改：`docs/development/2026-07-09-jdbc-driver-management-implementation.md`
- 修改：`docs/changes/features/2026-07-09-jdbc-driver-management-rollout.md`

**步骤 1：运行 Go 测试**

```bash
go test ./...
```

预期：PASS。

**步骤 2：运行 Java agent 测试**

```bash
cd jdbc-agent && ./gradlew test
```

预期：PASS。

**步骤 3：运行集成测试**

```bash
./scripts/test-jdbc-agent.sh
```

预期：PASS。

**步骤 4：运行前端构建**

```bash
cd frontend && npm run build
```

预期：PASS。

**步骤 5：运行 Wails 构建**

```bash
wails build
```

预期：PASS。

**步骤 6：手工验收**

按清单验证：

- 无 Java 环境时能提示安装托管 JRE。
- 离线导入 H2 driver package 后能查询。
- 驱动管理页能安装、校验、卸载、重新导入。
- 数据库连接表单能选择首批 JDBC 类型。
- agent 崩溃后前端提示重连。

**步骤 7：记录结果并提交**

```bash
git add docs/development/2026-07-09-jdbc-driver-management-implementation.md docs/changes/features/2026-07-09-jdbc-driver-management-rollout.md
git commit -m "docs: record jdbc driver management verification"
```

## 执行顺序建议

1. 任务 1 到 3 建立 Go 侧元数据、驱动包、运行时基础。
2. 任务 4 到 6 建立 Java agent 的可测试核心。
3. 任务 7 到 8 打通 Go 与 Java agent，并替换现有数据库路径。
4. 任务 9 到 11 暴露管理 API、实现 UI、扩展连接类型。
5. 任务 12 到 14 做集成测试、错误处理和最终验证。

## 主要风险和检查点

- **gRPC/protoc 工具链风险**：任务 4 或 7 前必须确认本机 protoc、Go plugin、Gradle protobuf plugin 可用。
- **Java agent 打包风险**：任务 4 结束必须能生成可运行 jar，否则不要进入 Go gateway。
- **厂商驱动 license 风险**：首版 manifest 只放可下载 metadata，不把受限 jar 提交进仓库。
- **前端范围风险**：任务 10 只做驱动管理，不顺手做数据编辑、导入导出。
- **现有数据库功能回归风险**：任务 8 后必须用 fake gateway 保证原 Wails API 仍可被现有前端调用。
