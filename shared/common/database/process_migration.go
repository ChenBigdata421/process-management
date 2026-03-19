package database

import (
	"fmt"
	"log"

	command_migration "jxt-evidence-system/process-management/cmd/migrate/migration"
	_ "jxt-evidence-system/process-management/cmd/migrate/migration/version"
	_ "jxt-evidence-system/process-management/cmd/migrate/migration/version-local"
	"jxt-evidence-system/process-management/shared/common/models"
	tenantdb "jxt-evidence-system/process-management/shared/common/tenantdb"

	"gorm.io/gorm"
)

// getCommandMigrationFunc 返回命令端迁移函数
// 注意：仅支持 MySQL，如果是 PostgreSQL 则跳过迁移（query 服务只需要读取）
func getProcessMigrationFunc() tenantdb.MigrationFunc {
	return func(tenantID int, db *gorm.DB) error {

		// 先创建 sys_migration 表
		if err := db.Debug().AutoMigrate(&models.Migration{}); err != nil {
			return fmt.Errorf("租户 %d 创建 sys_migration 表失败: %w", tenantID, err)
		}

		// 使用多租户隔离接口执行迁移
		command_migration.Migrate.SetTenantDb(tenantID, db.Debug())
		if err := command_migration.Migrate.MigrateTenant(tenantID); err != nil {
			return fmt.Errorf("租户 %d 数据库迁移失败: %w", tenantID, err)
		}

		log.Printf("[database] 租户 %d 数据库迁移完成", tenantID)
		return nil
	}
}
