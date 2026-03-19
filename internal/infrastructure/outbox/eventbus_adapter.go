package outbox

import (
	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
	outboxadapters "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters"
)

// NewEventBusAdapter 创建 EventBus 适配器
// 将 jxt-core EventBus 适配为 outbox.EventPublisher 接口
//
// 使用 jxt-core 提供的 EventBusAdapter，无需自己实现
func NewEventBusAdapter() outbox.EventPublisher {
	// 从 SDK Runtime 获取全局 EventBus 实例（NATS JetStream）
	eventBus := sdk.Runtime.GetEventBus()

	// 使用 jxt-core 提供的适配器 ⭐
	// 该适配器会自动：
	// 1. 转换 Outbox Envelope 为 EventBus Envelope
	// 2. 转换 EventBus PublishResult 为 Outbox PublishResult
	// 3. 启动 ACK 结果转换 goroutine
	return outboxadapters.NewEventBusAdapter(eventBus)
}
