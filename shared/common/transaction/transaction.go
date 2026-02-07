package transaction

import "context"

// TransactionManager 事务管理器接口
type TransactionManager interface {
	// BeginTx 开启事务
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction 事务接口
type Transaction interface {
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
}

// TransactionFunc 事务函数类型
type TransactionFunc func(Transaction) error

// RunInTransaction 在事务中执行函数
func RunInTransaction(ctx context.Context, tm TransactionManager, fn TransactionFunc) error {
	tx, err := tm.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if tx != nil {
				// 发生panic时回滚事务
				_ = tx.Rollback()
			}
			panic(p) // 重新抛出panic
		} else if err != nil {
			if tx != nil {
				// 发生错误时回滚事务
				_ = tx.Rollback()
			}
		}
	}()

	// 执行事务函数
	if err = fn(tx); err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}
