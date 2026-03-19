package port

import (
	"context"

	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
)

// BWCInfoService 执法记录仪信息基础设施服务接口（Port）
// 职责：定义执法记录仪信息服务的契约，供应用层依赖
type BWCInfoService interface {
	// GetBWCById 根据执法记录仪ID查询信息
	GetBWCById(ctx context.Context, bwcID int32) (*dto.BWCInfo, error)

	// GetBWCByNo 根据执法记录仪编号查询信息
	GetBWCByNo(ctx context.Context, bwcNo string) (*dto.BWCInfo, error)

	// ValidateBWCExists 验证执法记录仪是否存在
	ValidateBWCExists(ctx context.Context, bwcID int32) error
}
