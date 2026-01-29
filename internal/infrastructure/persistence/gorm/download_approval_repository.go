package persistence

import (
	"context"
	"log"
	"time"

	download_approval "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/errors"

	"gorm.io/gorm"
)

// GormMediaDownloadApprovalRepository GORM实现的媒体下载审批仓储
type GormMediaDownloadApprovalRepository struct {
	GormRepository
}

// Save 保存审批记录
func (r *GormMediaDownloadApprovalRepository) Save(ctx context.Context, approval *download_approval.MediaDownloadApproval) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	return db.Create(approval).Error
}

// Update 更新审批记录
func (r *GormMediaDownloadApprovalRepository) Update(ctx context.Context, approval *download_approval.MediaDownloadApproval) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	return db.Save(approval).Error
}

// FindByID 根据ID查找审批记录
func (r *GormMediaDownloadApprovalRepository) FindByID(ctx context.Context, id valueobject.DownloadApprovalID) (*download_approval.MediaDownloadApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approval download_approval.MediaDownloadApproval
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
func (r *GormMediaDownloadApprovalRepository) FindByMediaAndUser(ctx context.Context, mediaID valueobject.MediaID, userID int64) (*download_approval.MediaDownloadApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approval download_approval.MediaDownloadApproval
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
func (r *GormMediaDownloadApprovalRepository) FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*download_approval.MediaDownloadApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approval download_approval.MediaDownloadApproval
	log.Printf("[GormMediaDownloadApprovalRepository] FindByInstanceID: searching for instanceId=%s", instanceID.String())
	err = db.
		Where("instance_id = ?", instanceID).
		First(&approval).Error
	if err != nil {
		log.Printf("[GormMediaDownloadApprovalRepository] FindByInstanceID: query failed: %v", err)
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrApprovalNotFound
		}
		return nil, err
	}
	log.Printf("[GormMediaDownloadApprovalRepository] FindByInstanceID: found approval: id=%s, mediaId=%d", approval.ID.String(), approval.MediaID)
	return &approval, nil
}

// FindByMediaIDs 批量查询多个媒体的审批状态
func (r *GormMediaDownloadApprovalRepository) FindByMediaIDs(ctx context.Context, mediaIDs []valueobject.MediaID, userID int64) ([]*download_approval.MediaDownloadApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approvals []*download_approval.MediaDownloadApproval

	// 使用子查询获取每个媒体的最新审批记录
	subQuery := db.
		Model(&download_approval.MediaDownloadApproval{}).
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
func (r *GormMediaDownloadApprovalRepository) FindPendingByUser(ctx context.Context, userID int64) ([]*download_approval.MediaDownloadApproval, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	var approvals []*download_approval.MediaDownloadApproval
	err = db.
		Where("user_id = ? AND status = ?", userID, download_approval.ApprovalStatusPending).
		Order("created_at DESC").
		Find(&approvals).Error
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

// UpdateExpiredRecords 更新过期的审批记录
func (r *GormMediaDownloadApprovalRepository) UpdateExpiredRecords(ctx context.Context) (int64, error) {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return 0, err
	}
	result := db.
		Model(&download_approval.MediaDownloadApproval{}).
		Where("status = ? AND expired_at < ?", download_approval.ApprovalStatusApproved, time.Now()).
		Update("status", download_approval.ApprovalStatusExpired)
	return result.RowsAffected, result.Error
}
