package event

import (
	jxtevent "github.com/ChenBigdata421/jxt-core/sdk/pkg/domain/event"
)

// Event 类型别名，用于向后兼容
type Event = jxtevent.EnterpriseEvent

// DomainEvent 类型别名，用于向后兼容
type DomainEvent = jxtevent.EnterpriseDomainEvent

// NewDomainEvent 创建领域事件的辅助函数
func NewDomainEvent(eventType string, aggregateID interface{}, aggregateType string, payload interface{}) Event {
	return jxtevent.NewEnterpriseDomainEvent(eventType, aggregateID, aggregateType, payload)
}
