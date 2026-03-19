package client

import (
	"context"
	proto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/bwc"
)

// BWCInfoServiceClient 执法记录仪信息服务客户端接口
type BWCInfoServiceClient interface {
	// GetBWCById 根据执法记录仪ID查询信息
	GetBWCById(ctx context.Context, tenantId int32, id int32) (*proto.BWCInfoReply, error)

	// GetBWCByNo 根据执法记录仪编号查询信息
	GetBWCByNo(ctx context.Context, tenantId int32, no string) (*proto.BWCInfoReply, error)

	// GetBWCsByManagerId 根据管理员ID查询执法记录仪列表
	GetBWCsByManagerId(ctx context.Context, tenantId int32, managerId int32) (*proto.BWCListReply, error)

	// GetBWCsByRequisitionerId 根据领用人ID查询执法记录仪列表
	GetBWCsByRequisitionerId(ctx context.Context, tenantId int32, requisitionerId int32) (*proto.BWCListReply, error)
}
