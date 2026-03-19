package client

import (
	"context"
	proto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/storagesite"
)

// StorageSiteInfoServiceClient 存储站点信息服务客户端接口
type StorageSiteInfoServiceClient interface {
	// GetStorageSiteById 根据存储站点ID查询信息
	GetStorageSiteById(ctx context.Context, tenantId int32, id int32) (*proto.StorageSiteInfoReply, error)

	// GetStorageSiteByNo 根据存储站点编号查询信息
	GetStorageSiteByNo(ctx context.Context, tenantId int32, storageSiteNo string) (*proto.StorageSiteInfoReply, error)

	// GetStorageSiteByName 根据存储站点名称查询信息
	GetStorageSiteByName(ctx context.Context, tenantId int32, storageSiteName string) (*proto.StorageSiteInfoReply, error)

	// GetStorageSiteByIp 根据存储站点IP查询信息
	GetStorageSiteByIp(ctx context.Context, tenantId int32, storageSiteIp string) (*proto.StorageSiteInfoReply, error)
}
