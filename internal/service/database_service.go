package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"AHaSSHTools/internal/config"
	"github.com/go-sql-driver/mysql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	gateway       DatabaseGateway
	mu            sync.RWMutex
}

type DatabaseGateway interface {
	ConnectDatabase(ctx context.Context, sessionID string, cfg config.DatabaseConfig) error
	ExecuteQuery(ctx context.Context, sessionID string, query string) (*QueryResult, error)
	ListTables(ctx context.Context, sessionID, database string) ([]string, error)
	ListDatabases(ctx context.Context, sessionID string) ([]string, error)
	GetTableSchema(ctx context.Context, sessionID, table string) (*config.TableSchema, error)
	GetTableSchemaInSchema(ctx context.Context, sessionID, schema, table string) (*config.TableSchema, error)
	CloseDatabase(ctx context.Context, sessionID string) error
}

type schemaDatabaseGateway interface {
	ListSchemas(ctx context.Context, sessionID, database string) ([]string, error)
	ListTablesInSchema(ctx context.Context, sessionID, database, schema string) ([]string, error)
	ListObjects(ctx context.Context, sessionID, database, schema string, types []string) ([]string, error)
	ListRoutines(ctx context.Context, sessionID, database, schema string, functions bool) ([]string, error)
}

type databaseSchemaGateway interface {
	GetTableSchemaInDatabaseAndSchema(ctx context.Context, sessionID, database, schema, table string) (*config.TableSchema, error)
}

func NewDatabaseService(configManager *config.ConfigManager) *DatabaseService {
	return NewDatabaseServiceWithGateway(configManager, nil)
}

func NewDatabaseServiceWithGateway(configManager *config.ConfigManager, gateway DatabaseGateway) *DatabaseService {
	return &DatabaseService{
		configManager: configManager,
		sessionStore:  make(map[string]*DatabaseSession),
		openFunc:      sql.Open,
		gateway:       gateway,
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
	return ds.ConnectDatabaseWithProfile(sessionID, host, port, user, password, dbType, database, "")
}

func (ds *DatabaseService) ConnectDatabaseWithProfile(sessionID, host string, port int, user, password, dbType, database, driverProfileID string) error {
	return ds.ConnectDatabaseWithProfileAndProperties(sessionID, host, port, user, password, dbType, database, driverProfileID, nil)
}

func (ds *DatabaseService) ConnectDatabaseWithProfileAndProperties(sessionID, host string, port int, user, password, dbType, database, driverProfileID string, properties map[string]string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	normalizedType := strings.ToLower(strings.TrimSpace(dbType))
	databaseName := strings.TrimSpace(database)
	if normalizedType == "postgresql" && databaseName == "" {
		databaseName = "postgres"
	}
	cfg := config.DatabaseConfig{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		DBType:          normalizedType,
		Database:        databaseName,
		Timeout:         10 * time.Second,
		DriverProfileID: strings.TrimSpace(driverProfileID),
		Properties:      properties,
	}

	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()
		if err := ds.gateway.ConnectDatabase(ctx, sessionID, cfg); err != nil {
			return err
		}
		ds.mu.Lock()
		ds.sessionStore[sessionID] = &DatabaseSession{
			ID:        sessionID,
			Config:    cfg,
			Connected: true,
		}
		ds.mu.Unlock()
		return nil
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

func jdbcTestConnectionProperties(dbType string) map[string]string {
	if dbType != "mysql" {
		return nil
	}
	return map[string]string{
		"connectTimeout": "8000",
		"socketTimeout":  "8000",
	}
}

func explainJDBCTestConnectionError(err error, cfg config.DatabaseConfig) error {
	if err == nil || cfg.DBType != "mysql" {
		return err
	}
	message := strings.ToLower(err.Error())
	handshakeTimeout := errors.Is(err, context.DeadlineExceeded) ||
		status.Code(err) == codes.DeadlineExceeded ||
		strings.Contains(message, "deadlineexceeded") ||
		(strings.Contains(message, "communications link failure") &&
			strings.Contains(message, "has not received any packets"))
	if !handshakeTimeout {
		return err
	}
	return &JDBCError{
		Code: JDBCErrorDBConnectFailed,
		Message: fmt.Sprintf(
			"连接 MySQL %s:%d 超时：未收到 MySQL 服务端握手。请确认目标是 MySQL 原生端口；如果经过 TCP 代理或端口转发，代理必须在客户端发送数据前转发服务端握手。",
			cfg.Host,
			cfg.Port,
		),
		Err: err,
	}
}

func (ds *DatabaseService) ExecuteQuery(sessionID, query string) (*QueryResult, error) {
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ds.gateway.ExecuteQuery(ctx, sessionID, query)
	}

	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	trimmed := sanitizeJDBCSQL(query)
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
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ds.gateway.ListTables(ctx, sessionID, "")
	}

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
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ds.gateway.ListDatabases(ctx, sessionID)
	}

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
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ds.gateway.ListTables(ctx, sessionID, database)
	}

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

