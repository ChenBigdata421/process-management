package websocket

import (
	"sync"

	websocket "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	"jxt-evidence-system/process-management/shared/common/di"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

var (
	registrations = make([]func(), 0)
	registerOnce  sync.Once
)

// jiyuanjie 负责依赖注入所有的 Repo 实例
func RegisterDependencies() {
	//遍历所有api的依赖注入方法
	registerOnce.Do(func() {
		for _, f := range registrations {
			f()
		}
	})
}

func init() {
	registrations = append(registrations, registerHubDependencies)
}

// Hub 的依赖注入
func registerHubDependencies() {
	if err := di.Provide(func() websocket.WebSocketNotifier {
		hub := NewHub()
		go hub.Run() // 在注册时启动 Hub
		return hub
	}); err != nil {
		logger.Fatalf("failed to provide Hub: %v", err)
	}
}
