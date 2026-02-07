package persistence

import (
	"context"
	"jxt-evidence-system/process-management/shared/common/transaction"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"gorm.io/gorm"
)

// GormTransaction GORM事务实现
type GormTransaction struct {
	tx *gorm.DB
}

func (t *GormTransaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *GormTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

// GormTransactionManager GORM事务管理器实现
type GormTransactionManager struct {
	GormRepository
}

func (m *GormTransactionManager) BeginTx(ctx context.Context) (transaction.Transaction, error) {
	db, err := m.GetOrm(ctx)
	if err != nil {
		return nil, err
	}

	// 开启事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		logger.Error("开启事务失败", "error", tx.Error)
		return nil, tx.Error
	}
	return &GormTransaction{tx: tx}, nil
}

// GetTx 从事务中获取GORM事务对象
func GetTx(tx transaction.Transaction) *gorm.DB {
	if gormTx, ok := tx.(*GormTransaction); ok {
		return gormTx.tx
	}
	return nil
}
