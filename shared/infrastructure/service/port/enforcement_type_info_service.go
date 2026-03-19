package port

import (
	"context"

	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
)

// EnforcementTypeInfoService 执法类型信息基础设施服务接口（Port）
// 职责：定义执法类型信息服务的契约，供应用层依赖
type EnforcementTypeInfoService interface {
	// ValidateEnforcementTypeExists 验证执法类型是否存在
	ValidateEnforcementTypeExists(ctx context.Context, enforcementTypeID int64) error

	// GetEnforcementTypeCode 获取执法类型编码
	GetEnforcementTypeCode(ctx context.Context, enforcementTypeID int64) (string, error)

	// GetEnforcementTypeName 获取执法类型名称
	GetEnforcementTypeName(ctx context.Context, enforcementTypeID int64) (string, error)

	// GetEnforcementTypeInfo 获取完整执法类型信息
	GetEnforcementTypeInfo(ctx context.Context, enforcementTypeID int64) (*dto.EnforcementTypeInfo, error)

	// GetEnforcementTypeByCode 根据编码获取执法类型信息
	GetEnforcementTypeByCode(ctx context.Context, code string) (*dto.EnforcementTypeInfo, error)
}
