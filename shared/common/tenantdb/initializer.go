// shared/common/tenantdb/initializer.go
package tenantdb

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/tenant/provider"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MigrationFunc 迁移执行函数类型
type MigrationFunc func(tenantID int, db *gorm.DB) error

// globalMigrationFunc 全局迁移函数（由服务端设置）
var globalMigrationFunc MigrationFunc
var globalMigrationMu sync.RWMutex

// SetMigrationFunc 设置迁移执行函数
func SetMigrationFunc(fn MigrationFunc) {
	globalMigrationMu.Lock()
	globalMigrationFunc = fn
	globalMigrationMu.Unlock()
}

// InitStats 初始化统计信息
type InitStats struct {
	Total    int            // 总租户数
	Success  int            // 成功初始化数
	Failed   int            // 失败数
	Failures map[int]string // 失败详情：租户ID -> 错误信息
	mu       sync.RWMutex
}

// RecordSuccess 记录成功
func (s *InitStats) RecordSuccess() {
	s.mu.Lock()
	s.Success++
	s.mu.Unlock()
}

// RecordFailure 记录失败
func (s *InitStats) RecordFailure(tenantID int, reason string) {
	s.mu.Lock()
	s.Failed++
	if s.Failures == nil {
		s.Failures = make(map[int]string)
	}
	s.Failures[tenantID] = reason
	s.mu.Unlock()
}

// buildDSN 根据数据库类型构建连接字符串
func buildDSN(dbConfig *provider.ServiceDatabaseConfig) string {
	switch dbConfig.Driver {
	case "mysql":
		// MySQL DSN（Command 服务使用）
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	case "postgres":
		// PostgreSQL DSN（Query 服务使用）
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=1&TimeZone=Asia/Shanghai",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	default:
		log.Fatalf("[tenantdb] 不支持的数据库驱动: %s", dbConfig.Driver)
		return ""
	}
}

// connectWithRetry 带重试的数据库连接
func connectWithRetry(driver, dsn string, maxRetry int, interval time.Duration) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i < maxRetry; i++ {
		var db *gorm.DB
		var err error

		switch driver {
		case "mysql":
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		case "postgres":
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		default:
			return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
		}

		if err == nil {
			// 配置连接池
			sqlDB, _ := db.DB()
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetConnMaxIdleTime(60 * time.Second)
			sqlDB.SetConnMaxLifetime(3600 * time.Second)
			return db, nil
		}

		lastErr = err
		log.Printf("[tenantdb] 数据库连接失败 (尝试 %d/%d): %v", i+1, maxRetry, err)
		time.Sleep(interval)
	}

	return nil, fmt.Errorf("连接数据库失败 (重试 %d 次后): %w", maxRetry, lastErr)
}

// InitializeTenant 初始化单个租户的数据库连接
//
// 根据 serviceCode 区分数据库类型（MySQL 或 PostgreSQL），统一使用 SetTenantDB() 注册
func InitializeTenant(tenantID int, cache *Cache, migrationFn MigrationFunc) (*gorm.DB, error) {
	// 1. 获取数据库配置
	dbConfig, err := cache.GetDatabaseConfig(tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取数据库配置失败: %w", err)
	}

	// 2. 构建 DSN
	dsn := buildDSN(dbConfig)

	// 3. 建立数据库连接（带重试）
	db, err := connectWithRetry(dbConfig.Driver, dsn, 3, 5*time.Second)
	if err != nil {
		return nil, err
	}

	// 4. 注册到 sdk.Runtime（统一使用 SetTenantDB）
	sdk.Runtime.SetTenantDB(tenantID, db)
	dbType := "MySQL"
	if dbConfig.Driver == "postgres" {
		dbType = "PostgreSQL"
	}
	log.Printf("[tenantdb] 租户 %d TenantDB 初始化成功 (%s, serviceCode=%s)", tenantID, dbType, cache.GetServiceCode())

	// 5. 执行数据库迁移（仅 dev 环境）
	if config.ApplicationConfig.Mode == "dev" && migrationFn != nil {
		log.Printf("[tenantdb] 租户 %d 开始执行数据库迁移 (serviceCode=%s)", tenantID, cache.GetServiceCode())
		if err := migrationFn(tenantID, db); err != nil {
			return nil, fmt.Errorf("[tenantdb] 租户 %d 数据库迁移失败: %w", tenantID, err)
		}
		log.Printf("[tenantdb] 租户 %d 数据库迁移完成", tenantID)
	}

	return db, nil
}

// InitializeAll 批量初始化所有租户
func InitializeAll(cache *Cache, migrationFn MigrationFunc) (*InitStats, error) {
	tenantIDs := cache.GetTenantIDs()
	total := len(tenantIDs)

	if total == 0 {
		log.Println("[tenantdb] 没有租户需要初始化")
		return &InitStats{Total: 0}, nil
	}

	log.Printf("[tenantdb] 开始初始化 %d 个租户的数据库连接...", total)

	stats := &InitStats{
		Total:    total,
		Failures: make(map[int]string),
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数为 5

	for _, tenantID := range tenantIDs {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(tid int) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			if _, err := InitializeTenant(tid, cache, migrationFn); err != nil {
				stats.RecordFailure(tid, err.Error())
				log.Printf("[tenantdb] 租户 %d 数据库初始化失败: %v", tid, err)
			} else {
				stats.RecordSuccess()
			}
		}(tenantID)
	}

	wg.Wait()

	log.Printf("[tenantdb] 租户数据库初始化完成: 成功=%d, 失败=%d, 总计=%d",
		stats.Success, stats.Failed, stats.Total)

	if stats.Failed > 0 {
		for tid, reason := range stats.Failures {
			log.Printf("[tenantdb] 租户 %d 失败原因: %s", tid, reason)
		}
	}

	return stats, nil
}
