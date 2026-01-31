package service

import (
	"context"
	"log"

	"jxt-evidence-system/process-management/internal/application/command"
	download_approval "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval"
	download_approval_repository "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval/repository"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// DownloadApprovalService 下载审批服务
type DownloadApprovalService struct {
	approvalRepo download_approval_repository.MediaDownloadApprovalRepository
}

// NewDownloadApprovalService 创建下载审批服务
func NewDownloadApprovalService(approvalRepo download_approval_repository.MediaDownloadApprovalRepository) *DownloadApprovalService {
	return &DownloadApprovalService{
		approvalRepo: approvalRepo,
	}
}

// GetApprovalStatus 获取下载审批状态
func (s *DownloadApprovalService) GetApprovalStatus(ctx context.Context, mediaID valueobject.MediaID, userID int64) (*command.DownloadApprovalStatusDTO, error) {
	approval, err := s.approvalRepo.FindByMediaAndUser(ctx, mediaID, userID)
	if err != nil {
		return nil, err
	}

	// 没有找到记录
	if approval == nil {
		return &command.DownloadApprovalStatusDTO{
			CanDownload:   false,
			Status:        "none",
			DownloadCount: 0,
		}, nil
	}

	// 检查是否过期
	if approval.IsExpired() && approval.Status == download_approval.ApprovalStatusApproved {
		// 更新状态为过期
		approval.MarkExpired()
		if err := s.approvalRepo.Update(ctx, approval); err != nil {
			log.Printf("[DownloadApprovalService] Failed to update expired status: %v", err)
		}
	}

	dto := &command.DownloadApprovalStatusDTO{
		ApprovalID:    approval.ID,
		CanDownload:   approval.CanDownload(),
		Status:        string(approval.Status),
		RejectReason:  approval.RejectReason,
		DownloadCount: approval.DownloadCount,
	}

	if approval.ApprovedAt != nil {
		dto.ApprovedAt = *approval.ApprovedAt
	}
	if approval.ExpiredAt != nil {
		dto.ExpiredAt = *approval.ExpiredAt
	}
	if approval.Status == download_approval.ApprovalStatusRejected {
		dto.RejectReason = approval.RejectReason
	}

	return dto, nil
}

// SubmitApproval 提交下载审批申请
func (s *DownloadApprovalService) SubmitApproval(ctx context.Context, userID int64, cmd *command.SubmitDownloadApprovalCommand) (*download_approval.MediaDownloadApproval, error) {
	// 检查是否已有待审批的记录
	existing, err := s.approvalRepo.FindByMediaAndUser(ctx, cmd.MediaID, userID)
	if err != nil {
		return nil, err
	}

	// 如果已有待审批记录，返回现有记录
	if existing != nil && existing.Status == download_approval.ApprovalStatusPending {
		return existing, nil
	}

	// 如果已审批通过且未过期，返回现有记录
	if existing != nil && existing.CanDownload() {
		return existing, nil
	}

	// 创建新的审批记录
	approval := download_approval.NewMediaDownloadApproval(userID, cmd)
	if err := s.approvalRepo.Save(ctx, approval); err != nil {
		return nil, err
	}

	log.Printf("[DownloadApprovalService] Created download approval: id=%s, mediaId=%d, userId=%d", approval.ID.String(), cmd.MediaID, userID)
	return approval, nil
}

// UpdateApprovalInstance 更新审批记录的工作流实例ID
func (s *DownloadApprovalService) UpdateApprovalInstance(ctx context.Context, approvalID valueobject.DownloadApprovalID, instanceID valueobject.InstanceID, taskID valueobject.TaskID) error {
	log.Printf("[DownloadApprovalService] UpdateApprovalInstance called: approvalID=%s, instanceID=%s, taskID=%s", approvalID.String(), instanceID.String(), taskID.String())

	approval, err := s.approvalRepo.FindByID(ctx, approvalID)
	if err != nil {
		log.Printf("[DownloadApprovalService] FindByID failed: %v", err)
		return err
	}

	log.Printf("[DownloadApprovalService] Found approval: id=%s, mediaId=%d, userId=%d, currentInstanceId=%s", approval.ID.String(), approval.MediaID, approval.UserID, approval.InstanceID.String())

	approval.SetInstanceID(instanceID)
	approval.SetTaskID(taskID)

	log.Printf("[DownloadApprovalService] Updated approval: instanceId=%s, taskId=%s", approval.InstanceID.String(), approval.TaskID.String())

	err = s.approvalRepo.Update(ctx, approval)
	if err != nil {
		log.Printf("[DownloadApprovalService] Update failed: %v", err)
		return err
	}

	log.Printf("[DownloadApprovalService] Update succeeded")
	return nil
}

// ApproveDownload 审批通过
func (s *DownloadApprovalService) ApproveDownload(ctx context.Context, instanceID valueobject.InstanceID, validDays int) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	if validDays <= 0 {
		validDays = 30 // 默认30天有效期
	}

	approval.Approve(validDays)
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DownloadApprovalService] Approved download: id=%d, mediaId=%d, expiredAt=%v", approval.ID, approval.MediaID, approval.ExpiredAt)
	return nil
}

