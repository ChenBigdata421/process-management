package port

import (
	"context"

	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
)

// UserInfoService 用户信息基础设施服务接口（Port）
// 职责：定义用户信息服务的契约，供应用层依赖
type UserInfoService interface {
	// GetUserById 根据用户ID查询用户信息
	GetUserById(ctx context.Context, userID int32) (*dto.UserInfo, error)

	// GetUserByPoliceNo 根据警号查询用户信息
	GetUserByPoliceNo(ctx context.Context, policeNo string) (*dto.UserInfo, error)

	// ValidateUserExists 验证用户是否存在
	ValidateUserExists(ctx context.Context, userID int32) error
}
