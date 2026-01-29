package download_approval_repository

import (
	"context"

	download_approval "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// MediaDownloadApprovalRepository 媒体下载审批仓储接口
type MediaDownloadApprovalRepository interface {
	// Save 保存审批记录
	Save(ctx context.Context, approval *download_approval.MediaDownloadApproval) error

	// Update 更新审批记录
	Update(ctx context.Context, approval *download_approval.MediaDownloadApproval) error

	// FindByID 根据ID查找审批记录
	FindByID(ctx context.Context, id valueobject.DownloadApprovalID) (*download_approval.MediaDownloadApproval, error)

	// FindByMediaAndUser 根据媒体ID和用户ID查找最新的审批记录
	FindByMediaAndUser(ctx context.Context, mediaID valueobject.MediaID, userID int64) (*download_approval.MediaDownloadApproval, error)

	// FindByInstanceID 根据工作流实例ID查找审批记录
	FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*download_approval.MediaDownloadApproval, error)

	// FindByMediaIDs 批量查询多个媒体的审批状态
	FindByMediaIDs(ctx context.Context, mediaIDs []valueobject.MediaID, userID int64) ([]*download_approval.MediaDownloadApproval, error)

	// FindPendingByUser 查找用户的待审批记录
	FindPendingByUser(ctx context.Context, userID int64) ([]*download_approval.MediaDownloadApproval, error)

	// UpdateExpiredRecords 更新过期的审批记录
	UpdateExpiredRecords(ctx context.Context) (int64, error)
}