func (s *DownloadApprovalService) PendingDownload(ctx context.Context, instanceID valueobject.InstanceID) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.Pending()
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DownloadApprovalService] Pending download: id=%d, mediaId=%d", approval.ID, approval.MediaID)
	return nil
}

func (s *DownloadApprovalService) NoneDownload(ctx context.Context, instanceID valueobject.InstanceID) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.None()
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DownloadApprovalService] None download: id=%d, mediaId=%d", approval.ID, approval.MediaID)
	return nil
}

// RejectDownload 审批驳回
func (s *DownloadApprovalService) RejectDownload(ctx context.Context, instanceID valueobject.InstanceID, reason string) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.Reject(reason)
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DownloadApprovalService] Rejected download: id=%d, mediaId=%d, reason=%s", approval.ID, approval.MediaID, reason)
	return nil
}

// RecordDownload 记录下载
func (s *DownloadApprovalService) RecordDownload(ctx context.Context, mediaID valueobject.MediaID, userID int64) error {
	approval, err := s.approvalRepo.FindByMediaAndUser(ctx, mediaID, userID)
	if err != nil {
		return err
	}

	if approval == nil || !approval.CanDownload() {
		return nil // 没有审批记录或不能下载，不记录
	}

	approval.IncrementDownloadCount()
	return s.approvalRepo.Update(ctx, approval)
}

// BatchGetApprovalStatus 批量获取下载审批状态
func (s *DownloadApprovalService) BatchGetApprovalStatus(ctx context.Context, mediaIDs []valueobject.MediaID, userID int64) ([]*command.BatchDownloadApprovalStatusDTO, error) {
	approvals, err := s.approvalRepo.FindByMediaIDs(ctx, mediaIDs, userID)
	if err != nil {
		return nil, err
	}

	// 创建mediaID到审批记录的映射
	approvalMap := make(map[valueobject.MediaID]*download_approval.MediaDownloadApproval)
	for _, a := range approvals {
		approvalMap[a.MediaID] = a
	}

	// 构建结果
	results := make([]*command.BatchDownloadApprovalStatusDTO, 0, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		dto := &command.BatchDownloadApprovalStatusDTO{
			MediaID:       mediaID,
			CanDownload:   false,
			Status:        "none",
			DownloadCount: 0,
		}

		if approval, ok := approvalMap[mediaID]; ok {
			// 检查是否过期
			if approval.IsExpired() && approval.Status == download_approval.ApprovalStatusApproved {
				approval.MarkExpired()
				_ = s.approvalRepo.Update(ctx, approval)
			}

			dto.ApprovalID = approval.ID
			dto.CanDownload = approval.CanDownload()
			dto.Status = string(approval.Status)
			dto.DownloadCount = approval.DownloadCount
		}

		results = append(results, dto)
	}

	return results, nil
}

// GetApprovalByInstanceID 根据工作流实例ID获取审批记录
func (s *DownloadApprovalService) GetApprovalByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*download_approval.MediaDownloadApproval, error) {
	return s.approvalRepo.FindByInstanceID(ctx, instanceID)
}
