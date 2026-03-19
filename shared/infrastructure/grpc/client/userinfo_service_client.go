package grpc_client

import (
	"context"
	"fmt"
	"sync"

	client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	userProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/user"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// ============================================
// 用户信息服务客户端实现 - 专注于用户信息业务逻辑
// ============================================

// UserInfoServiceClient 用户信息服务客户端实现
// 职责：实现用户信息服务接口，处理用户信息相关业务逻辑
type UserInfoServiceClient struct {
	connectionManager *ConnectionManager
	userClient        userProto.UserInfoServiceClient
	initOnce          func()
}

// NewUserInfoClient 创建用户信息服务客户端
func NewUserInfoServiceClient(connManager *ConnectionManager) client.UserInfoServiceClient {
	client := &UserInfoServiceClient{
		connectionManager: connManager,
	}

	// 懒加载客户端初始化
	var once sync.Once
	client.initOnce = func() {
		once.Do(func() {
			conn := connManager.GetConnection()
			client.userClient = userProto.NewUserInfoServiceClient(conn)
			logger.Info("用户服务客户端初始化完成")
		})
	}

	return client
}

// GetUserById 根据用户ID查询用户信息
func (c *UserInfoServiceClient) GetUserById(ctx context.Context, tenantId int32, userId int32) (*userProto.UserInfoReply, error) {
	c.initOnce()

	req := &userProto.GetUserByIdReq{
		TenantId: tenantId,
		UserId:   userId,
	}

	logger.Debug(fmt.Sprintf("调用GetUserById，tenantId: %d, userId: %d", tenantId, userId))

	var resp *userProto.UserInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.userClient.GetUserById(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetUserById失败: %v", err))
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	logger.Debug(fmt.Sprintf("GetUserById成功，userId: %d, 用户名: %s", resp.UserId, resp.UserName))
	return resp, nil
}

// GetUserByPoliceNo 根据警号查询用户信息
func (c *UserInfoServiceClient) GetUserByPoliceNo(ctx context.Context, tenantId int32, policeNo string) (*userProto.UserInfoReply, error) {
	c.initOnce()

	req := &userProto.GetUserByPoliceNoReq{
		TenantId: tenantId,
		PoliceNo: policeNo,
	}

	logger.Debug(fmt.Sprintf("调用GetUserByPoliceNo，tenantId: %d, policeNo: %s", tenantId, policeNo))

	var resp *userProto.UserInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.userClient.GetUserByPoliceNo(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetUserByPoliceNo失败: %v", err))
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetUserByPoliceNo返回空结果，policeNo: %s", policeNo))
		return nil, fmt.Errorf("用户信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetUserByPoliceNo成功，policeNo: %s, userId: %d, 用户名: %s",
		policeNo, resp.UserId, resp.UserName))
	return resp, nil
}
