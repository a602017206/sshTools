package config

import "time"

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

type TableSchema struct {
	TableName string         `json:"table_name"`
	Columns   []ColumnSchema `json:"columns"`
}

type ColumnSchema struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Nullable      bool   `json:"nullable"`
	IsPrimaryKey  bool   `json:"is_primary_key"`
	ColumnSize    int    `json:"column_size"`
	DecimalDigits int    `json:"decimal_digits"`
	DefaultValue  string `json:"default_value"`
	HasDefault    bool   `json:"has_default"`
	Description   string `json:"description"`
}

func GetDefaultPort(dbType string) int {
	switch dbType {
	case "mysql":
		return 3306
	case "postgresql":
		return 5432
	case "sqlite":
		return 0
	case "oracle":
		return 1521
	case "sqlserver":
		return 1433
	case "dm":
		return 5236
	case "kingbase":
		return 54321
	case "opengauss":
		return 5432
	default:
		return 3306
	}
}
