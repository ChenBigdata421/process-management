package service

import (
	"context"
	"log"

	"jxt-evidence-system/process-management/internal/application/command"
	delete_approval "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval"
	delete_approval_repository "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval/repository"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// DeleteApprovalService 删除审批服务
type DeleteApprovalService struct {
	approvalRepo delete_approval_repository.MediaDeleteApprovalRepository
}

// NewDeleteApprovalService 创建删除审批服务
func NewDeleteApprovalService(approvalRepo delete_approval_repository.MediaDeleteApprovalRepository) *DeleteApprovalService {
	return &DeleteApprovalService{
		approvalRepo: approvalRepo,
	}
}

// GetApprovalStatus 获取删除审批状态
func (s *DeleteApprovalService) GetApprovalStatus(ctx context.Context, mediaID valueobject.MediaID, userID int64) (*command.DeleteApprovalStatusDTO, error) {
	approval, err := s.approvalRepo.FindByMediaAndUser(ctx, mediaID, userID)
	if err != nil {
		return nil, err
	}

	// 没有找到记录
	if approval == nil {
		return &command.DeleteApprovalStatusDTO{
			Status: "none",
		}, nil
	}

	dto := &command.DeleteApprovalStatusDTO{
		ApprovalID:   approval.ID,
		Status:       string(approval.Status),
		RejectReason: approval.RejectReason,
	}

	if approval.ApprovedAt != nil {
		dto.ApprovedAt = *approval.ApprovedAt
	}
	if approval.Status == delete_approval.ApprovalStatusRejected {
		dto.RejectReason = approval.RejectReason
	}

	return dto, nil
}

// SubmitApproval 提交下载审批申请
func (s *DeleteApprovalService) SubmitApproval(ctx context.Context, userID int64, cmd *command.SubmitDeleteApprovalCommand) (*delete_approval.MediaDeleteApproval, error) {
	// 检查是否已有待审批的记录
	existing, err := s.approvalRepo.FindByMediaAndUser(ctx, cmd.MediaID, userID)
	if err != nil {
		return nil, err
	}

	// 如果已有待审批记录，返回现有记录
	if existing != nil && existing.Status == delete_approval.ApprovalStatusPending {
		return existing, nil
	}

	// 创建新的审批记录
	approval := delete_approval.NewMediaDeleteApproval(userID, cmd)
	if err := s.approvalRepo.Save(ctx, approval); err != nil {
		return nil, err
	}

	log.Printf("[DeleteApprovalService] Created delete approval: id=%s, mediaId=%d, userId=%d", approval.ID.String(), cmd.MediaID, userID)
	return approval, nil
}

// UpdateApprovalInstance 更新审批记录的工作流实例ID
func (s *DeleteApprovalService) UpdateApprovalInstance(ctx context.Context, approvalID valueobject.DeleteApprovalID, instanceID valueobject.InstanceID, taskID valueobject.TaskID) error {
	log.Printf("[DeleteApprovalService] UpdateApprovalInstance called: approvalID=%s, instanceID=%s, taskID=%s", approvalID.String(), instanceID.String(), taskID.String())

	approval, err := s.approvalRepo.FindByID(ctx, approvalID)
	if err != nil {
		log.Printf("[DeleteApprovalService] FindByID failed: %v", err)
		return err
	}

	log.Printf("[DeleteApprovalService] Found approval: id=%s, mediaId=%d, userId=%d, currentInstanceId=%s", approval.ID.String(), approval.MediaID, approval.UserID, approval.InstanceID.String())

	approval.SetInstanceID(instanceID)
	approval.SetTaskID(taskID)

	log.Printf("[DeleteApprovalService] Updated approval: instanceId=%s, taskId=%s", approval.InstanceID.String(), approval.TaskID.String())

	err = s.approvalRepo.Update(ctx, approval)
	if err != nil {
		log.Printf("[DeleteApprovalService] Update failed: %v", err)
		return err
	}

	log.Printf("[DeleteApprovalService] Update succeeded")
	return nil
}

// ApproveDelete 审批通过
func (s *DeleteApprovalService) ApproveDelete(ctx context.Context, instanceID valueobject.InstanceID) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.Approve()
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DeleteApprovalService] Approved delete: id=%d, mediaId=%d", approval.ID, approval.MediaID)
	return nil
}

func (s *DeleteApprovalService) PendingDelete(ctx context.Context, instanceID valueobject.InstanceID) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.Pending()
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DeleteApprovalService] Pending delete: id=%d, mediaId=%d", approval.ID, approval.MediaID)
	return nil
}

func (s *DeleteApprovalService) NoneDelete(ctx context.Context, instanceID valueobject.InstanceID) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.None()
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DeleteApprovalService] None delete: id=%d, mediaId=%d", approval.ID, approval.MediaID)
	return nil
}

// RejectDelete 审批驳回
func (s *DeleteApprovalService) RejectDelete(ctx context.Context, instanceID valueobject.InstanceID, reason string) error {
	approval, err := s.approvalRepo.FindByInstanceID(ctx, instanceID)
	if err != nil {
		return err
	}

	approval.Reject(reason)
	if err := s.approvalRepo.Update(ctx, approval); err != nil {
		return err
	}

	log.Printf("[DeleteApprovalService] Rejected delete: id=%d, mediaId=%d, reason=%s", approval.ID, approval.MediaID, reason)
	return nil
}

// GetApprovalByInstanceID 根据工作流实例ID获取审批记录
func (s *DeleteApprovalService) GetApprovalByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*delete_approval.MediaDeleteApproval, error) {
	return s.approvalRepo.FindByInstanceID(ctx, instanceID)
}
