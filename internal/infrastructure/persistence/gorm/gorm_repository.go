package persistence

import (
	"context"
	"fmt"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"gorm.io/gorm"
)

type GormRepository struct{}

// GetDB 从 SDK Runtime 获取数据库连接
func (e *GormRepository) GetDB(ctx context.Context) *gorm.DB {
	db := sdk.Runtime.GetTenantDB("*")
	if db == nil {
		panic("database not initialized, call database.Setup first")
	}
	return db.WithContext(ctx)
}

// GetOrm 获取带上下文的数据库连接（兼容旧代码）
func (e *GormRepository) GetOrm(ctx context.Context) (*gorm.DB, error) {
	db := sdk.Runtime.GetTenantDB("*")
	if db == nil {
		return nil, fmt.Errorf("database not initialized, call database.Setup first")
	}
	return db.WithContext(ctx), nil
}
