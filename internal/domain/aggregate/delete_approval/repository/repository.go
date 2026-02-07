package delete_approval_repository

import (
	"context"

	delete_approval "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// MediaDeleteApprovalRepository 媒体下载审批仓储接口
type MediaDeleteApprovalRepository interface {
	// Save 保存审批记录
	Save(ctx context.Context, approval *delete_approval.MediaDeleteApproval) error

	// Update 更新审批记录
	Update(ctx context.Context, approval *delete_approval.MediaDeleteApproval) error

	// FindByID 根据ID查找审批记录
	FindByID(ctx context.Context, id valueobject.DeleteApprovalID) (*delete_approval.MediaDeleteApproval, error)

	// FindByMediaAndUser 根据媒体ID和用户ID查找最新的审批记录
	FindByMediaAndUser(ctx context.Context, mediaID valueobject.MediaID, userID int64) (*delete_approval.MediaDeleteApproval, error)

	// FindByInstanceID 根据工作流实例ID查找审批记录
	FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*delete_approval.MediaDeleteApproval, error)

	// FindByMediaIDs 批量查询多个媒体的审批状态
	FindByMediaIDs(ctx context.Context, mediaIDs []valueobject.MediaID, userID int64) ([]*delete_approval.MediaDeleteApproval, error)

	// FindPendingByUser 查找用户的待审批记录
	FindPendingByUser(ctx context.Context, userID int64) ([]*delete_approval.MediaDeleteApproval, error)
}
