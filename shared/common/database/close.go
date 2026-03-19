package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"gorm.io/gorm"
)

//并发关闭：使用 goroutine 并发关闭多个数据库连接，提高效率。
//错误处理：收集所有数据库关闭过程中的错误，而不是在第一个错误出现时就停止。
//超时控制：为整个关闭过程设置了一个总体超时（60秒），可以根据需要调整。
//日志记录：添加了更详细的日志记录，便于调试和监控。
//代码结构：将单个数据库的关闭逻辑抽取到 closeDatabase 函数中，提高了代码的可读性和可维护性。

// Close 优雅关闭所有数据库连接
func Close(ctx context.Context) error {
	return CloseTenantDBs(ctx)
}

// CloseTenantDBs 优雅关闭所有租户数据库连接
func CloseTenantDBs(ctx context.Context) error {
	var errors []error
	var errMu sync.Mutex
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	sdk.Runtime.GetTenantDBs(func(tenantID int, db *gorm.DB) bool {
		wg.Add(1)
		go func(tid int, database *gorm.DB) {
			defer wg.Done()
			if err := closeDatabase(ctx, fmt.Sprintf("tenant-%d", tid), database); err != nil {
				errMu.Lock()
				errors = append(errors, fmt.Errorf("error closing database for tenant %d: %w", tid, err))
				errMu.Unlock()
			}
		}(tenantID, db)
		return true
	})

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing tenant databases: %v", errors)
	}

	return nil
}

func closeDatabase(ctx context.Context, dbName string, db *gorm.DB) error {
	log.Printf("Closing database: %s", dbName)
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying SQL DB: %w", err)
	}

	// 停止接受新的查询
	sqlDB.SetMaxOpenConns(0)

	// 等待活跃连接释放
	for {
		if sqlDB.Stats().InUse == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for active connections to close")
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 关闭数据库连接
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}

	log.Printf("Database closed successfully: %s", dbName)
	return nil
}
