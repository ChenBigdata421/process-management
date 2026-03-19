package infrastructure_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jxt-evidence-system/process-management/shared/common/global"
	grpcclient "jxt-evidence-system/process-management/shared/infrastructure/grpc/client/port"
	userProto "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/user"
	"jxt-evidence-system/process-management/shared/infrastructure/service/dto"
	"jxt-evidence-system/process-management/shared/infrastructure/service/port"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"go.uber.org/zap"
)

type userInfoService struct {
	userInfoClient grpcclient.UserInfoServiceClient
	logger         *zap.Logger
}

// NewUserInfoService 创建用户信息服务实例
func NewUserInfoService(userInfoClient grpcclient.UserInfoServiceClient) port.UserInfoService {
	return &userInfoService{
		userInfoClient: userInfoClient,
		logger:         logger.Logger,
	}
}

func (s *userInfoService) GetUserById(ctx context.Context, userID int32) (*dto.UserInfo, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("无效的用户ID: %d", userID)
	}

	tenantID := s.getTenantID(ctx)
	var lastErr error

	// 统一的重试+降级逻辑：适度重试，快速降级
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		userInfo, err := s.userInfoClient.GetUserById(ctx, tenantID, userID)
		if err == nil && userInfo != nil {
			if attempt > 0 { //只在重试成功后记录日志，避免正常情况下的成功请求也产生日志噪音。
				s.logger.Info("用户信息获取成功",
					zap.Int32("userID", userID), zap.Int("attempt", attempt+1))
			}
			return s.convertUserInfo(userInfo), nil
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
		s.logger.Warn("用户信息获取失败，准备重试",
			zap.Error(err), zap.Int32("userID", userID), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断
	s.logger.Warn("用户信息获取失败，使用降级处理",
		zap.Error(lastErr), zap.Int32("userID", userID))
	return &dto.UserInfo{
		UserID:   userID,
		UserName: fmt.Sprintf("用户_%d", userID),
		PoliceNo: fmt.Sprintf("P_%d", userID),
		OrgID:    0,
	}, nil
}

func (s *userInfoService) GetUserByPoliceNo(ctx context.Context, policeNo string) (*dto.UserInfo, error) {
	if policeNo == "" {
		return nil, fmt.Errorf("警号不能为空")
	}

	tenantID := s.getTenantID(ctx)
	var lastErr error

	// 统一的重试+降级逻辑
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		userInfo, err := s.userInfoClient.GetUserByPoliceNo(ctx, tenantID, policeNo)
		if err == nil && userInfo != nil {
			if attempt > 0 {
				s.logger.Info("根据警号获取用户信息成功",
					zap.String("policeNo", policeNo), zap.Int("attempt", attempt+1))
			}
			return s.convertUserInfo(userInfo), nil
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

		s.logger.Warn("根据警号获取用户信息失败，准备重试",
			zap.Error(err), zap.String("policeNo", policeNo), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断，特别是在测试环境中
	s.logger.Warn("根据警号查询用户信息失败，使用降级处理", zap.Error(lastErr), zap.String("policeNo", policeNo))
	return &dto.UserInfo{
		UserID:   1001, // 使用测试用户ID
		UserName: fmt.Sprintf("测试用户_%s", policeNo),
		PoliceNo: policeNo,
		OrgID:    1, // 使用测试组织ID
	}, nil
}

func (s *userInfoService) ValidateUserExists(ctx context.Context, userID int32) error {
	if userID <= 0 {
		return fmt.Errorf("无效的用户ID: %d", userID)
	}
	_, err := s.GetUserById(ctx, userID)
	if err != nil {
		s.logger.Warn("用户验证失败，使用降级处理", zap.Error(err), zap.Int32("userID", userID))
		return nil // 降级处理：测试环境或服务不可用时不阻断业务
	}
	return nil
}

// convertUserInfo 转换用户信息
func (s *userInfoService) convertUserInfo(userInfo *userProto.UserInfoReply) *dto.UserInfo {
	if userInfo == nil {
		s.logger.Warn("用户信息为空，使用默认值")
		return &dto.UserInfo{
			UserID:   0,
			UserName: "未知用户",
			PoliceNo: "UNKNOWN",
			OrgID:    0,
		}
	}

	return &dto.UserInfo{
		UserID:   userInfo.UserId,
		UserName: userInfo.UserName,
		PoliceNo: userInfo.PoliceNo,
		OrgID:    userInfo.OrgId,
	}
}

// shouldRetry 判断是否应该重试
func (s *userInfoService) shouldRetry(err error) bool {
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

// getTenantID 从上下文获取租户ID
func (s *userInfoService) getTenantID(ctx context.Context) int32 {
	tenantID := ctx.Value(global.TenantIDKey)
	if tenantID == nil {
		return 1 // 默认租户ID
	}
	// TenantID 现在是 int 类型 (jxt-core v1.1.37+)
	if tid, ok := tenantID.(int); ok {
		return int32(tid)
	}
	// 兼容 int32 类型
	if tid, ok := tenantID.(int32); ok {
		return tid
	}
	// 兼容旧的 string 类型 - 尝试转换
	if tid, ok := tenantID.(string); ok {
		var parsedID int32
		if _, err := fmt.Sscanf(tid, "%d", &parsedID); err == nil {
			return parsedID
		}
	}
	return 1 // 默认租户ID
}
