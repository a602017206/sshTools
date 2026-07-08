# Database Connection Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deliver a working MySQL/PostgreSQL database connection flow with query execution and table browsing wired end-to-end in the SSHTools UI.

**Architecture:** Keep the backend service as the source of truth for database sessions, add driver registration + context timeouts, and expose a small Wails API surface. On the frontend, ensure connection metadata is preserved, prompt for credentials if needed, and drive the DatabasePanel via direct Wails calls (no custom DOM events).

**Tech Stack:** Go (database/sql, Wails), Svelte, Wails JS bindings, MySQL/Postgres drivers, go-sqlmock for tests.

---

### Task 1: Add database drivers and shared test helpers

**Files:**
- Create: `internal/service/database_drivers.go`
- Modify: `go.mod`
- Create: `internal/service/database_service_test.go`

**Step 1: Write the failing test**

```go
func TestDatabaseService_OpenFailsWithUnknownDriver(t *testing.T) {
    ds := NewDatabaseService(nil)
    err := ds.ConnectDatabase("db-test", "localhost", 3306, "root", "pw", "unknown", "db")
    if err == nil {
        t.Fatalf("expected error for unknown driver")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestDatabaseService_OpenFailsWithUnknownDriver -v`
Expected: FAIL with missing driver or empty driver error

**Step 3: Write minimal implementation**

```go
// internal/service/database_drivers.go
package service

import (
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
)
```

Update `go.mod` with:
```
require (
    github.com/go-sql-driver/mysql v1.7.1
    github.com/lib/pq v1.10.9
    github.com/DATA-DOG/go-sqlmock v1.5.2
)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestDatabaseService_OpenFailsWithUnknownDriver -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/database_drivers.go internal/service/database_service_test.go go.mod go.sum
git commit -m "feat: register database drivers and test helpers"
```

---

### Task 2: Make database service timeout-safe and Exec/Query correct

**Files:**
- Modify: `internal/service/database_service.go`
- Test: `internal/service/database_service_test.go`

**Step 1: Write the failing test**

```go
func TestDatabaseService_ExecuteQuery_ExecsNonSelect(t *testing.T) {
    ds, mock, cleanup := newMockDatabaseService(t, "mysql")
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestDatabaseService_ExecuteQuery_ExecsNonSelect -v`
Expected: FAIL because ExecuteQuery always uses Query

**Step 3: Write minimal implementation**

```go
// Add a helper:
func isQueryReturningRows(query string) bool { /* detect SELECT/SHOW/DESCRIBE/WITH */ }

// ExecuteQuery: use context.WithTimeout
// For non-row queries: ExecContext + RowsAffected
```

Add a `sync.RWMutex` to guard `sessionStore`, and use `PingContext` for liveness checks.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestDatabaseService_ExecuteQuery_ExecsNonSelect -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/database_service.go internal/service/database_service_test.go
git commit -m "feat: add timeouts and exec handling for database queries"
```

---

### Task 3: Preserve DB metadata and password flow in frontend

**Files:**
- Modify: `frontend/src/components/AddAssetDialog.svelte`
- Modify: `frontend/src/App.svelte`
- Modify: `frontend/src/components/DatabasePanel.svelte`

**Step 1: Write the failing test**

For frontend, document expected behavior in a TODO test note (manual test list):

```markdown
- Create DB asset with db_type=postgresql and database name, save, reload — values persist
- Connect to DB asset: prompt for password if not saved; uses saved password if present
- Database panel loads table list and executes query successfully
```

**Step 2: Run test to verify it fails**

Manual run: `wails dev` and verify failures above in current UI.

**Step 3: Write minimal implementation**

```svelte
// AddAssetDialog: save metadata.db_type and load it when editing.
metadata: { database: formData.database || undefined, db_type: formData.dbType }

// App.svelte loadAssetsFromBackend: map type + metadata.
type: conn.type || 'ssh', metadata: conn.metadata || {}

// App.svelte handleDatabaseConnect: use HasPassword/GetPassword; prompt via InputDialog
```

Update `DatabasePanel.svelte` to remove custom DOM event listeners and rely on `ExecuteDatabaseQuery`, `ListDatabaseTables`, and `GetTableColumns` directly. Trigger `loadTables()` on mount; add a left sidebar table list UI and a small query history list (max 50).

**Step 4: Run test to verify it passes**

Manual run: `wails dev` and re-check the manual test list.

**Step 5: Commit**

```bash
git add frontend/src/components/AddAssetDialog.svelte frontend/src/App.svelte frontend/src/components/DatabasePanel.svelte
git commit -m "feat: wire database metadata and panel data flow"
```

---

### Task 4: Close lifecycle and cleanup

**Files:**
- Modify: `frontend/src/App.svelte`
- Modify: `internal/service/database_service.go`

**Step 1: Write the failing test**

```go
func TestDatabaseService_CloseDatabase_RemovesSession(t *testing.T) {
    ds, _, cleanup := newMockDatabaseService(t, "mysql")
    defer cleanup()

    if err := ds.CloseDatabase("db-test"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if _, err := ds.GetSession("db-test"); err == nil {
        t.Fatalf("expected session to be removed")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestDatabaseService_CloseDatabase_RemovesSession -v`
Expected: FAIL if CloseDatabase doesn't remove session or is unsafe

**Step 3: Write minimal implementation**

```go
// Ensure CloseDatabase acquires lock, closes DB, deletes session
```

In `App.svelte`, call `CloseDatabase(sessionId)` when closing the panel.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestDatabaseService_CloseDatabase_RemovesSession -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/database_service.go internal/service/database_service_test.go frontend/src/App.svelte
git commit -m "feat: close database sessions on panel close"
```

---

### Task 5: Verification

**Files:**
- None

**Step 1: Run Go tests**

Run: `go test ./...`
Expected: PASS (note any pre-existing failures)

**Step 2: Run frontend build**

Run: `cd frontend && npm run build`
Expected: PASS

**Step 3: Run Wails build**

Run: `wails build`
Expected: PASS

**Step 4: Commit**

```bash
git add -A
git commit -m "test: verify database feature"
```
