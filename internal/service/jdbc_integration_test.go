//go:build integration

package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"AHaSSHTools/internal/config"
)

func TestJDBCAgentH2EndToEnd(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	agentJar := os.Getenv("JDBC_AGENT_JAR")
	if agentJar == "" {
		agentJar = filepath.Join("..", "..", "jdbc-agent", "build", "libs", "sshtools-jdbc-agent-all.jar")
	}
	h2Jar := os.Getenv("H2_JAR")
	if h2Jar == "" {
		t.Fatal("H2_JAR is required")
	}

	zipPath := filepath.Join(root, "h2-driver-package.zip")
	createIntegrationDriverPackage(t, zipPath, h2Jar)
	installer := NewDriverInstallService(paths)
	installResult, err := installer.ImportOfflinePackage(zipPath)
	if err != nil {
		t.Fatalf("import H2 package failed: %v", err)
	}

	manager := NewAgentProcessManager(nil, AgentProcessConfig{
		JavaPath: "/Library/Java/JavaVirtualMachines/jdk-21.jdk/Contents/Home/bin/java",
		AgentJar: agentJar,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	handle, err := manager.Start(ctx)
	if err != nil {
		t.Fatalf("start agent failed: %v", err)
	}
	defer manager.Stop()

	client, closeClient, err := NewGRPCJdbcAgentClient(ctx, "127.0.0.1", handle.Port)
	if err != nil {
		t.Fatalf("connect agent grpc failed: %v", err)
	}
	defer closeClient()

	gateway := NewJdbcGatewayService(client, handle.Token)
	gateway.SetProfileResolver(func(ctx context.Context, cfg config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return config.JDBCDriverProfile{
			ID:          installResult.ProfileID,
			Version:     installResult.Version,
			DriverClass: "org.h2.Driver",
			URLTemplate: "jdbc:h2:mem:{database};DB_CLOSE_DELAY=-1",
			InstallPath: installResult.InstallPath,
			Jars:        []config.JDBCJar{{Name: "h2.jar"}},
		}, nil
	})

	cfg := config.DatabaseConfig{DBType: "h2", Database: "integration"}
	if err := gateway.ConnectDatabase(ctx, "h2-e2e", cfg); err != nil {
		t.Fatalf("connect database failed: %v", err)
	}
	if _, err := gateway.ExecuteQuery(ctx, "h2-e2e", "create table users (id int primary key, name varchar(32))"); err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	if _, err := gateway.ExecuteQuery(ctx, "h2-e2e", "insert into users values (1, 'alice')"); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	result, err := gateway.ExecuteQuery(ctx, "h2-e2e", "select id, name from users")
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if len(result.Rows) != 1 || result.Rows[0][1] != "alice" {
		t.Fatalf("unexpected query result: %+v", result)
	}
	tables, err := gateway.ListTables(ctx, "h2-e2e", "")
	if err != nil {
		t.Fatalf("list tables failed: %v", err)
	}
	if len(tables) == 0 {
		t.Fatalf("expected tables")
	}
	schema, err := gateway.GetTableSchema(ctx, "h2-e2e", "USERS")
	if err != nil {
		t.Fatalf("list columns failed: %v", err)
	}
	if len(schema.Columns) < 2 {
		t.Fatalf("expected columns: %+v", schema)
	}
	if err := gateway.CloseDatabase(ctx, "h2-e2e"); err != nil {
		t.Fatalf("close session failed: %v", err)
	}
}

func TestJDBCAgentRecoversSessionAfterCrash(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	agentJar := os.Getenv("JDBC_AGENT_JAR")
	if agentJar == "" {
		agentJar = filepath.Join("..", "..", "jdbc-agent", "build", "libs", "sshtools-jdbc-agent-all.jar")
	}
	h2Jar := os.Getenv("H2_JAR")
	if h2Jar == "" {
		t.Fatal("H2_JAR is required")
	}

	zipPath := filepath.Join(root, "h2-driver-package.zip")
	createIntegrationDriverPackage(t, zipPath, h2Jar)
	installResult, err := NewDriverInstallService(paths).ImportOfflinePackage(zipPath)
	if err != nil {
		t.Fatalf("import H2 package failed: %v", err)
	}

	javaPath := "/Library/Java/JavaVirtualMachines/jdk-21.jdk/Contents/Home/bin/java"
	runtimeService := NewRuntimeService(paths, javaPath)
	runtimeService.UseSystemJava(true)
	manager := NewAgentProcessManager(nil, AgentProcessConfig{JavaPath: javaPath, AgentJar: agentJar})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	handle, err := manager.Start(ctx)
	if err != nil {
		t.Fatalf("start agent failed: %v", err)
	}
	defer manager.Stop()

	supervisor := NewJDBCAgentSupervisor(runtimeService, manager, nil, agentJar)
	gateway := NewManagedJDBCGateway(supervisor)
	gateway.SetProfileResolver(func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return config.JDBCDriverProfile{
			ID:          installResult.ProfileID,
			Version:     installResult.Version,
			DriverClass: "org.h2.Driver",
			URLTemplate: "jdbc:h2:file:{database};WRITE_DELAY=0",
			InstallPath: installResult.InstallPath,
			Jars:        []config.JDBCJar{{Name: "h2.jar"}},
		}, nil
	})

	databasePath := filepath.Join(root, "recovery-db")
	cfg := config.DatabaseConfig{DBType: "h2", Database: databasePath}
	if err := gateway.ConnectDatabase(ctx, "h2-recovery", cfg); err != nil {
		t.Fatalf("connect database failed: %v", err)
	}
	if _, err := gateway.ExecuteQuery(ctx, "h2-recovery", "create table recovery_test (id int primary key, payload varchar(32))"); err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	if _, err := gateway.ExecuteQuery(ctx, "h2-recovery", "insert into recovery_test values (1, 'after-restart')"); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if err := handle.Process.Stop(); err != nil {
		t.Fatalf("stop agent process failed: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	result, err := gateway.ExecuteQuery(ctx, "h2-recovery", "select payload from recovery_test where id = 1")
	if err != nil {
		t.Fatalf("query after agent crash failed: %v", err)
	}
	if len(result.Rows) != 1 || len(result.Rows[0]) != 1 || result.Rows[0][0] != "after-restart" {
		t.Fatalf("unexpected recovery result: %+v", result)
	}
}

func createIntegrationDriverPackage(t *testing.T, zipPath, h2Jar string) {
	t.Helper()
	jarBytes, err := os.ReadFile(h2Jar)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(jarBytes)

	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	files := map[string][]byte{
		"package.json":     []byte(`{"id":"h2","name":"H2","version":"2.2.224","driverClass":"org.h2.Driver","urlTemplate":"jdbc:h2:mem:{database}","defaultPort":0,"jre":">=17","jars":["h2.jar"]}`),
		"checksums.sha256": []byte(hex.EncodeToString(sum[:]) + "  jars/h2.jar\n"),
		"jars/h2.jar":      jarBytes,
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(body); err != nil {
			t.Fatal(err)
		}
	}
}
