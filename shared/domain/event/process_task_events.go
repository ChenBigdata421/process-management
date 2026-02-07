package event

import (
	"time"
)

// 文书事件类型定义
const (
	EventTypeProcessMediaDeleted = "MediaDeleted" // 删除
)

// MediaDeletedPayload 媒体删除事件
type ProcessMediaDeletedPayload struct {
	MediaID   string    `json:"mediaId"`
	DeleteBy  int       `json:"deleteBy"`
	DeletedAt time.Time `json:"deletedAt"`
}

// NewMediaDeletedEvent 创建媒体删除事件
func NewProcessMediaDeletedEvent(Id string, mediaID string, deleteBy int, deletedAt time.Time) Event {

	return NewDomainEvent(
		EventTypeProcessMediaDeleted,
		Id,
		"Task",
		ProcessMediaDeletedPayload{
			MediaID:   mediaID,
			DeleteBy:  deleteBy,
			DeletedAt: deletedAt,
		},
	)
}
