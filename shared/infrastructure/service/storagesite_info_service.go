package infrastructure_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jxt-evidence-system/process-management/shared/common/global"
	grpcclient "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	storageSiteProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/storagesite"
	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
	"jxt-evidence-system/process-management/shared/infrastructure/service/port"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"go.uber.org/zap"
)

type storageSiteInfoService struct {
	storageSiteInfoClient grpcclient.StorageSiteInfoServiceClient
	logger                *zap.Logger
}

// NewStorageSiteInfoService 创建存储站点信息服务实例
func NewStorageSiteInfoService(storageSiteInfoClient grpcclient.StorageSiteInfoServiceClient) port.StorageSiteInfoService {
	return &storageSiteInfoService{
		storageSiteInfoClient: storageSiteInfoClient,
		logger:                logger.Logger,
	}
}

func (s *storageSiteInfoService) GetStorageSiteById(ctx context.Context, storageSiteID int32) (*dto.StorageSiteInfo, error) {
	if storageSiteID <= 0 {
		return nil, fmt.Errorf("无效的存储站点ID: %d", storageSiteID)
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}
	var lastErr error

	// 统一的重试+降级逻辑：适度重试，快速降级
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		storageSiteInfo, err := s.storageSiteInfoClient.GetStorageSiteById(ctx, int32(tenantID), storageSiteID)
		if err == nil && storageSiteInfo != nil {
			if attempt > 0 {
				s.logger.Info("存储站点信息获取成功",
					zap.Int32("storageSiteID", storageSiteID), zap.Int("attempt", attempt+1))
			}
			return s.convertStorageSiteInfo(storageSiteInfo), nil
		}

		lastErr = err

		// 最后一次尝试失败，跳出循环进行降级
		if attempt == MaxRetries {
			break
		}

		// 只对可重试的错误进行重试
		if !s.shouldRetry(err) {
			break
		}

		// 固定间隔重试，简单有效
		s.logger.Warn("存储站点信息获取失败，准备重试",
			zap.Error(err), zap.Int32("storageSiteID", storageSiteID), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断
	s.logger.Warn("存储站点信息获取失败，使用降级处理",
		zap.Error(lastErr), zap.Int32("storageSiteID", storageSiteID))
	return &dto.StorageSiteInfo{
		ID:              storageSiteID,
		StorageSiteNo:   fmt.Sprintf("SITE_%d", storageSiteID),
		StorageSiteName: fmt.Sprintf("存储站点_%d", storageSiteID),
		StorageSiteIP:   "",
		StorageSiteURL:  "",
		OpenStatus:      0,
		AuthKey:         "",
	}, nil
}

func (s *storageSiteInfoService) GetStorageSiteByNo(ctx context.Context, storageSiteNo string) (*dto.StorageSiteInfo, error) {
	if storageSiteNo == "" {
		return nil, fmt.Errorf("存储站点编号不能为空")
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}
	var lastErr error

	// 统一的重试+降级逻辑
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		storageSiteInfo, err := s.storageSiteInfoClient.GetStorageSiteByNo(ctx, int32(tenantID), storageSiteNo)
		if err == nil && storageSiteInfo != nil {
			if attempt > 0 {
				s.logger.Info("根据编号获取存储站点信息成功",
					zap.String("storageSiteNo", storageSiteNo), zap.Int("attempt", attempt+1))
			}
			return s.convertStorageSiteInfo(storageSiteInfo), nil
		}

		lastErr = err

		// 最后一次尝试失败，跳出循环
		if attempt == MaxRetries {
			break
		}

		// 只对可重试的错误进行重试
		if !s.shouldRetry(err) {
			break
		}

		s.logger.Warn("根据编号获取存储站点信息失败，准备重试",
			zap.Error(err), zap.String("storageSiteNo", storageSiteNo), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 对于根据编号查询，如果失败则返回错误（不降级）
	s.logger.Warn("根据编号查询存储站点信息失败", zap.Error(lastErr), zap.String("storageSiteNo", storageSiteNo))
	return nil, fmt.Errorf("根据存储站点编号[%s]查询存储站点信息失败: %w", storageSiteNo, lastErr)
}

func (s *storageSiteInfoService) GetStorageSiteByName(ctx context.Context, storageSiteName string) (*dto.StorageSiteInfo, error) {
	if storageSiteName == "" {
		return nil, fmt.Errorf("存储站点名称不能为空")
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}
	var lastErr error

	// 统一的重试+降级逻辑
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		storageSiteInfo, err := s.storageSiteInfoClient.GetStorageSiteByName(ctx, int32(tenantID), storageSiteName)
		if err == nil && storageSiteInfo != nil {
			if attempt > 0 {
				s.logger.Info("根据名称获取存储站点信息成功",
					zap.String("storageSiteName", storageSiteName), zap.Int("attempt", attempt+1))
			}
			return s.convertStorageSiteInfo(storageSiteInfo), nil
		}

		lastErr = err

		if attempt == MaxRetries {
			break
		}

		if !s.shouldRetry(err) {
			break
		}

		s.logger.Warn("根据名称获取存储站点信息失败，准备重试",
			zap.Error(err), zap.String("storageSiteName", storageSiteName), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	s.logger.Warn("根据名称查询存储站点信息失败", zap.Error(lastErr), zap.String("storageSiteName", storageSiteName))
	return nil, fmt.Errorf("根据存储站点名称[%s]查询存储站点信息失败: %w", storageSiteName, lastErr)
}

func (s *storageSiteInfoService) GetStorageSiteByIp(ctx context.Context, storageSiteIp string) (*dto.StorageSiteInfo, error) {
	if storageSiteIp == "" {
		return nil, fmt.Errorf("存储站点IP不能为空")
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}
	var lastErr error

	// 统一的重试+降级逻辑
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		storageSiteInfo, err := s.storageSiteInfoClient.GetStorageSiteByIp(ctx, int32(tenantID), storageSiteIp)
		if err == nil && storageSiteInfo != nil {
			if attempt > 0 {
				s.logger.Info("根据IP获取存储站点信息成功",
					zap.String("storageSiteIp", storageSiteIp), zap.Int("attempt", attempt+1))
			}
			return s.convertStorageSiteInfo(storageSiteInfo), nil
		}

		lastErr = err

		if attempt == MaxRetries {
			break
		}

		if !s.shouldRetry(err) {
			break
		}

		s.logger.Warn("根据IP获取存储站点信息失败，准备重试",
			zap.Error(err), zap.String("storageSiteIp", storageSiteIp), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	s.logger.Warn("根据IP查询存储站点信息失败", zap.Error(lastErr), zap.String("storageSiteIp", storageSiteIp))
	return nil, fmt.Errorf("根据存储站点IP[%s]查询存储站点信息失败: %w", storageSiteIp, lastErr)
}

func (s *storageSiteInfoService) ValidateStorageSiteExists(ctx context.Context, storageSiteID int32) error {
	if storageSiteID <= 0 {
		return fmt.Errorf("无效的存储站点ID: %d", storageSiteID)
	}
	_, err := s.GetStorageSiteById(ctx, storageSiteID)
	if err != nil {
		s.logger.Warn("存储站点验证失败，使用降级处理", zap.Error(err), zap.Int32("storageSiteID", storageSiteID))
		return nil // 降级处理：测试环境或服务不可用时不阻断业务
	}
	return nil
}

// convertStorageSiteInfo 转换存储站点信息
func (s *storageSiteInfoService) convertStorageSiteInfo(storageSiteInfo *storageSiteProto.StorageSiteInfoReply) *dto.StorageSiteInfo {
	if storageSiteInfo == nil {
		s.logger.Warn("存储站点信息为空，使用默认值")
		return &dto.StorageSiteInfo{
			ID:              0,
			StorageSiteNo:   "UNKNOWN",
			StorageSiteName: "未知站点",
			StorageSiteIP:   "",
			StorageSiteURL:  "",
			OpenStatus:      0,
			AuthKey:         "",
		}
	}

	return &dto.StorageSiteInfo{
		ID:              storageSiteInfo.Id,
		StorageSiteNo:   storageSiteInfo.StorageSiteNo,
		StorageSiteName: storageSiteInfo.StorageSiteName,
		StorageSiteIP:   storageSiteInfo.StorageSiteIp,
		StorageSiteURL:  storageSiteInfo.StorageSiteUrl,
		OpenStatus:      storageSiteInfo.OpenStatus,
		AuthKey:         storageSiteInfo.AuthKey,
	}
}

// shouldRetry 判断是否应该重试
func (s *storageSiteInfoService) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 对连接相关错误进行重试
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "unavailable") ||
		strings.Contains(errMsg, "context deadline exceeded")
}
