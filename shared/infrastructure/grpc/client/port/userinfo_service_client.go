package client

import (
	"context"
	proto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/user"
)

// UserInfoServiceClient 用户信息服务客户端接口
type UserInfoServiceClient interface {
	// GetUserById 根据用户ID查询用户信息
	GetUserById(ctx context.Context, tenantId int32, userId int32) (*proto.UserInfoReply, error)

	// GetUserByPoliceNo 根据警号查询用户信息
	GetUserByPoliceNo(ctx context.Context, tenantId int32, policeNo string) (*proto.UserInfoReply, error)
}
