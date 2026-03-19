package grpc_client

import (
	"context"
	"fmt"
	"sync"

	client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	bwcProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/bwc"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// ============================================
// 执法记录仪信息服务客户端实现 - 专注于执法记录仪信息业务逻辑
// ============================================

// BWCInfoServiceClient 执法记录仪信息服务客户端实现
// 职责：实现执法记录仪信息服务接口，处理执法记录仪信息相关业务逻辑
type BWCInfoServiceClient struct {
	connectionManager *ConnectionManager
	bwcClient         bwcProto.BWCInfoServiceClient
	initOnce          func()
}

// NewBWCInfoServiceClient 创建执法记录仪信息服务客户端
func NewBWCInfoServiceClient(connManager *ConnectionManager) client.BWCInfoServiceClient {
	client := &BWCInfoServiceClient{
		connectionManager: connManager,
	}

	// 懒加载客户端初始化
	var once sync.Once
	client.initOnce = func() {
		once.Do(func() {
			conn := connManager.GetConnection()
			client.bwcClient = bwcProto.NewBWCInfoServiceClient(conn)
			logger.Info("执法记录仪服务客户端初始化完成")
		})
	}

	return client
}

// GetBWCById 根据执法记录仪ID查询信息
func (c *BWCInfoServiceClient) GetBWCById(ctx context.Context, tenantId int32, id int32) (*bwcProto.BWCInfoReply, error) {
	c.initOnce()

	req := &bwcProto.GetBWCByIdReq{
		TenantId: tenantId,
		Id:       id,
	}

	logger.Debug(fmt.Sprintf("调用GetBWCById，tenantId: %d, id: %d", tenantId, id))

	var resp *bwcProto.BWCInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.bwcClient.GetBWCById(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetBWCById失败: %v", err))
		return nil, fmt.Errorf("查询执法记录仪信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetBWCById返回空结果，id: %d", id))
		return nil, fmt.Errorf("执法记录仪信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetBWCById成功，id: %d, 名称: %s",
		id, resp.BwcName))
	return resp, nil
}

// GetBWCByNo 根据执法记录仪编号查询信息
func (c *BWCInfoServiceClient) GetBWCByNo(ctx context.Context, tenantId int32, bwcNo string) (*bwcProto.BWCInfoReply, error) {
	c.initOnce()

	req := &bwcProto.GetBWCByNoReq{
		TenantId: tenantId,
		BwcNo:    bwcNo,
	}

	logger.Debug(fmt.Sprintf("调用GetBWCByNo，tenantId: %d, no: %s", tenantId, bwcNo))

	var resp *bwcProto.BWCInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.bwcClient.GetBWCByNo(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetBWCByNo失败: %v", err))
		return nil, fmt.Errorf("查询执法记录仪信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetBWCByNo返回空结果，no: %s", bwcNo))
		return nil, fmt.Errorf("执法记录仪信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetBWCByNo成功，no: %s, id: %d, 名称: %s",
		bwcNo, resp.Id, resp.BwcName))
	return resp, nil
}

// GetBWCsByManagerId 根据管理员ID查询执法记录仪列表
func (c *BWCInfoServiceClient) GetBWCsByManagerId(ctx context.Context, tenantId int32, managerId int32) (*bwcProto.BWCListReply, error) {
	c.initOnce()

	// 检查连接健康状态
	if !c.connectionManager.IsHealthy() {
		if c.connectionManager.IsReconnecting() {
			return nil, fmt.Errorf("gRPC连接正在重连中(第%d次尝试)，请稍后重试",
				c.connectionManager.GetReconnectAttempts())
		}
		return nil, fmt.Errorf("gRPC连接不健康，连接状态: %v",
			c.connectionManager.GetConnectionState())
	}

	req := &bwcProto.GetBWCsByManagerIdReq{
		TenantId:  tenantId,
		ManagerId: managerId,
	}

	logger.Debug(fmt.Sprintf("调用GetBWCsByManagerId，tenantId: %d, managerId: %d", tenantId, managerId))

	resp, err := c.bwcClient.GetBWCsByManagerId(ctx, req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetBWCsByManagerId失败: %v", err))
		// 检查是否是连接相关错误，触发健康检查
		c.connectionManager.CheckHealth()
		return nil, fmt.Errorf("查询管理员执法记录仪列表失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetBWCsByManagerId返回空结果，managerId: %d", managerId))
		return nil, fmt.Errorf("管理员执法记录仪列表查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetBWCsByManagerId成功，managerId: %d, 数量: %d",
		managerId, resp.Total))
	return resp, nil
}

// GetBWCsByRequisitionerId 根据领用人ID查询执法记录仪列表
func (c *BWCInfoServiceClient) GetBWCsByRequisitionerId(ctx context.Context, tenantId int32, requisitionerId int32) (*bwcProto.BWCListReply, error) {
	c.initOnce()

	// 检查连接健康状态
	if !c.connectionManager.IsHealthy() {
		if c.connectionManager.IsReconnecting() {
			return nil, fmt.Errorf("gRPC连接正在重连中(第%d次尝试)，请稍后重试",
				c.connectionManager.GetReconnectAttempts())
		}
		return nil, fmt.Errorf("gRPC连接不健康，连接状态: %v",
			c.connectionManager.GetConnectionState())
	}

	req := &bwcProto.GetBWCsByRequisitionerIdReq{
		TenantId:        tenantId,
		RequisitionerId: requisitionerId,
	}

	logger.Debug(fmt.Sprintf("调用GetBWCsByRequisitionerId，tenantId: %d, requisitionerId: %d", tenantId, requisitionerId))

	resp, err := c.bwcClient.GetBWCsByRequisitionerId(ctx, req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetBWCsByRequisitionerId失败: %v", err))
		// 检查是否是连接相关错误，触发健康检查
		c.connectionManager.CheckHealth()
		return nil, fmt.Errorf("查询领用人执法记录仪列表失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetBWCsByRequisitionerId返回空结果，requisitionerId: %d", requisitionerId))
		return nil, fmt.Errorf("领用人执法记录仪列表查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetBWCsByRequisitionerId成功，requisitionerId: %d, 数量: %d",
		requisitionerId, resp.Total))
	return resp, nil
}
