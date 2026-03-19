package client

import (
	"context"
	proto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/org"
)

// OrgInfoServiceClient 组织信息服务客户端接口
type OrgInfoServiceClient interface {
	// GetOrgById 根据组织ID查询组织信息
	GetOrgById(ctx context.Context, tenantId int32, orgId int32) (*proto.OrgInfoReply, error)

	// GetOrgByCode 根据组织编码查询组织信息
	GetOrgByCode(ctx context.Context, tenantId int32, orgCode string) (*proto.OrgInfoReply, error)

	// GetOrgByName 根据组织名称查询组织信息
	GetOrgByName(ctx context.Context, tenantId int32, orgName string) (*proto.OrgInfoReply, error)

	// GetOrgFullName 获取组织全名（包含上级路径）
	GetOrgFullName(ctx context.Context, tenantId int32, orgId int32) (*proto.OrgFullNameReply, error)
}
