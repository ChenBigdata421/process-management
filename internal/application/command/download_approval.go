package command

import (
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"time"
)

// SubmitDownloadApprovalCommand 提交下载审批申请命令
type SubmitDownloadApprovalCommand struct {
	MediaID    valueobject.MediaID    `uri:"mediaId" binding:"required"`
	InstanceID valueobject.InstanceID `json:"instanceId"`
	TaskID     valueobject.TaskID     `json:"taskId"`
	Reason     string                 `json:"reason"`
}

// BatchGetDownloadApprovalCommand 批量查询下载审批状态命令
type BatchGetDownloadApprovalStatusCommand struct {
	MediaIDs []valueobject.MediaID `json:"mediaIds" binding:"required"`
}

// GetDownloadApprovalCommand 批量查询下载审批状态命令
type GetDownloadApprovalStatusCommand struct {
	MediaID valueobject.MediaID `uri:"mediaId" binding:"required"`
}

// RecordDownloadCommand 记录下载命令
type RecordDownloadCommand struct {
	MediaID    valueobject.MediaID `uri:"mediaId" binding:"required"`
	ApprovalID int64               `json:"approvalId"`
}

// DownloadApprovalStatusDTO 下载审批状态DTO
type DownloadApprovalStatusDTO struct {
	ApprovalID    valueobject.DownloadApprovalID `json:"approvalId,omitempty"`
	CanDownload   bool                           `json:"canDownload"`
	Status        string                         `json:"status"` // none/pending/approved/rejected/expired
	ApprovedAt    time.Time                      `json:"approvedAt,omitempty"`
	ExpiredAt     time.Time                      `json:"expiredAt,omitempty"`
	DownloadCount int                            `json:"downloadCount"`
	RejectReason  string                         `json:"rejectReason,omitempty"`
}

// SubmitDownloadApprovalResultDTO 提交下载审批结果DTO
type SubmitDownloadApprovalResultDTO struct {
	ApprovalID valueobject.DownloadApprovalID `json:"approvalId"`
	InstanceID valueobject.InstanceID         `json:"instanceId"`
	TaskID     string                         `json:"taskId"`
	Status     string                         `json:"status"`
}

// BatchDownloadApprovalStatusDTO 批量下载审批状态DTO
type BatchDownloadApprovalStatusDTO struct {
	MediaID       valueobject.MediaID            `json:"mediaId"`
	CanDownload   bool                           `json:"canDownload"`
	Status        string                         `json:"status"`
	ApprovalID    valueobject.DownloadApprovalID `json:"approvalId,omitempty"`
	DownloadCount int                            `json:"downloadCount"`
}
