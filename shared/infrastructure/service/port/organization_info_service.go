package port

import (
	"context"

	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
)

// OrganizationInfoService 组织信息基础设施服务接口（Port）
// 职责：定义组织信息服务的契约，供应用层依赖
type OrganizationInfoService interface {
	// ValidateOrgExists 验证组织是否存在
	ValidateOrgExists(ctx context.Context, orgID int) error

	// GetOrgCode 获取组织编码
	GetOrgCode(ctx context.Context, orgID int) (string, error)

	// GetOrgFullName 获取组织全名
	GetOrgFullName(ctx context.Context, orgID int) (string, error)

	// GetOrgInfo 获取完整组织信息
	GetOrgInfo(ctx context.Context, orgID int) (*dto.OrganizationInfo, error)
}
