package service

import (
	// 保留 database/sql 兼容回退驱动注册；主应用路径会优先走 JdbcGatewayService。
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)
