// shared/common/tenantdb/watcher.go
package tenantdb

import (
	"log"
	"sync"
	"time"
)

// TenantWatcher 租户监听器
type TenantWatcher struct {
	cache         *Cache
	stopCh        chan struct{}
	wg            sync.WaitGroup
	maxConcurrent int           // 最大并发连接数（限流）
	semaphore     chan struct{} // 信号量
	retryTimes    int           // 重试次数
	retryInterval time.Duration // 重试间隔
	checkInterval time.Duration // 检查间隔
	knownTenants  map[int]struct{}
	mu            sync.RWMutex
	onTenantAdded  func(tenantID int) error // 新租户添加回调
}

// globalWatcher 全局监听器实例
var globalWatcher *TenantWatcher
var watcherMu sync.RWMutex

// NewTenantWatcher 创建租户监听器
func NewTenantWatcher(cache *Cache, config *WatcherConfig) *TenantWatcher {
	if config == nil {
		config = DefaultWatcherConfig()
	}

	return &TenantWatcher{
		cache:         cache,
		stopCh:        make(chan struct{}),
		maxConcurrent: config.MaxConcurrent,
		semaphore:     make(chan struct{}, config.MaxConcurrent),
		retryTimes:    config.RetryTimes,
		retryInterval: config.RetryInterval,
		checkInterval: 30 * time.Second, // 默认 30 秒检查一次
		knownTenants:  make(map[int]struct{}),
		onTenantAdded: config.OnTenantAdded,
	}
}

// Start 启动监听
func (w *TenantWatcher) Start() {
	log.Println("[tenantdb] 启动动态租户监听器...")

	// 初始化已知租户列表
	w.updateKnownTenants()

	w.wg.Add(1)
	go w.watchLoop()

	log.Printf("[tenantdb] 动态租户监听器启动成功，每%d秒检查新租户", int(w.checkInterval.Seconds()))
}

// Stop 停止监听
func (w *TenantWatcher) Stop() {
	log.Println("[tenantdb] 停止动态租户监听器...")
	close(w.stopCh)
	w.wg.Wait()
	log.Println("[tenantdb] 动态租户监听器已停止")
}

// watchLoop 监听循环
func (w *TenantWatcher) watchLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.checkNewTenants()
		}
	}
}

// updateKnownTenants 更新已知租户列表
func (w *TenantWatcher) updateKnownTenants() {
	w.mu.Lock()
	defer w.mu.Unlock()

	tenantIDs := w.cache.GetTenantIDs()
	w.knownTenants = make(map[int]struct{})
	for _, id := range tenantIDs {
		w.knownTenants[id] = struct{}{}
	}
}

// checkNewTenants 检查新租户
func (w *TenantWatcher) checkNewTenants() {
	currentTenants := w.cache.GetTenantIDs()

	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, tenantID := range currentTenants {
		if _, exists := w.knownTenants[tenantID]; !exists {
			// 发现新租户
			log.Printf("[tenantdb] 检测到新租户 %d，开始初始化数据库连接", tenantID)
			w.handleNewTenant(tenantID)
			w.knownTenants[tenantID] = struct{}{}
		}
	}
}

// handleNewTenant 处理新租户
func (w *TenantWatcher) handleNewTenant(tenantID int) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		// 获取信号量（限流）
		w.semaphore <- struct{}{}
		defer func() { <-w.semaphore }()

		// 带重试的初始化
		var lastErr error
		for i := 0; i < w.retryTimes; i++ {
			if _, err := InitializeTenant(tenantID, w.cache, nil); err != nil {
				lastErr = err
				log.Printf("[tenantdb] 新租户 %d 数据库连接失败 (尝试 %d/%d): %v",
					tenantID, i+1, w.retryTimes, err)
				time.Sleep(w.retryInterval)
				continue
			}
			// 数据库初始化成功
			log.Printf("[tenantdb] 新租户 %d 数据库连接成功", tenantID)

			// 调用自定义回调（如 Casbin 策略初始化）
			if w.onTenantAdded != nil {
				if err := w.onTenantAdded(tenantID); err != nil {
					log.Printf("[tenantdb] 新租户 %d 自定义初始化失败: %v", tenantID, err)
				}
			}
			return
		}

		log.Printf("[tenantdb] 新租户 %d 数据库连接失败，等待下次检查周期重试: %v",
			tenantID, lastErr)
	}()
}

// StartWatcher 启动全局动态租户监听器
func StartWatcher(cache *Cache, config *WatcherConfig) error {
	if cache == nil {
		return nil
	}

	watcherMu.Lock()
	defer watcherMu.Unlock()

	if globalWatcher != nil {
		log.Println("[tenantdb] 全局监听器已存在，跳过启动")
		return nil
	}

	globalWatcher = NewTenantWatcher(cache, config)
	globalWatcher.Start()

	return nil
}

// StopWatcher 停止全局监听器
func StopWatcher() {
	watcherMu.Lock()
	defer watcherMu.Unlock()

	if globalWatcher != nil {
		globalWatcher.Stop()
		globalWatcher = nil
	}
}
