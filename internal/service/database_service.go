package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"AHaSSHTools/internal/config"
	"github.com/go-sql-driver/mysql"
)

type DatabaseSession struct {
	ID        string
	Config    config.DatabaseConfig
	DB        *sql.DB
	Connected bool
}

type QueryResult struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	Affected int             `json:"affected"`
}

type DatabaseService struct {
	configManager *config.ConfigManager
	sessionStore  map[string]*DatabaseSession
	openFunc      func(driverName, dsn string) (*sql.DB, error)
	mu            sync.RWMutex
}

func NewDatabaseService(configManager *config.ConfigManager) *DatabaseService {
	return &DatabaseService{
		configManager: configManager,
		sessionStore:  make(map[string]*DatabaseSession),
		openFunc:      sql.Open,
	}
}

func (ds *DatabaseService) GetDSN(cfg config.DatabaseConfig) (string, error) {
	switch cfg.DBType {
	case "mysql":
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 10 * time.Second
		}
		mysqlCfg := mysql.Config{
			User:                 cfg.User,
			Passwd:               cfg.Password,
			Net:                  "tcp",
			Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			DBName:               cfg.Database,
			ParseTime:            true,
			Loc:                  time.Local,
			Timeout:              timeout,
			ReadTimeout:          30 * time.Second,
			WriteTimeout:         30 * time.Second,
			AllowNativePasswords: true,
		}
		return mysqlCfg.FormatDSN(), nil
	case "postgresql":
		connectTimeout := 10
		if cfg.Timeout > 0 {
			connectTimeout = int(cfg.Timeout.Seconds())
		}
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.Database,
			connectTimeout,
		), nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}
}

func (ds *DatabaseService) GetDriverName(dbType string) string {
	switch dbType {
	case "mysql":
		return "mysql"
	case "postgresql":
		return "postgres"
	default:
		return ""
	}
}

