package database

import (
	"context"
	"log"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/config"
)

// Close 关闭数据库连接
func Close(ctx context.Context) error {
	// 如果未配置多租户，只关闭默认数据库
	if !config.TenantsConfig.Enabled {
		return closeDatabase(ctx, "*")
	}

	// 如果多租户为true，则关闭每个租户的数据库
	for k := range config.TenantsConfig.List {
		if err := closeDatabase(ctx, config.TenantsConfig.List[k].ID); err != nil {
			log.Printf("Error closing database for tenant %s: %v", config.TenantsConfig.List[k].ID, err)
		}
	}

	return nil
}

func closeDatabase(ctx context.Context, tenantID string) error {
	db := sdk.Runtime.GetTenantDB(tenantID)
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
