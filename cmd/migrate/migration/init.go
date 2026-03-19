package migration

import (
	"path/filepath"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/migration"
)

// Migrate 全局迁移注册表实例（使用 jxt-core 的实现）
var Migrate = migration.GetRegistry()

// GetFilename 从文件路径提取版本号（前13位）
func GetFilename(s string) string {
	s = filepath.Base(s)
	return s[:13]
}
