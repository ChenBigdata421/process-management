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
	ApprovalStatusExpired  ApprovalStatus = "expired"  // 已过期
	ApprovalStatusNone     ApprovalStatus = "none"
)

// MediaDownloadApproval 媒体下载审批记录
type MediaDownloadApproval struct {
	ID             valueobject.DownloadApprovalID `json:"id" gorm:"primaryKey;type:uuid;comment:主键ID"`
	MediaID        valueobject.MediaID            `json:"mediaId" gorm:"column:media_id;type:uuid;index:idx_media_user;comment:媒体ID"`
	UserID         int64                          `json:"userId" gorm:"column:user_id;type:bigint;not null;index:idx_media_user;comment:申请人ID"`
	InstanceID     valueobject.InstanceID         `json:"instanceId" gorm:"column:instance_id;type:uuid;index:idx_instance;comment:工作流实例ID"`
	TaskID         valueobject.TaskID             `json:"taskId" gorm:"column:task_id;type:uuid;comment:当前任务ID"`
	Status         ApprovalStatus                 `json:"status" gorm:"column:status;type:varchar(20);not null;default:pending;index:idx_status;comment:审批状态"`
	ApplyReason    string                         `json:"applyReason" gorm:"column:apply_reason;type:varchar(500);comment:申请原因"`
	RejectReason   string                         `json:"rejectReason" gorm:"column:reject_reason;type:varchar(500);comment:驳回原因"`
	ApprovedAt     *time.Time                     `json:"approvedAt" gorm:"column:approved_at;comment:审批通过时间"`
	ExpiredAt      *time.Time                     `json:"expiredAt" gorm:"column:expired_at;comment:过期时间"`
	DownloadCount  int                            `json:"downloadCount" gorm:"column:download_count;default:0;comment:下载次数"`
	LastDownloadAt *time.Time                     `json:"lastDownloadAt" gorm:"column:last_download_at;comment:最后下载时间"`

	models.ModelTime
}

// TableName 指定表名
func (MediaDownloadApproval) TableName() string {
	return "media_download_approval"
}

// NewMediaDownloadApproval 创建新的下载审批记录
func NewMediaDownloadApproval(userID int64, cmd *command.SubmitDownloadApprovalCommand) *MediaDownloadApproval {
	now := time.Now()
	return &MediaDownloadApproval{
		ID:          valueobject.NewDownloadApprovalID(),
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
func (m *MediaDownloadApproval) SetInstanceID(instanceID valueobject.InstanceID) {
	m.InstanceID = instanceID
}

// SetTaskID 设置当前任务ID
func (m *MediaDownloadApproval) SetTaskID(taskID valueobject.TaskID) {
	m.TaskID = taskID
}

// Approve 审批通过
func (m *MediaDownloadApproval) Approve(validDays int) {
	now := time.Now()
	expiredAt := now.AddDate(0, 0, validDays) // 默认有效期
	m.Status = ApprovalStatusApproved
	m.ApprovedAt = &now
	m.ExpiredAt = &expiredAt
	m.UpdatedAt = now
}

// Reject 审批驳回
func (m *MediaDownloadApproval) Reject(reason string) {
	now := time.Now()
	m.Status = ApprovalStatusRejected
	m.RejectReason = reason
	m.UpdatedAt = now
}

// Pending 审批中
func (m *MediaDownloadApproval) Pending() {
	now := time.Now()
	m.Status = ApprovalStatusPending
	m.UpdatedAt = now
}

// 置空
func (m *MediaDownloadApproval) None() {
	now := time.Now()
	m.Status = ApprovalStatusNone
	m.UpdatedAt = now
}

// MarkExpired 标记过期
func (m *MediaDownloadApproval) MarkExpired() {
	now := time.Now()
	m.Status = ApprovalStatusExpired
	m.UpdatedAt = now
}

// IncrementDownloadCount 增加下载次数
func (m *MediaDownloadApproval) IncrementDownloadCount() {
	now := time.Now()
	m.DownloadCount++
	m.LastDownloadAt = &now
	m.UpdatedAt = now
}

// CanDownload 判断是否可以下载
func (m *MediaDownloadApproval) CanDownload() bool {
	if m.Status != ApprovalStatusApproved {
		return false
	}
	// 检查是否过期
	if m.ExpiredAt != nil && time.Now().After(*m.ExpiredAt) {
		return false
	}
	return true
}

// IsExpired 判断是否已过期
func (m *MediaDownloadApproval) IsExpired() bool {
	if m.Status == ApprovalStatusExpired {
		return true
	}
	if m.Status == ApprovalStatusApproved && m.ExpiredAt != nil && time.Now().After(*m.ExpiredAt) {
		return true
	}
	return false
}
