package client

import (
	"context"
	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
)

// EnforcementTypeServiceClient 执法类型信息服务客户端接口
type EnforcementTypeServiceClient interface {
	// GetEnforcementTypeById 根据执法类型ID查询执法类型信息
	GetEnforcementTypeById(ctx context.Context, tenantId int32, id int64) (*dto.EnforcementTypeInfo, error)

	// GetEnforcementTypeByCode 根据执法类型编码查询执法类型信息
	GetEnforcementTypeByCode(ctx context.Context, tenantId int32, code string) (*dto.EnforcementTypeInfo, error)

	// GetEnforcementTypeByName 根据执法类型名称查询执法类型信息
	GetEnforcementTypeByName(ctx context.Context, tenantId int32, name string) (*dto.EnforcementTypeInfo, error)
}
