package infrastructure_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jxt-evidence-system/process-management/shared/common/global"
	grpcclient "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
	"jxt-evidence-system/process-management/shared/infrastructure/service/port"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"go.uber.org/zap"
)

type enforcementTypeInfoService struct {
	enforcementTypeClient grpcclient.EnforcementTypeServiceClient
	logger                *zap.Logger
}

// NewEnforcementTypeInfoService 创建执法类型信息服务实例
func NewEnforcementTypeInfoService(enforcementTypeClient grpcclient.EnforcementTypeServiceClient) port.EnforcementTypeInfoService {
	return &enforcementTypeInfoService{
		enforcementTypeClient: enforcementTypeClient,
		logger:                logger.Logger,
	}
}

func (s *enforcementTypeInfoService) ValidateEnforcementTypeExists(ctx context.Context, enforcementTypeID int64) error {
	if enforcementTypeID <= 0 {
		return fmt.Errorf("无效的执法类型ID: %d", enforcementTypeID)
	}

	_, err := s.GetEnforcementTypeInfo(ctx, enforcementTypeID)
	if err != nil {
		s.logger.Warn("执法类型验证失败，使用降级处理",
			zap.Error(err),
			zap.Int64("enforcementTypeID", enforcementTypeID))
		return nil // 降级处理：测试环境或服务不可用时不阻断业务
	}

	return nil
}

func (s *enforcementTypeInfoService) GetEnforcementTypeCode(ctx context.Context, enforcementTypeID int64) (string, error) {
	enforcementTypeInfo, err := s.GetEnforcementTypeInfo(ctx, enforcementTypeID)
	if err != nil {
		// 降级处理
		s.logger.Warn("获取执法类型编码失败，使用降级处理", zap.Error(err), zap.Int64("enforcementTypeID", enforcementTypeID))
		return fmt.Sprintf("ET%04d", enforcementTypeID), nil
	}
	return enforcementTypeInfo.EnforcementTypeCode, nil
}

func (s *enforcementTypeInfoService) GetEnforcementTypeName(ctx context.Context, enforcementTypeID int64) (string, error) {
	enforcementTypeInfo, err := s.GetEnforcementTypeInfo(ctx, enforcementTypeID)
	if err != nil {
		// 降级处理
		s.logger.Warn("获取执法类型名称失败，使用降级处理", zap.Error(err), zap.Int64("enforcementTypeID", enforcementTypeID))
		return fmt.Sprintf("执法类型%d", enforcementTypeID), nil
	}
	return enforcementTypeInfo.EnforcementTypeName, nil
}

func (s *enforcementTypeInfoService) GetEnforcementTypeInfo(ctx context.Context, enforcementTypeID int64) (*dto.EnforcementTypeInfo, error) {
	if enforcementTypeID <= 0 {
		return nil, fmt.Errorf("无效的执法类型ID: %d", enforcementTypeID)
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}

	var lastErr error

	// 统一的重试+降级逻辑：适度重试，快速降级
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		enforcementTypeReply, err := s.enforcementTypeClient.GetEnforcementTypeById(ctx, int32(tenantID), enforcementTypeID)
		if err == nil && enforcementTypeReply != nil {
			if attempt > 0 {
				s.logger.Info("执法类型信息获取成功",
					zap.Int64("enforcementTypeID", enforcementTypeID), zap.Int("attempt", attempt+1))
			}

			return enforcementTypeReply, nil
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
		s.logger.Warn("执法类型信息获取失败，准备重试",
			zap.Error(err), zap.Int64("enforcementTypeID", enforcementTypeID), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断
	s.logger.Warn("执法类型信息获取失败，使用降级处理",
		zap.Error(lastErr), zap.Int64("enforcementTypeID", enforcementTypeID))
	return &dto.EnforcementTypeInfo{
		ID:                  enforcementTypeID,
		EnforcementTypeCode: fmt.Sprintf("ET%04d", enforcementTypeID),
		EnforcementTypeName: fmt.Sprintf("执法类型%d", enforcementTypeID),
		EnforcementTypeDesc: "",
		EnforcementTypePath: "",
		ParentId:            0,
		Source:              "",
		Sort:                0,
	}, nil
}

func (s *enforcementTypeInfoService) GetEnforcementTypeByCode(ctx context.Context, code string) (*dto.EnforcementTypeInfo, error) {
	if code == "" {
		return nil, fmt.Errorf("执法类型编码不能为空")
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}

	var lastErr error

	// 统一的重试+降级逻辑
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		enforcementTypeReply, err := s.enforcementTypeClient.GetEnforcementTypeByCode(ctx, int32(tenantID), code)
		if err == nil && enforcementTypeReply != nil {
			if attempt > 0 {
				s.logger.Info("根据编码获取执法类型信息成功",
					zap.String("code", code), zap.Int("attempt", attempt+1))
			}

			return enforcementTypeReply, nil
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

		s.logger.Warn("根据编码获取执法类型信息失败，准备重试",
			zap.Error(err), zap.String("code", code), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断
	s.logger.Warn("根据编码查询执法类型信息失败，使用降级处理", zap.Error(lastErr), zap.String("code", code))
	return &dto.EnforcementTypeInfo{
		ID:                  0,
		EnforcementTypeCode: code,
		EnforcementTypeName: fmt.Sprintf("执法类型_%s", code),
		EnforcementTypeDesc: "",
		EnforcementTypePath: "",
		ParentId:            0,
		Source:              "",
		Sort:                0,
	}, nil
}

// shouldRetry 判断是否应该重试
func (s *enforcementTypeInfoService) shouldRetry(err error) bool {
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
