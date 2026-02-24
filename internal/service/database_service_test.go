package service

import (
	"database/sql"
	"testing"

	"AHaSSHTools/internal/config"
	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDatabaseService(t *testing.T, dbType, database string) (*DatabaseService, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	ds := NewDatabaseService(nil)
	ds.openFunc = func(driverName, dsn string) (*sql.DB, error) {
		return db, nil
	}

	ds.sessionStore["db-test"] = &DatabaseSession{
		ID:        "db-test",
		Config:    config.DatabaseConfig{DBType: dbType, Database: database},
		DB:        db,
		Connected: true,
	}

	cleanup := func() {
		_ = db.Close()
	}

	return ds, mock, cleanup
}

func TestDatabaseService_ExecuteQuery_ExecsNonSelect(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	mock.ExpectExec("UPDATE users SET active=1").WillReturnResult(sqlmock.NewResult(0, 3))

	result, err := ds.ExecuteQuery("db-test", "UPDATE users SET active=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Affected != 3 {
		t.Fatalf("expected affected=3, got %d", result.Affected)
	}
}

func TestDatabaseService_ExecuteQuery_SelectReturnsRows(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "alice").AddRow(2, "bob")
	mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)

	result, err := ds.ExecuteQuery("db-test", "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestDatabaseService_CloseDatabase_RemovesSession(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	mock.ExpectClose()

	if err := ds.CloseDatabase("db-test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := ds.GetSession("db-test"); err == nil {
		t.Fatalf("expected session to be removed")
	}
}

func TestDatabaseService_ListDatabases_MySQL(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"Database"}).AddRow("db1").AddRow("db2")
	mock.ExpectQuery("SHOW DATABASES").WillReturnRows(rows)

	result, err := ds.ListDatabases("db-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(result))
	}
}

func TestDatabaseService_ListDatabases_Postgres(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "postgresql", "postgres")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"datname"}).AddRow("postgres").AddRow("appdb")
	mock.ExpectQuery("SELECT datname FROM pg_database").WillReturnRows(rows)

	result, err := ds.ListDatabases("db-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(result))
	}
}

func TestDatabaseService_ListTablesInDatabase_MySQL(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"Tables_in_appdb"}).AddRow("users").AddRow("orders")
	mock.ExpectQuery("SHOW TABLES FROM `appdb`").WillReturnRows(rows)

	result, err := ds.ListTablesInDatabase("db-test", "appdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(result))
	}
}

func TestDatabaseService_ListTablesInDatabase_PostgresCurrent(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "postgresql", "appdb")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"tablename"}).AddRow("users")
	mock.ExpectQuery("SELECT tablename FROM pg_tables").WillReturnRows(rows)

	result, err := ds.ListTablesInDatabase("db-test", "appdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 table, got %d", len(result))
	}
}