// ListSchemas returns schemas when the active JDBC gateway exposes metadata browsing.
func (ds *DatabaseService) ListSchemas(sessionID, database string) ([]string, error) {
	gateway, ok := ds.gateway.(schemaDatabaseGateway)
	if !ok {
		return nil, fmt.Errorf("当前数据库连接不支持 Schema 浏览")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return gateway.ListSchemas(ctx, sessionID, database)
}

// ListTablesInSchema returns tables scoped to the requested catalog and schema.
func (ds *DatabaseService) ListTablesInSchema(sessionID, database, schema string) ([]string, error) {
	return ds.ListObjects(sessionID, database, schema, nil)
}

// ListObjects returns JDBC objects filtered by standard metadata types.
func (ds *DatabaseService) ListObjects(sessionID, database, schema string, types []string) ([]string, error) {
	gateway, ok := ds.gateway.(schemaDatabaseGateway)
	if !ok {
		return nil, fmt.Errorf("当前数据库连接不支持 Schema 浏览")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return gateway.ListObjects(ctx, sessionID, database, schema, types)
}

func (ds *DatabaseService) ListRoutines(sessionID, database, schema string, functions bool) ([]string, error) {
	gateway, ok := ds.gateway.(schemaDatabaseGateway)
	if !ok {
		return nil, fmt.Errorf("当前数据库连接不支持例程浏览")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return gateway.ListRoutines(ctx, sessionID, database, schema, functions)
}

func (ds *DatabaseService) GetTableSchema(sessionID, table string) (*config.TableSchema, error) {
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ds.gateway.GetTableSchema(ctx, sessionID, table)
	}

	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, err := tableSchemaQuery(session.Config.DBType)
	if err != nil {
		return nil, err
	}
	rows, err := session.DB.QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}
	defer rows.Close()

	columns, err := scanTableSchemaRows(rows, session.Config.DBType)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns found for table: %s", table)
	}
	return &config.TableSchema{TableName: table, Columns: columns}, nil
}

func tableSchemaQuery(databaseType string) (string, error) {
	switch databaseType {
	case "mysql":
		return "SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? ORDER BY ORDINAL_POSITION", nil
	case "postgresql":
		return "SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position", nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", databaseType)
	}
}

func scanTableSchemaRows(rows *sql.Rows, databaseType string) ([]config.ColumnSchema, error) {
	columns := make([]config.ColumnSchema, 0)
	switch databaseType {
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
	return columns, nil
}

// GetTableSchemaInSchema returns structured column metadata scoped to a schema when supported by the active gateway.
func (ds *DatabaseService) GetTableSchemaInSchema(sessionID, database, schema, table string) (*config.TableSchema, error) {
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if gateway, ok := ds.gateway.(databaseSchemaGateway); ok {
			return gateway.GetTableSchemaInDatabaseAndSchema(ctx, sessionID, database, schema, table)
		}
		return ds.gateway.GetTableSchemaInSchema(ctx, sessionID, schema, table)
	}
	return ds.GetTableSchema(sessionID, table)
}

func (ds *DatabaseService) CloseDatabase(sessionID string) error {
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := ds.gateway.CloseDatabase(ctx, sessionID)
		ds.mu.Lock()
		delete(ds.sessionStore, sessionID)
		ds.mu.Unlock()
		return err
	}

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
	return ds.TestConnectionWithProperties(host, port, user, password, dbType, database, nil)
}

