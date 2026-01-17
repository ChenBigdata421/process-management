//go:build !sqlite3

package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var opens = map[string]func(string) gorm.Dialector{
	"postgres": postgres.Open,
	// 如需支持其他数据库，请添加相应的驱动：
	// "mysql":     mysql.Open,
	// "sqlserver": sqlserver.Open,
}
