package command

import (
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"time"
)

// SubmitDeleteApprovalCommand 提交删除审批申请命令
type SubmitDeleteApprovalCommand struct {
	MediaID    valueobject.MediaID    `uri:"mediaId" binding:"required"`
	InstanceID valueobject.InstanceID `json:"instanceId"`
	TaskID     valueobject.TaskID     `json:"taskId"`
	Reason     string                 `json:"reason"`
}

// GetDeleteApprovalCommand 查询删除审批状态命令
type GetDeleteApprovalStatusCommand struct {
	MediaID valueobject.MediaID `uri:"mediaId" binding:"required"`
}

// DeleteApprovalStatusDTO 删除审批状态DTO
type DeleteApprovalStatusDTO struct {
	ApprovalID   valueobject.DeleteApprovalID `json:"approvalId,omitempty"`
	Status       string                       `json:"status"` // none/pending/approved/rejected/expired
	ApprovedAt   time.Time                    `json:"approvedAt,omitempty"`
	RejectReason string                       `json:"rejectReason,omitempty"`
}

// SubmitDeleteApprovalResultDTO 提交删除审批结果DTO
type SubmitDeleteApprovalResultDTO struct {
	ApprovalID valueobject.DownloadApprovalID `json:"approvalId"`
	InstanceID valueobject.InstanceID         `json:"instanceId"`
	TaskID     string                         `json:"taskId"`
	Status     string                         `json:"status"`
}