func (ds *DatabaseService) TestConnectionWithProperties(host string, port int, user, password, dbType, database string, properties map[string]string) error {
	normalizedType := strings.ToLower(strings.TrimSpace(dbType))
	databaseName := strings.TrimSpace(database)
	if normalizedType == "postgresql" && databaseName == "" {
		databaseName = "postgres"
	}
	cfg := config.DatabaseConfig{
		Host:       host,
		Port:       port,
		User:       user,
		Password:   password,
		DBType:     normalizedType,
		Database:   databaseName,
		Timeout:    10 * time.Second,
		Properties: mergeJDBCConnectionProperties(jdbcTestConnectionProperties(normalizedType), properties),
	}
	if ds.gateway != nil {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()
		sessionID := fmt.Sprintf("jdbc-connection-test-%d", time.Now().UnixNano())
		if err := ds.gateway.ConnectDatabase(ctx, sessionID, cfg); err != nil {
			return explainJDBCTestConnectionError(err, cfg)
		}
		if err := ds.gateway.CloseDatabase(ctx, sessionID); err != nil {
			return fmt.Errorf("关闭 JDBC 测试连接失败: %w", err)
		}
		return nil
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

func mergeJDBCConnectionProperties(defaults, supplied map[string]string) map[string]string {
	if len(defaults) == 0 && len(supplied) == 0 {
		return nil
	}
	merged := make(map[string]string, len(defaults)+len(supplied))
	for key, value := range defaults {
		merged[key] = value
	}
	for key, value := range supplied {
		merged[key] = value
	}
	return merged
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

// CloseAllSessions closes every active database session.
func (ds *DatabaseService) CloseAllSessions() error {
	if ds == nil {
		return nil
	}
	ids := ds.ListSessions()
	var errs []error
	for _, id := range ids {
		if err := ds.CloseDatabase(id); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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

// TableDDL represents the DDL (Data Definition Language) for a table
type TableDDL struct {
	TableName string `json:"table_name"`
	DDL       string `json:"ddl"`
	DBType    string `json:"db_type"`
}

// GetTableDDL returns the CREATE TABLE statement for a given table
func (ds *DatabaseService) GetTableDDL(sessionID, database, table string) (*TableDDL, error) {
	return ds.GetTableDDLInSchema(sessionID, database, "", table)
}

// GetTableDDLInSchema returns the CREATE TABLE statement for a table in a specific schema.
func (ds *DatabaseService) GetTableDDLInSchema(sessionID, database, schemaName, table string) (*TableDDL, error) {
	session, err := ds.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if ds.gateway != nil {
		switch session.Config.DBType {
		case "mysql":
			escapedTable := escapeMySQLIdentifier(table)
			query := fmt.Sprintf("SHOW CREATE TABLE `%s`", escapedTable)
			if database != "" {
				query = fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", escapeMySQLIdentifier(database), escapedTable)
			}
			result, err := ds.ExecuteQuery(sessionID, query)
			if err != nil {
				return nil, fmt.Errorf("获取表结构失败: %w", err)
			}
			if len(result.Rows) == 0 || len(result.Rows[0]) < 2 {
				return nil, fmt.Errorf("表 %s 未返回有效 DDL", table)
			}
			return &TableDDL{
				TableName: fmt.Sprint(result.Rows[0][0]),
				DDL:       fmt.Sprint(result.Rows[0][1]),
				DBType:    "mysql",
			}, nil
		case "postgresql", "kingbase", "opengauss":
			schema, err := ds.GetTableSchemaInSchema(sessionID, database, schemaName, table)
			if err != nil {
				return nil, fmt.Errorf("获取表结构失败: %w", err)
			}
			return postgreSQLCompatibleTableDDL(table, session.Config.DBType, schema)
		default:
			schema, err := ds.GetTableSchemaInSchema(sessionID, database, schemaName, table)
			if err != nil {
				return nil, fmt.Errorf("获取表结构失败: %w", err)
			}
			return postgreSQLCompatibleTableDDL(table, session.Config.DBType, schema)
		}
	}

	if session.DB == nil {
		return nil, fmt.Errorf("database connection not available: %s", sessionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch session.Config.DBType {
	case "mysql":
		return ds.getMySQLTableDDL(ctx, session.DB, database, table)
	case "postgresql":
		return ds.getPostgreSQLTableDDL(ctx, session.DB, table)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", session.Config.DBType)
	}
}

func postgreSQLCompatibleTableDDL(table, dbType string, schema *config.TableSchema) (*TableDDL, error) {
	if schema == nil || len(schema.Columns) == 0 {
		return nil, fmt.Errorf("表 %s 未返回字段定义", table)
	}
	var ddl strings.Builder
	ddl.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quotePostgreSQLIdentifier(table)))
	for index, column := range schema.Columns {
		if strings.TrimSpace(column.Name) == "" || strings.TrimSpace(column.Type) == "" {
			return nil, fmt.Errorf("表 %s 包含无效字段定义", table)
		}
		columnType := postgreSQLCompatibleColumnType(column)
		ddl.WriteString(fmt.Sprintf("    %s %s", quotePostgreSQLIdentifier(column.Name), columnType))
		if !column.Nullable {
			ddl.WriteString(" NOT NULL")
		}
		if column.IsPrimaryKey {
			ddl.WriteString(" PRIMARY KEY")
		}
		if column.HasDefault {
			ddl.WriteString(" DEFAULT ")
			ddl.WriteString(column.DefaultValue)
		}
		if index < len(schema.Columns)-1 {
			ddl.WriteString(",")
		}
		ddl.WriteString("\n")
	}
	ddl.WriteString(");")
	return &TableDDL{TableName: table, DDL: ddl.String(), DBType: dbType}, nil
}

func postgreSQLCompatibleColumnType(column config.ColumnSchema) string {
	typeName := strings.TrimSpace(column.Type)
	lowerType := strings.ToLower(typeName)
	if column.ColumnSize > 0 && (strings.Contains(lowerType, "character") || strings.Contains(lowerType, "varchar")) {
		return fmt.Sprintf("%s(%d)", typeName, column.ColumnSize)
	}
	if column.ColumnSize > 0 && column.DecimalDigits > 0 && (strings.Contains(lowerType, "numeric") || strings.Contains(lowerType, "decimal")) {
		return fmt.Sprintf("%s(%d,%d)", typeName, column.ColumnSize, column.DecimalDigits)
	}
	return typeName
}

func quotePostgreSQLIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func (ds *DatabaseService) getMySQLTableDDL(ctx context.Context, db *sql.DB, database, table string) (*TableDDL, error) {
	escapedTable := escapeMySQLIdentifier(table)
	query := fmt.Sprintf("SHOW CREATE TABLE %s", escapedTable)
	if database != "" {
		escapedDB := escapeMySQLIdentifier(database)
		query = fmt.Sprintf("SHOW CREATE TABLE %s.%s", escapedDB, escapedTable)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table DDL: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no DDL found for table: %s", table)
	}

	// Get column count to handle different MySQL versions
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var tableName, createStatement string
	if len(columns) == 2 {
		// MySQL 5.7 and MariaDB
		if err := rows.Scan(&tableName, &createStatement); err != nil {
			return nil, fmt.Errorf("failed to scan DDL: %w", err)
		}
	} else {
		// MySQL 8.0+ has additional columns
		var extra1, extra2 sql.NullString
		if err := rows.Scan(&tableName, &createStatement, &extra1, &extra2); err != nil {
			return nil, fmt.Errorf("failed to scan DDL: %w", err)
		}
	}

	return &TableDDL{
		TableName: tableName,
		DDL:       createStatement,
		DBType:    "mysql",
	}, nil
}

func (ds *DatabaseService) getPostgreSQLTableDDL(ctx context.Context, db *sql.DB, table string) (*TableDDL, error) {
	// Query column information
	query := `
		SELECT 
			column_name,
			data_type,
			COALESCE(character_maximum_length, 0) as max_length,
			is_nullable,
			column_default
		FROM information_schema.columns 
		WHERE table_name = $1 
			AND table_schema = 'public'
		ORDER BY ordinal_position
	`
	rows, err := db.QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table columns: %w", err)
	}
	defer rows.Close()

	type columnInfo struct {
		name       string
		dataType   string
		maxLength  int
		nullable   string
		defaultVal interface{}
	}

	columns := make([]columnInfo, 0)
	for rows.Next() {
		var ci columnInfo
		if err := rows.Scan(&ci.name, &ci.dataType, &ci.maxLength, &ci.nullable, &ci.defaultVal); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		columns = append(columns, ci)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("table not found: %s", table)
	}

	// Build CREATE TABLE statement
	var ddl strings.Builder
	ddl.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table))

	for i, col := range columns {
		dbType := col.dataType
		if col.maxLength > 0 && (col.dataType == "character varying" || col.dataType == "character") {
			dbType = fmt.Sprintf("%s(%d)", col.dataType, col.maxLength)
		}

		ddl.WriteString(fmt.Sprintf("    %s %s", col.name, dbType))

		if col.nullable == "NO" {
			ddl.WriteString(" NOT NULL")
		}

		if col.defaultVal != nil {
			ddl.WriteString(fmt.Sprintf(" DEFAULT %v", col.defaultVal))
		}

		if i < len(columns)-1 {
			ddl.WriteString(",\n")
		} else {
			ddl.WriteString("\n")
		}
	}

	ddl.WriteString(");")

	return &TableDDL{
		TableName: table,
		DDL:       ddl.String(),
		DBType:    "postgresql",
	}, nil
}
