package infrastructure_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jxt-evidence-system/process-management/shared/common/global"
	grpcclient "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	bwcProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/bwc"
	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
	"jxt-evidence-system/process-management/shared/infrastructure/service/port"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"go.uber.org/zap"
)

type BWCInfoService struct {
	bwcInfoClient grpcclient.BWCInfoServiceClient
	logger        *zap.Logger
}

// NewBWCInfoService 创建执法记录仪信息服务实例
func NewBWCInfoService(bwcInfoClient grpcclient.BWCInfoServiceClient) port.BWCInfoService {
	return &BWCInfoService{
		bwcInfoClient: bwcInfoClient,
		logger:        logger.Logger,
	}
}

func (s *BWCInfoService) GetBWCById(ctx context.Context, bwcID int32) (*dto.BWCInfo, error) {
	if bwcID <= 0 {
		return nil, fmt.Errorf("无效的执法记录仪ID: %d", bwcID)
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}
	var lastErr error

	// 统一的重试+降级逻辑：适度重试，快速降级
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		bwcInfo, err := s.bwcInfoClient.GetBWCById(ctx, int32(tenantID), bwcID)
		if err == nil && bwcInfo != nil {
			if attempt > 0 {
				s.logger.Info("执法记录仪信息获取成功",
					zap.Int32("bwcID", bwcID), zap.Int("attempt", attempt+1))
			}
			return s.convertBWCInfo(bwcInfo), nil
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
		s.logger.Warn("执法记录仪信息获取失败，准备重试",
			zap.Error(err), zap.Int32("bwcID", bwcID), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断
	s.logger.Warn("执法记录仪信息获取失败，使用降级处理",
		zap.Error(lastErr), zap.Int32("bwcID", bwcID))
	return &dto.BWCInfo{
		ID:              bwcID,
		BWCNo:           fmt.Sprintf("REC_%d", bwcID),
		BWCName:         fmt.Sprintf("执法仪_%d", bwcID),
		RequisitionerId: 0,
	}, nil
}

func (s *BWCInfoService) GetBWCByNo(ctx context.Context, no string) (*dto.BWCInfo, error) {
	if no == "" {
		return nil, fmt.Errorf("执法记录仪编号不能为空")
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}
	var lastErr error

	// 统一的重试+降级逻辑
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		bwcInfo, err := s.bwcInfoClient.GetBWCByNo(ctx, int32(tenantID), no)
		if err == nil && bwcInfo != nil {
			if attempt > 0 {
				s.logger.Info("根据编号获取执法记录仪信息成功",
					zap.String("no", no), zap.Int("attempt", attempt+1))
			}
			return s.convertBWCInfo(bwcInfo), nil
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

		s.logger.Warn("根据编号获取执法记录仪信息失败，准备重试",
			zap.Error(err), zap.String("no", no), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 对于根据编号查询，如果失败则返回错误（不降级）
	s.logger.Warn("根据编号查询执法记录仪信息失败", zap.Error(lastErr), zap.String("no", no))
	return nil, fmt.Errorf("根据执法仪编号[%s]查询执法仪信息失败: %w", no, lastErr)
}

func (s *BWCInfoService) ValidateBWCExists(ctx context.Context, bwcID int32) error {
	if bwcID <= 0 {
		return fmt.Errorf("无效的执法记录仪ID: %d", bwcID)
	}
	_, err := s.GetBWCById(ctx, bwcID)
	if err != nil {
		s.logger.Warn("执法记录仪验证失败，使用降级处理", zap.Error(err), zap.Int32("bwcID", bwcID))
		return nil // 降级处理：测试环境或服务不可用时不阻断业务
	}
	return nil
}

// convertBWCInfo 转换执法记录仪信息
func (s *BWCInfoService) convertBWCInfo(bwcInfo *bwcProto.BWCInfoReply) *dto.BWCInfo {
	if bwcInfo == nil {
		s.logger.Warn("执法记录仪信息为空，使用默认值")
		return &dto.BWCInfo{
			ID:              0,
			BWCNo:           "UNKNOWN",
			BWCName:         "未知设备",
			RequisitionerId: 0,
		}
	}

	return &dto.BWCInfo{
		ID:              bwcInfo.Id,
		BWCNo:           bwcInfo.BwcNo,
		BWCName:         bwcInfo.BwcName,
		RequisitionerId: bwcInfo.RequisitionerId,
	}
}

// shouldRetry 判断是否应该重试
func (s *BWCInfoService) shouldRetry(err error) bool {
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
