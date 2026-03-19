package port

import (
	"context"

	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
)

// StorageSiteInfoService 存储站点信息基础设施服务接口（Port）
// 职责：定义存储站点信息服务的契约，供应用层依赖
type StorageSiteInfoService interface {
	// GetStorageSiteById 根据存储站点ID查询信息
	GetStorageSiteById(ctx context.Context, storageSiteID int32) (*dto.StorageSiteInfo, error)

	// GetStorageSiteByNo 根据存储站点编号查询信息
	GetStorageSiteByNo(ctx context.Context, storageSiteNo string) (*dto.StorageSiteInfo, error)

	// GetStorageSiteByName 根据存储站点名称查询信息
	GetStorageSiteByName(ctx context.Context, storageSiteName string) (*dto.StorageSiteInfo, error)

	// GetStorageSiteByIp 根据存储站点IP查询信息
	GetStorageSiteByIp(ctx context.Context, storageSiteIp string) (*dto.StorageSiteInfo, error)

	// ValidateStorageSiteExists 验证存储站点是否存在
	ValidateStorageSiteExists(ctx context.Context, storageSiteID int32) error
}
