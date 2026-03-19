package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"jxt-evidence-system/process-management/shared/common/global"
	"gorm.io/gorm"
)

type GormRepository struct{}

// GetDB 从 SDK Runtime 获取数据库连接
// 注意：此方法保留panic行为，用于向后兼容
func (e *GormRepository) GetDB(ctx context.Context) *gorm.DB {
	// 从上下文中获取tenantID
	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok {
		panic("tenant id not exist in context")
	}

	db := sdk.Runtime.GetTenantDB(tenantID)
	if db == nil {
		panic(fmt.Sprintf("database not initialized for tenant: %d", tenantID))
	}
	return db.WithContext(ctx)
}

// GetOrm 获取带上下文的数据库连接（兼容旧代码）
func (e *GormRepository) GetOrm(ctx context.Context) (*gorm.DB, error) {
	// 从上下文中获取tenantID
	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok {
		return nil, errors.New("tenant id not exist in context")
	}

	// 验证租户ID必须大于0
	if tenantID <= 0 {
		return nil, fmt.Errorf("invalid tenant id: %d (must be > 0)", tenantID)
	}

	db := sdk.Runtime.GetTenantDB(tenantID)
	if db == nil {
		return nil, fmt.Errorf("database not initialized for tenant: %d", tenantID)
	}
	return db.WithContext(ctx), nil
}
