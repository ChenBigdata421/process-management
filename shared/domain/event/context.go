package event

import (
	"context"

	jxtevent "github.com/ChenBigdata421/jxt-core/sdk/pkg/domain/event"
)

// 定义上下文键类型，避免与其他包的键冲突
type contextKey string

// 定义用于存储事件的上下文键
const eventsKey contextKey = "domain_events"

// GetEventsFromContext 从上下文中获取领域事件列表
// 如果上下文中没有事件，则返回nil
// 注意：已迁移到 jxt-core，使用 EnterpriseEvent 接口
func GetEventsFromContext(ctx context.Context) []jxtevent.EnterpriseEvent {
	if ctx == nil {
		return nil
	}

	value := ctx.Value(eventsKey)
	if value == nil {
		return nil
	}

	events, ok := value.([]jxtevent.EnterpriseEvent)
	if !ok {
		return nil
	}

	return events
}

// SetEventsToContext 将领域事件列表存储到上下文中
// 返回包含事件的新上下文
// 注意：已迁移到 jxt-core，使用 EnterpriseEvent 接口
func SetEventsToContext(ctx context.Context, events []jxtevent.EnterpriseEvent) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, eventsKey, events)
}

// AddEventToContext 向上下文中添加一个领域事件
// 如果上下文中已有事件列表，则将新事件追加到列表中
// 如果上下文中没有事件列表，则创建一个新列表
// 注意：由于 context.Context 是不可变的，所以我们必须返回新的上下文
// 注意：已迁移到 jxt-core，使用 EnterpriseEvent 接口
func AddEventToContext(ctx context.Context, event jxtevent.EnterpriseEvent) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	events := GetEventsFromContext(ctx)
	if events == nil {
		events = make([]jxtevent.EnterpriseEvent, 0)
	}

	events = append(events, event)
	return SetEventsToContext(ctx, events)
}

// ClearEventsFromContext 清除上下文中的所有领域事件
// 返回不包含事件的新上下文
func ClearEventsFromContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return context.WithValue(ctx, eventsKey, nil)
}
