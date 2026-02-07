package outbox

import (
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
)

// NewTopicMapper 创建 Topic 映射器
// ⭐ 将聚合类型映射到 NATS JetStream 的 Subject
//
// Topic 映射规则：
// - 格式：file-storage.{aggregateType}.events
// - 例如：file-storage.media.events
//
// 这个映射器用于：
// 1. 将 Outbox 中的事件映射到对应的 NATS Subject
// 2. 确保事件被发布到正确的 Subject
// 3. 支持事件的订阅和处理
func NewTopicMapper() outbox.TopicMapper {
	return outbox.NewMapBasedTopicMapper(map[string]string{
		"Task": "process.task.events",
	}, "process.default.events")
}
