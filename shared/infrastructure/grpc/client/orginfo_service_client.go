package grpc_client

import (
	"context"
	"fmt"
	"sync"

	client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	orgProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/org"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// ============================================
// 组织信息服务客户端实现 - 专注于组织信息业务逻辑
// ============================================

// OrgInfoServiceClient 组织信息服务客户端实现
// 职责：实现组织信息服务接口，处理组织信息相关业务逻辑
type OrgInfoServiceClient struct {
	connectionManager *ConnectionManager
	orgClient         orgProto.OrgInfoServiceClient
	initOnce          func()
}

// NewOrgInfoServiceClient 创建组织信息服务客户端
func NewOrgInfoServiceClient(connManager *ConnectionManager) client.OrgInfoServiceClient {
	client := &OrgInfoServiceClient{
		connectionManager: connManager,
	}

	// 懒加载客户端初始化
	var once sync.Once
	client.initOnce = func() {
		once.Do(func() {
			conn := connManager.GetConnection()
			client.orgClient = orgProto.NewOrgInfoServiceClient(conn)
			logger.Info("组织服务客户端初始化完成")
		})
	}

	return client
}

// GetOrgById 根据组织ID查询组织信息
func (c *OrgInfoServiceClient) GetOrgById(ctx context.Context, tenantId int32, orgId int32) (*orgProto.OrgInfoReply, error) {
	c.initOnce()

	req := &orgProto.GetOrgByIdReq{
		TenantId: tenantId,
		OrgId:    orgId,
	}

	logger.Debug(fmt.Sprintf("调用GetOrgById，tenantId: %d, orgId: %d", tenantId, orgId))

	var resp *orgProto.OrgInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.orgClient.GetOrgById(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetOrgById失败: %v", err))
		return nil, fmt.Errorf("查询组织信息失败: %w", err)
	}

	logger.Debug(fmt.Sprintf("GetOrgById成功，orgId: %d, 组织名: %s", resp.OrgId, resp.OrgName))
	return resp, nil
}

// GetOrgByCode 根据组织编码查询组织信息
func (c *OrgInfoServiceClient) GetOrgByCode(ctx context.Context, tenantId int32, orgCode string) (*orgProto.OrgInfoReply, error) {
	c.initOnce()

	req := &orgProto.GetOrgByCodeReq{
		TenantId: tenantId,
		OrgCode:  orgCode,
	}

	logger.Debug(fmt.Sprintf("调用GetOrgByCode，tenantId: %d, orgCode: %s", tenantId, orgCode))

	var resp *orgProto.OrgInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.orgClient.GetOrgByCode(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetOrgByCode失败: %v", err))
		return nil, fmt.Errorf("查询组织信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetOrgByCode返回空结果，orgCode: %s", orgCode))
		return nil, fmt.Errorf("组织信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetOrgByCode成功，orgCode: %s, orgId: %d, 组织名: %s",
		orgCode, resp.OrgId, resp.OrgName))
	return resp, nil
}

// GetOrgByName 根据组织名称查询组织信息
func (c *OrgInfoServiceClient) GetOrgByName(ctx context.Context, tenantId int32, orgName string) (*orgProto.OrgInfoReply, error) {
	c.initOnce()

	req := &orgProto.GetOrgByNameReq{
		TenantId: tenantId,
		OrgName:  orgName,
	}

	logger.Debug(fmt.Sprintf("调用GetOrgByName，tenantId: %d, orgName: %s", tenantId, orgName))

	var resp *orgProto.OrgInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.orgClient.GetOrgByName(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetOrgByName失败: %v", err))
		return nil, fmt.Errorf("查询组织信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetOrgByName返回空结果，orgName: %s", orgName))
		return nil, fmt.Errorf("组织信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetOrgByName成功，orgName: %s, orgId: %d",
		orgName, resp.OrgId))
	return resp, nil
}

// GetOrgFullName 获取组织全名（包含上级路径）
func (c *OrgInfoServiceClient) GetOrgFullName(ctx context.Context, tenantId int32, orgId int32) (*orgProto.OrgFullNameReply, error) {
	c.initOnce()

	req := &orgProto.GetOrgFullNameReq{
		TenantId: tenantId,
		OrgId:    orgId,
	}

	logger.Debug(fmt.Sprintf("调用GetOrgFullName，tenantId: %d, orgId: %d", tenantId, orgId))

	var resp *orgProto.OrgFullNameReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.orgClient.GetOrgFullName(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetOrgFullName失败: %v", err))
		return nil, fmt.Errorf("查询组织全名失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetOrgFullName返回空结果，orgId: %d", orgId))
		return nil, fmt.Errorf("组织全名查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetOrgFullName成功，orgId: %d, 组织全名: %s",
		orgId, resp.FullName))
	return resp, nil
}
