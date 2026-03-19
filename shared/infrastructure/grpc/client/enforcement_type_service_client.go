package grpc_client

import (
	"context"
	"fmt"
	"sync"

	client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	enforcementTypeProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/enforcement_type"
	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// ============================================
// 执法类型信息服务客户端实现 - 专注于执法类型信息业务逻辑
// ============================================

// EnforcementTypeServiceClient 执法类型信息服务客户端实现
// 职责：实现执法类型信息服务接口，处理执法类型信息相关业务逻辑
type EnforcementTypeServiceClient struct {
	connectionManager     *ConnectionManager
	enforcementTypeClient enforcementTypeProto.EnforcementTypeInfoServiceClient
	initOnce              func()
}

// NewEnforcementTypeServiceClient 创建执法类型信息服务客户端
func NewEnforcementTypeServiceClient(connManager *ConnectionManager) client.EnforcementTypeServiceClient {
	client := &EnforcementTypeServiceClient{
		connectionManager: connManager,
	}

	// 懒加载客户端初始化
	var once sync.Once
	client.initOnce = func() {
		once.Do(func() {
			conn := connManager.GetConnection()
			client.enforcementTypeClient = enforcementTypeProto.NewEnforcementTypeInfoServiceClient(conn)
			logger.Info("执法类型服务客户端初始化完成")
		})
	}

	return client
}

// GetEnforcementTypeById 根据执法类型ID查询执法类型信息
func (c *EnforcementTypeServiceClient) GetEnforcementTypeById(ctx context.Context, tenantId int32, id int64) (*dto.EnforcementTypeInfo, error) {
	c.initOnce()

	req := &enforcementTypeProto.GetEnforcementTypeByIdReq{
		TenantId: tenantId,
		Id:       id,
	}

	logger.Debug(fmt.Sprintf("调用GetEnforcementTypeById，tenantId: %d, id: %d", tenantId, id))

	var resp *enforcementTypeProto.EnforcementTypeInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.enforcementTypeClient.GetEnforcementTypeById(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetEnforcementTypeById失败: %v", err))
		return nil, fmt.Errorf("查询执法类型信息失败: %w", err)
	}

	logger.Debug(fmt.Sprintf("GetEnforcementTypeById成功，id: %d, 执法类型名: %s", resp.Id, resp.EnforcementTypeName))
	return convertProtoToDTO(resp), nil
}

// GetEnforcementTypeByCode 根据执法类型编码查询执法类型信息
func (c *EnforcementTypeServiceClient) GetEnforcementTypeByCode(ctx context.Context, tenantId int32, code string) (*dto.EnforcementTypeInfo, error) {
	c.initOnce()

	req := &enforcementTypeProto.GetEnforcementTypeByCodeReq{
		TenantId:            tenantId,
		EnforcementTypeCode: code,
	}

	logger.Debug(fmt.Sprintf("调用GetEnforcementTypeByCode，tenantId: %d, code: %s", tenantId, code))

	var resp *enforcementTypeProto.EnforcementTypeInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.enforcementTypeClient.GetEnforcementTypeByCode(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetEnforcementTypeByCode失败: %v", err))
		return nil, fmt.Errorf("查询执法类型信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetEnforcementTypeByCode返回空结果，code: %s", code))
		return nil, fmt.Errorf("执法类型信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetEnforcementTypeByCode成功，code: %s, id: %d, 执法类型名: %s",
		code, resp.Id, resp.EnforcementTypeName))
	return convertProtoToDTO(resp), nil
}

// GetEnforcementTypeByName 根据执法类型名称查询执法类型信息
func (c *EnforcementTypeServiceClient) GetEnforcementTypeByName(ctx context.Context, tenantId int32, name string) (*dto.EnforcementTypeInfo, error) {
	c.initOnce()

	req := &enforcementTypeProto.GetEnforcementTypeByNameReq{
		TenantId:            tenantId,
		EnforcementTypeName: name,
	}

	logger.Debug(fmt.Sprintf("调用GetEnforcementTypeByName，tenantId: %d, name: %s", tenantId, name))

	var resp *enforcementTypeProto.EnforcementTypeInfoReply
	err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = c.enforcementTypeClient.GetEnforcementTypeByName(ctx, req)
		return err
	})

	if err != nil {
		logger.Error(fmt.Sprintf("GetEnforcementTypeByName失败: %v", err))
		return nil, fmt.Errorf("查询执法类型信息失败: %w", err)
	}

	if resp == nil {
		logger.Warn(fmt.Sprintf("GetEnforcementTypeByName返回空结果，name: %s", name))
		return nil, fmt.Errorf("执法类型信息查询结果为空")
	}

	logger.Debug(fmt.Sprintf("GetEnforcementTypeByName成功，name: %s, id: %d, 执法类型名: %s",
		name, resp.Id, resp.EnforcementTypeName))
	return convertProtoToDTO(resp), nil
}

// convertProtoToDTO 将 proto 响应转换为 DTO
func convertProtoToDTO(proto *enforcementTypeProto.EnforcementTypeInfoReply) *dto.EnforcementTypeInfo {
	if proto == nil {
		return nil
	}
	return &dto.EnforcementTypeInfo{
		ID:                  proto.Id,
		EnforcementTypeCode: proto.EnforcementTypeCode,
		EnforcementTypeName: proto.EnforcementTypeName,
		EnforcementTypeDesc: proto.EnforcementTypeDesc,
		EnforcementTypePath: proto.EnforcementTypePath,
		ParentId:            proto.ParentId,
		Source:              proto.Source,
		Sort:                proto.Sort,
	}
}
