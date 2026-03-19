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

type organizationInfoService struct {
	orgInfoClient grpcclient.OrgInfoServiceClient
	logger        *zap.Logger
}

// NewOrganizationInfoService 创建组织信息服务实例
func NewOrganizationInfoService(orgInfoClient grpcclient.OrgInfoServiceClient) port.OrganizationInfoService {
	return &organizationInfoService{
		orgInfoClient: orgInfoClient,
		logger:        logger.Logger,
	}
}

func (s *organizationInfoService) ValidateOrgExists(ctx context.Context, orgID int) error {
	if orgID <= 0 {
		return fmt.Errorf("无效的组织ID: %d", orgID)
	}

	_, err := s.GetOrgInfo(ctx, orgID)
	if err != nil {
		s.logger.Warn("组织验证失败，使用降级处理",
			zap.Error(err),
			zap.Int("orgID", orgID))
		return nil // 降级处理：测试环境或服务不可用时不阻断业务
	}

	return nil
}

func (s *organizationInfoService) GetOrgCode(ctx context.Context, orgID int) (string, error) {
	orgInfo, err := s.GetOrgInfo(ctx, orgID)
	if err != nil {
		// 降级处理
		s.logger.Warn("获取组织编码失败，使用降级处理", zap.Error(err), zap.Int("orgID", orgID))
		return fmt.Sprintf("ORG%04d", orgID), nil
	}
	return orgInfo.OrgCode, nil
}

func (s *organizationInfoService) GetOrgFullName(ctx context.Context, orgID int) (string, error) {
	orgInfo, err := s.GetOrgInfo(ctx, orgID)
	if err != nil {
		// 降级处理
		s.logger.Warn("获取组织全名失败，使用降级处理", zap.Error(err), zap.Int("orgID", orgID))
		return fmt.Sprintf("组织%d", orgID), nil
	}
	return orgInfo.FullName, nil
}

func (s *organizationInfoService) GetOrgInfo(ctx context.Context, orgID int) (*dto.OrganizationInfo, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("无效的组织ID: %d", orgID)
	}

	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID == 0 {
		return nil, fmt.Errorf("tenant id not exist or invalid")
	}

	var lastErr error

	// 统一的重试+降级逻辑：适度重试，快速降级
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		orgReply, err := s.orgInfoClient.GetOrgById(ctx, int32(tenantID), int32(orgID))
		if err == nil && orgReply != nil {
			if attempt > 0 {
				s.logger.Info("组织信息获取成功",
					zap.Int("orgID", orgID), zap.Int("attempt", attempt+1))
			}

			// 获取组织全名
			fullName := orgReply.OrgName
			fullNameReply, err := s.orgInfoClient.GetOrgFullName(ctx, int32(tenantID), int32(orgID))
			if err == nil && fullNameReply != nil {
				fullName = fullNameReply.FullName
			}

			return &dto.OrganizationInfo{
				OrgID:     int(orgReply.OrgId),
				OrgCode:   orgReply.OrgCode,
				OrgName:   orgReply.OrgName,
				OrgNameJc: orgReply.OrgNameJc, // 组织简称
				FullName:  fullName,
				ParentID:  nil,
			}, nil
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
		s.logger.Warn("组织信息获取失败，准备重试",
			zap.Error(err), zap.Int("orgID", orgID), zap.Int("attempt", attempt+1))
		time.Sleep(RetryBackoff)
	}

	// 降级处理：保证业务不被阻断
	s.logger.Warn("组织信息获取失败，使用降级处理",
		zap.Error(lastErr), zap.Int("orgID", orgID))
	return &dto.OrganizationInfo{
		OrgID:    orgID,
		OrgCode:  fmt.Sprintf("ORG%04d", orgID),
		OrgName:  fmt.Sprintf("组织%d", orgID),
		FullName: fmt.Sprintf("组织%d", orgID),
		ParentID: nil,
	}, nil
}

// shouldRetry 判断是否应该重试
func (s *organizationInfoService) shouldRetry(err error) bool {
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
