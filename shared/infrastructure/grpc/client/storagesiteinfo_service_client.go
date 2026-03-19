package grpc_client

import (
	"context"
	"fmt"
	"sync"

	client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	storageSiteProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/storagesite"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// ============================================
// 存储站点信息服务客户端实现 - 专注于存储站点信息业务逻辑
// ============================================

// StorageSiteInfoServiceClient 存储站点信息服务客户端实现
// 职责：实现存储站点信息服务接口，处理存储站点信息相关业务逻辑
type StorageSiteInfoServiceClient struct {
	connectionManager *ConnectionManager
	storageSiteClient storageSiteProto.StorageSiteInfoServiceClient
	initOnce          func()
}

// NewStorageSiteInfoServiceClient 创建存储站点信息服务客户端
func NewStorageSiteInfoServiceClient(connManager *ConnectionManager) client.StorageSiteInfoServiceClient {
	client := &StorageSiteInfoServiceClient{
		connectionManager: connManager,
	}

	// 懒加载客户端初始化
	var once sync.Once
	client.initOnce = func() {
		once.Do(func() {
			conn := connManager.GetConnection()
			client.storageSiteClient = storageSiteProto.NewStorageSiteInfoServiceClient(conn)
			logger.Info("存储站点服务客户端初始化完成")
		})
	}

	return client
}

// GetStorageSiteById 根据存储站点ID查询信息
func (c *StorageSiteInfoServiceClient) GetStorageSiteById(ctx context.Context, tenantId int32, id int32) (*storageSiteProto.StorageSiteInfoReply, error) {
	c.initOnce()

	req := &storageSiteProto.GetStorageSiteByIdReq{
		TenantId: tenantId,
		Id:       id,
	}

	logger.Debug(fmt.Sprintf("调用GetStorageSiteById，tenantId: %d, id: %d", tenantId, id))

	var resp *storageSiteProto.StorageSiteInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.storageSiteClient.GetStorageSiteById(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetStorageSiteById失败: %v", err))
		return nil, fmt.Errorf("查询存储站点信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetStorageSiteById返回空结果，id: %d", id))
		return nil, fmt.Errorf("存储站点信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetStorageSiteById成功，id: %d, 名称: %s",
		id, resp.StorageSiteName))
	return resp, nil
}

// GetStorageSiteByNo 根据存储站点编号查询信息
func (c *StorageSiteInfoServiceClient) GetStorageSiteByNo(ctx context.Context, tenantId int32, storageSiteNo string) (*storageSiteProto.StorageSiteInfoReply, error) {
	c.initOnce()

	req := &storageSiteProto.GetStorageSiteByNoReq{
		TenantId:      tenantId,
		StorageSiteNo: storageSiteNo,
	}

	logger.Debug(fmt.Sprintf("调用GetStorageSiteByNo，tenantId: %d, storageSiteNo: %s", tenantId, storageSiteNo))

	var resp *storageSiteProto.StorageSiteInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.storageSiteClient.GetStorageSiteByNo(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetStorageSiteByNo失败: %v", err))
		return nil, fmt.Errorf("查询存储站点信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetStorageSiteByNo返回空结果，storageSiteNo: %s", storageSiteNo))
		return nil, fmt.Errorf("存储站点信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetStorageSiteByNo成功，storageSiteNo: %s, id: %d, 名称: %s",
		storageSiteNo, resp.Id, resp.StorageSiteName))
	return resp, nil
}

// GetStorageSiteByName 根据存储站点名称查询信息
func (c *StorageSiteInfoServiceClient) GetStorageSiteByName(ctx context.Context, tenantId int32, storageSiteName string) (*storageSiteProto.StorageSiteInfoReply, error) {
	c.initOnce()

	req := &storageSiteProto.GetStorageSiteByNameReq{
		TenantId:        tenantId,
		StorageSiteName: storageSiteName,
	}

	logger.Debug(fmt.Sprintf("调用GetStorageSiteByName，tenantId: %d, storageSiteName: %s", tenantId, storageSiteName))

	var resp *storageSiteProto.StorageSiteInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.storageSiteClient.GetStorageSiteByName(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetStorageSiteByName失败: %v", err))
		return nil, fmt.Errorf("查询存储站点信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetStorageSiteByName返回空结果，storageSiteName: %s", storageSiteName))
		return nil, fmt.Errorf("存储站点信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetStorageSiteByName成功，storageSiteName: %s, id: %d",
		storageSiteName, resp.Id))
	return resp, nil
}

// GetStorageSiteByIp 根据存储站点IP查询信息
func (c *StorageSiteInfoServiceClient) GetStorageSiteByIp(ctx context.Context, tenantId int32, storageSiteIp string) (*storageSiteProto.StorageSiteInfoReply, error) {
	c.initOnce()

	req := &storageSiteProto.GetStorageSiteByIpReq{
		TenantId:      tenantId,
		StorageSiteIp: storageSiteIp,
	}

	logger.Debug(fmt.Sprintf("调用GetStorageSiteByIp，tenantId: %d, storageSiteIp: %s", tenantId, storageSiteIp))

	var resp *storageSiteProto.StorageSiteInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.storageSiteClient.GetStorageSiteByIp(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetStorageSiteByIp失败: %v", err))
		return nil, fmt.Errorf("查询存储站点信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetStorageSiteByIp返回空结果，storageSiteIp: %s", storageSiteIp))
		return nil, fmt.Errorf("存储站点信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetStorageSiteByIp成功，storageSiteIp: %s, id: %d, 名称: %s",
		storageSiteIp, resp.Id, resp.StorageSiteName))
	return resp, nil
}
