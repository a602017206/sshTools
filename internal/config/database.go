package config

import "time"

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBType   string
	Database string
	Timeout  time.Duration
}

type TableSchema struct {
	TableName string         `json:"table_name"`
	Columns   []ColumnSchema `json:"columns"`
}

type ColumnSchema struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
}

func GetDefaultPort(dbType string) int {
	switch dbType {
	case "mysql":
		return 3306
	case "postgresql":
		return 5432
	default:
		return 3306
	}
}
