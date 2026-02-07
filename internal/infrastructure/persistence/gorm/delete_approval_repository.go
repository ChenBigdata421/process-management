package persistence

import (
	"context"
	"log"

	delete_approval "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/errors"

	"gorm.io/gorm"
)

// GormMediaDeleteApprovalRepository GORM实现的媒体下载审批仓储
type GormMediaDeleteApprovalRepository struct {
	GormRepository
}

// Save 保存审批记录
func (r *GormMediaDeleteApprovalRepository) Save(ctx context.Context, approval *delete_approval.MediaDeleteApproval) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	return db.Create(approval).Error
}

// Update 更新审批记录
func (r *GormMediaDeleteApprovalRepository) Update(ctx context.Context, approval *delete_approval.MediaDeleteApproval) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	return db.Save(approval).Error
}

// FindByID 根据ID查找审批记录
func (r *GormMediaDeleteApprovalRepository) FindByID(ctx context.Context, id valueobject.DeleteApprovalID) (*delete_approval.MediaDeleteApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approval delete_approval.MediaDeleteApproval
	err = db.First(&approval, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrApprovalNotFound
		}
		return nil, err
	}
	return &approval, nil
}

// FindByMediaAndUser 根据媒体ID和用户ID查找最新的审批记录
func (r *GormMediaDeleteApprovalRepository) FindByMediaAndUser(ctx context.Context, mediaID valueobject.MediaID, userID int64) (*delete_approval.MediaDeleteApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approval delete_approval.MediaDeleteApproval
	err = db.
		Where("media_id = ? AND user_id = ?", mediaID, userID).
		Order("created_at DESC").
		First(&approval).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回nil而不是错误
		}
		return nil, err
	}
	return &approval, nil
}

// FindByInstanceID 根据工作流实例ID查找审批记录
func (r *GormMediaDeleteApprovalRepository) FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*delete_approval.MediaDeleteApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approval delete_approval.MediaDeleteApproval
	log.Printf("[GormMediaDeleteApprovalRepository] FindByInstanceID: searching for instanceId=%s", instanceID.String())
	err = db.
		Where("instance_id = ?", instanceID).
		First(&approval).Error
	if err != nil {
		log.Printf("[GormMediaDeleteApprovalRepository] FindByInstanceID: query failed: %v", err)
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrApprovalNotFound
		}
		return nil, err
	}
	log.Printf("[GormMediaDeleteApprovalRepository] FindByInstanceID: found approval: id=%s, mediaId=%d", approval.ID.String(), approval.MediaID)
	return &approval, nil
}

// FindByMediaIDs 批量查询多个媒体的审批状态
func (r *GormMediaDeleteApprovalRepository) FindByMediaIDs(ctx context.Context, mediaIDs []valueobject.MediaID, userID int64) ([]*delete_approval.MediaDeleteApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approvals []*delete_approval.MediaDeleteApproval

	// 使用子查询获取每个媒体的最新审批记录
	subQuery := db.
		Model(&delete_approval.MediaDeleteApproval{}).
		Select("MAX(id)").
		Where("media_id IN ? AND user_id = ?", mediaIDs, userID).
		Group("media_id")

	err = db.
		Where("id IN (?)", subQuery).
		Find(&approvals).Error
	if err != nil {
		return nil, err
	}

	return approvals, nil
}

// FindPendingByUser 查找用户的待审批记录
func (r *GormMediaDeleteApprovalRepository) FindPendingByUser(ctx context.Context, userID int64) ([]*delete_approval.MediaDeleteApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approvals []*delete_approval.MediaDeleteApproval
	err = db.
		Where("user_id = ? AND status = ?", userID, delete_approval.ApprovalStatusPending).
		Order("created_at DESC").
		Find(&approvals).Error
	if err != nil {
		return nil, err
	}
	return approvals, nil
}
