package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tenantdb "jxt-evidence-system/process-management/shared/common/tenantdb"

	"github.com/ChenBigdata421/jxt-core/sdk"
	mycasbin "github.com/ChenBigdata421/jxt-core/sdk/pkg/casbin"
)

// Setup 配置命令数据库,支持多数据库连接,k可以理解成租户id
func ProcessDbSetup() {
	// 优先从租户缓存获取配置
	tenantCache := tenantdb.GetGlobalCache()
	if tenantCache != nil {
		// 使用 ETCD 租户缓存模式
		log.Println("[database] 使用租户缓存初始化 Command 数据库")

		// 获取迁移函数
		migrationFn := getProcessMigrationFunc()

		stats, err := tenantdb.InitializeAll(tenantCache, migrationFn)
		if err != nil {
			log.Fatalf("[database] 租户数据库初始化失败: %v", err)
		}
		log.Printf("[database] Command 数据库初始化完成: 成功=%d, 失败=%d", stats.Success, stats.Failed)
		return
	}

	// 降级：使用静态配置
	// 注意：最新架构要求必须与 ETCD 保持一致，不再支持降级到静态配置
	log.Fatalf("[database] 错误：租户缓存未初始化，请确保 ETCD 可用")
}

// ============================================
// Casbin 策略初始化函数（gRPC 模式）
// ============================================

// SetupTenantCasbin 通过 gRPC 初始化指定租户的 Casbin 策略
// 用于单个租户的 Casbin 初始化（如动态添加新租户时）
func SetupTenantCasbin(provider interface{}, tenantID int) error {
	// 检查 provider 是否实现 PolicyProvider 接口
	// 这里使用类型断言来验证 provider 是否符合接口要求
	type policyProvider interface {
		GetPolicies(ctx context.Context, tenantID int) ([]mycasbin.PolicyRule, error)
	}

	pp, ok := provider.(policyProvider)
	if !ok {
		return fmt.Errorf("provider does not implement PolicyProvider interface")
	}

	e, err := mycasbin.SetupWithProvider(pp, tenantID)
	if err != nil {
		return fmt.Errorf("casbin init via provider failed (租户 %d): %w", tenantID, err)
	}
	sdk.Runtime.SetTenantCasbin(tenantID, e)
	log.Printf("[Casbin] 租户 %d 策略初始化成功", tenantID)
	return nil
}

// SetupAllTenantsCasbin 初始化所有租户的 Casbin 策略
// 用于服务启动时批量初始化所有现有租户
func SetupAllTenantsCasbin(provider interface{}, cache *tenantdb.Cache) error {
	// 创建带超时的Context（最多等待5分钟）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 检查 provider 是否实现 PolicyProvider 接口
	type policyProvider interface {
		GetPolicies(ctx context.Context, tenantID int) ([]mycasbin.PolicyRule, error)
	}

	pp, ok := provider.(policyProvider)
	if !ok {
		return fmt.Errorf("provider does not implement PolicyProvider interface")
	}

	tenantIDs := cache.GetTenantIDs()
	if len(tenantIDs) == 0 {
		log.Println("[Casbin] 无租户需要初始化")
		return nil
	}

	log.Printf("[Casbin] 开始初始化 %d 个租户的策略", len(tenantIDs))

	// 并发初始化
	sem := make(chan struct{}, 5) // 最多 5 个并发
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, failCount int

	for _, tenantID := range tenantIDs {
		// 检查Context是否已取消
		select {
		case <-ctx.Done():
			log.Printf("[Casbin] 初始化被取消: %v", ctx.Err())
			wg.Wait() // 等待已启动的goroutine完成
			return ctx.Err()
		default:
		}

		wg.Add(1)
		go func(tid int) {
			defer wg.Done()

			// 获取信号量（支持Context取消）
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			err := SetupTenantCasbin(pp, tid)
			mu.Lock()
			if err != nil {
				failCount++
				log.Printf("[Casbin] 租户 %d 初始化失败: %v", tid, err)
			} else {
				successCount++
				log.Printf("[Casbin] 租户 %d 初始化成功", tid)
			}
			mu.Unlock()
		}(tenantID)
	}

	wg.Wait()

	log.Printf("[Casbin] 初始化完成: 成功=%d, 失败=%d", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("部分租户 Casbin 初始化失败: %d/%d", failCount, len(tenantIDs))
	}
	return nil
}
