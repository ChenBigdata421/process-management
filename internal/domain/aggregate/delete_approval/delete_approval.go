package download_approval_aggregate

import (
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/models"
)

// ApprovalStatus 审批状态
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"  // 审批中
	ApprovalStatusApproved ApprovalStatus = "approved" // 已通过
	ApprovalStatusRejected ApprovalStatus = "rejected" // 已驳回
	ApprovalStatusNone     ApprovalStatus = "none"
)

// MediaDeleteApproval 媒体下载审批记录
type MediaDeleteApproval struct {
	ID           valueobject.DeleteApprovalID `json:"id" gorm:"primaryKey;type:uuid;comment:主键ID"`
	MediaID      valueobject.MediaID          `json:"mediaId" gorm:"column:media_id;type:uuid;index:idx_media_user;comment:媒体ID"`
	UserID       int64                        `json:"userId" gorm:"column:user_id;type:bigint;not null;index:idx_media_user;comment:申请人ID"`
	InstanceID   valueobject.InstanceID       `json:"instanceId" gorm:"column:instance_id;type:uuid;index:idx_instance;comment:工作流实例ID"`
	TaskID       valueobject.TaskID           `json:"taskId" gorm:"column:task_id;type:uuid;comment:当前任务ID"`
	Status       ApprovalStatus               `json:"status" gorm:"column:status;type:varchar(20);not null;default:pending;index:idx_status;comment:审批状态"`
	ApplyReason  string                       `json:"applyReason" gorm:"column:apply_reason;type:varchar(500);comment:申请原因"`
	RejectReason string                       `json:"rejectReason" gorm:"column:reject_reason;type:varchar(500);comment:驳回原因"`
	ApprovedAt   *time.Time                   `json:"approvedAt" gorm:"column:approved_at;comment:审批通过时间"`
	models.ModelTime
}

// TableName 指定表名
func (MediaDeleteApproval) TableName() string {
	return "media_delete_approval"
}

// NewMediaDeleteApproval 创建新的下载审批记录
func NewMediaDeleteApproval(userID int64, cmd *command.SubmitDeleteApprovalCommand) *MediaDeleteApproval {
	now := time.Now()
	return &MediaDeleteApproval{
		ID:          valueobject.NewDeleteApprovalID(),
		MediaID:     cmd.MediaID,
		UserID:      userID,
		Status:      ApprovalStatusPending,
		InstanceID:  cmd.InstanceID,
		TaskID:      cmd.TaskID,
		ApplyReason: cmd.Reason,
		ModelTime: models.ModelTime{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// SetInstanceID 设置工作流实例ID
func (m *MediaDeleteApproval) SetInstanceID(instanceID valueobject.InstanceID) {
	m.InstanceID = instanceID
}

// SetTaskID 设置当前任务ID
func (m *MediaDeleteApproval) SetTaskID(taskID valueobject.TaskID) {
	m.TaskID = taskID
}

// Approve 审批通过
func (m *MediaDeleteApproval) Approve() {
	now := time.Now()
	m.Status = ApprovalStatusApproved
	m.ApprovedAt = &now
	m.UpdatedAt = now
}

// Reject 审批驳回
func (m *MediaDeleteApproval) Reject(reason string) {
	now := time.Now()
	m.Status = ApprovalStatusRejected
	m.RejectReason = reason
	m.UpdatedAt = now
}

// Pending 审批中
func (m *MediaDeleteApproval) Pending() {
	now := time.Now()
	m.Status = ApprovalStatusPending
	m.UpdatedAt = now
}

// 置空
func (m *MediaDeleteApproval) None() {
	now := time.Now()
	m.Status = ApprovalStatusNone
	m.UpdatedAt = now
}