func (ds *DatabaseService) ConnectDatabase(sessionID, host string, port int, user, password, dbType, database string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	normalizedType := strings.ToLower(strings.TrimSpace(dbType))
	databaseName := strings.TrimSpace(database)
	if normalizedType == "postgresql" && databaseName == "" {
		databaseName = "postgres"
	}
	cfg := config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBType:   normalizedType,
		Database: databaseName,
		Timeout:  10 * time.Second,
	}

	dsn, err := ds.GetDSN(cfg)
	if err != nil {
		return err
	}
	driverName := ds.GetDriverName(normalizedType)
	if driverName == "" {
		return fmt.Errorf("unsupported database type: %s", normalizedType)
	}

	db, err := ds.openFunc(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("database ping failed: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ds.mu.Lock()
	if existing, ok := ds.sessionStore[sessionID]; ok && existing.DB != nil {
		_ = existing.DB.Close()
	}
	ds.sessionStore[sessionID] = &DatabaseSession{
		ID:        sessionID,
		Config:    cfg,
		DB:        db,
		Connected: true,
	}
	ds.mu.Unlock()

	return nil
}

func (ds *DatabaseService) ExecuteQuery(sessionID, query string) (*QueryResult, error) {
	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, fmt.Errorf("query is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !isQueryReturningRows(trimmed) {
		result, err := session.DB.ExecContext(ctx, trimmed)
		if err != nil {
			return nil, fmt.Errorf("query execution failed: %w", err)
		}
		affected, _ := result.RowsAffected()
		return &QueryResult{Affected: int(affected)}, nil
	}

	rows, err := session.DB.QueryContext(ctx, trimmed)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    make([][]interface{}, 0),
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range scanArgs {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		row := make([]interface{}, len(values))
		for i, val := range values {
			switch v := val.(type) {
			case []byte:
				row[i] = string(v)
			default:
				row[i] = v
			}
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return result, nil
}

func (ds *DatabaseService) ListTables(sessionID string) ([]string, error) {
	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return ds.listTablesWithDB(ctx, session.Config.DBType, session.DB, "")
}

func (ds *DatabaseService) ListDatabases(sessionID string) ([]string, error) {
	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var (
		rows *sql.Rows
		qErr error
	)

	switch session.Config.DBType {
	case "mysql":
		rows, qErr = session.DB.QueryContext(ctx, "SHOW DATABASES")
	case "postgresql":
		rows, qErr = session.DB.QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", session.Config.DBType)
	}

	if qErr != nil {
		return nil, fmt.Errorf("failed to list databases: %w", qErr)
	}
	defer rows.Close()

	databases := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		databases = append(databases, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return databases, nil
}

func (ds *DatabaseService) ListTablesInDatabase(sessionID, database string) ([]string, error) {
	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch session.Config.DBType {
	case "mysql":
		return ds.listTablesWithDB(ctx, session.Config.DBType, session.DB, database)
	case "postgresql":
		if database == "" || database == session.Config.Database {
			return ds.listTablesWithDB(ctx, session.Config.DBType, session.DB, "")
		}

		cfg := session.Config
		cfg.Database = database
		dsn, err := ds.GetDSN(cfg)
		if err != nil {
			return nil, err
		}
		driverName := ds.GetDriverName(cfg.DBType)
		if driverName == "" {
			return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
		}

		tempDB, err := ds.openFunc(driverName, dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		defer tempDB.Close()

		return ds.listTablesWithDB(ctx, cfg.DBType, tempDB, "")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", session.Config.DBType)
	}
}

func (ds *DatabaseService) GetTableSchema(sessionID, table string) (*config.TableSchema, error) {
	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var (
		rows *sql.Rows
		qErr error
	)

	switch session.Config.DBType {
	case "mysql":
		query := "SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		rows, qErr = session.DB.QueryContext(ctx, query, table)
	case "postgresql":
		query := "SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position"
		rows, qErr = session.DB.QueryContext(ctx, query, table)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", session.Config.DBType)
	}

	if qErr != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", qErr)
	}
	defer rows.Close()

	columns := make([]config.ColumnSchema, 0)
	switch session.Config.DBType {
	case "mysql":
		for rows.Next() {
			var columnName, dataType, nullable, columnKey string
			if err := rows.Scan(&columnName, &dataType, &nullable, &columnKey); err != nil {
				return nil, fmt.Errorf("scan error: %w", err)
			}
			columns = append(columns, config.ColumnSchema{
				Name:         columnName,
				Type:         dataType,
				Nullable:     nullable == "YES",
				IsPrimaryKey: columnKey == "PRI",
			})
		}
	case "postgresql":
		for rows.Next() {
			var columnName, dataType, nullable string
			var columnDefault interface{}
			if err := rows.Scan(&columnName, &dataType, &nullable, &columnDefault); err != nil {
				return nil, fmt.Errorf("scan error: %w", err)
			}
			columns = append(columns, config.ColumnSchema{
				Name:         columnName,
				Type:         dataType,
				Nullable:     nullable == "YES",
				IsPrimaryKey: false,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns found for table: %s", table)
	}

	return &config.TableSchema{
		TableName: table,
		Columns:   columns,
	}, nil
}

func (ds *DatabaseService) CloseDatabase(sessionID string) error {
	ds.mu.Lock()
	session, exists := ds.sessionStore[sessionID]
	if exists {
		delete(ds.sessionStore, sessionID)
	}
	ds.mu.Unlock()

	if !exists || session == nil || session.DB == nil {
		return nil
	}

	if err := session.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

func (ds *DatabaseService) TestConnection(host string, port int, user, password, dbType, database string) error {
	normalizedType := strings.ToLower(strings.TrimSpace(dbType))
	databaseName := strings.TrimSpace(database)
	if normalizedType == "postgresql" && databaseName == "" {
		databaseName = "postgres"
	}
	cfg := config.DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBType:   normalizedType,
		Database: databaseName,
		Timeout:  10 * time.Second,
	}

	dsn, err := ds.GetDSN(cfg)
	if err != nil {
		return err
	}
	driverName := ds.GetDriverName(normalizedType)
	if driverName == "" {
		return fmt.Errorf("unsupported database type: %s", normalizedType)
	}

	db, err := ds.openFunc(driverName, dsn)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	return db.PingContext(ctx)
}

func (ds *DatabaseService) GetSession(sessionID string) (*DatabaseSession, error) {
	ds.mu.RLock()
	session, exists := ds.sessionStore[sessionID]
	ds.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if session == nil || !session.Connected {
		return nil, fmt.Errorf("database session not connected: %s", sessionID)
	}
	return session, nil
}

func (ds *DatabaseService) ListSessions() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	sessions := make([]string, 0, len(ds.sessionStore))
	for id := range ds.sessionStore {
		sessions = append(sessions, id)
	}
	return sessions
}

func isQueryReturningRows(query string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(query))
	switch {
	case strings.HasPrefix(trimmed, "select"):
		return true
	case strings.HasPrefix(trimmed, "show"):
		return true
	case strings.HasPrefix(trimmed, "describe"):
		return true
	case strings.HasPrefix(trimmed, "with"):
		return true
	case strings.HasPrefix(trimmed, "explain"):
		return true
	default:
		return false
	}
}

func (ds *DatabaseService) listTablesWithDB(ctx context.Context, dbType string, db *sql.DB, database string) ([]string, error) {
	var (
		rows *sql.Rows
		qErr error
	)

	switch dbType {
	case "mysql":
		if database == "" {
			rows, qErr = db.QueryContext(ctx, "SHOW TABLES")
		} else {
			escaped := escapeMySQLIdentifier(database)
			rows, qErr = db.QueryContext(ctx, fmt.Sprintf("SHOW TABLES FROM `%s`", escaped))
		}
	case "postgresql":
		rows, qErr = db.QueryContext(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if qErr != nil {
		return nil, fmt.Errorf("failed to list tables: %w", qErr)
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return tables, nil
}

func escapeMySQLIdentifier(input string) string {
	return strings.ReplaceAll(input, "`", "``")
}
