package grpc_client

import (
	"context"
	"fmt"
	"sync"

	mycasbin "github.com/ChenBigdata421/jxt-core/sdk/pkg/casbin"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	pb "jxt-evidence-system/process-management/shared/infrastructure/grpc/proto/casbin"
)

// GrpcCasbinPolicyProvider implements mycasbin.PolicyProvider interface
// Provides Casbin policies via gRPC calls to security-management service
type GrpcCasbinPolicyProvider struct {
	connManager *ConnectionManager
	client      pb.CasbinPolicyServiceClient
	initOnce    sync.Once
}

// NewGrpcCasbinPolicyProvider creates a new gRPC-based Casbin policy provider
func NewGrpcCasbinPolicyProvider(connManager *ConnectionManager) *GrpcCasbinPolicyProvider {
	return &GrpcCasbinPolicyProvider{
		connManager: connManager,
	}
}

// GetPolicies retrieves all policy rules for the specified tenant via gRPC
// Implements mycasbin.PolicyProvider interface
func (p *GrpcCasbinPolicyProvider) GetPolicies(ctx context.Context, tenantID int) ([]mycasbin.PolicyRule, error) {
	// 校验租户ID合法性
	if tenantID <= 0 {
		return nil, fmt.Errorf("invalid tenant ID: %d (must be positive)", tenantID)
	}
	if tenantID > 2147483647 { // math.MaxInt32
		return nil, fmt.Errorf("tenant ID overflow: %d (exceeds int32 max)", tenantID)
	}

	// Lazy initialization of gRPC client
	p.initOnce.Do(func() {
		p.client = pb.NewCasbinPolicyServiceClient(p.connManager.GetConnection())
		logger.Info("Casbin 策略服务客户端初始化完成")
	})

	if p.client == nil {
		return nil, fmt.Errorf("casbin policy client not initialized")
	}

	req := &pb.GetPoliciesRequest{TenantId: int32(tenantID)}

	var resp *pb.GetPoliciesResponse
	err := p.connManager.ExecuteWithNetworkRetry(ctx, func() error {
		var err error
		resp, err = p.client.GetPolicies(ctx, req)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("gRPC GetPolicies 失败: %w", err)
	}

	// 检查空策略（可能是租户未配置权限）
	if len(resp.Policies) == 0 {
		logger.Warnf("租户 %d 策略为空，可能未配置权限或策略已清空", tenantID)
		return []mycasbin.PolicyRule{}, nil // 返回空切片（非nil）
	}

	// 转换并校验策略数据
	rules := make([]mycasbin.PolicyRule, 0, len(resp.Policies))
	invalidCount := 0
	for _, policy := range resp.Policies {
		// 校验策略类型
		if policy.PType != "p" && policy.PType != "g" {
			logger.Warnf("租户 %d 跳过非法策略类型: %s", tenantID, policy.PType)
			invalidCount++
			continue
		}
		// 校验必填字段（p类型至少需要V0, V1, V2；g类型至少需要V0, V1）
		if policy.V0 == "" || policy.V1 == "" {
			logger.Warnf("租户 %d 跳过不完整策略: PType=%s, V0=%s, V1=%s",
				tenantID, policy.PType, policy.V0, policy.V1)
			invalidCount++
			continue
		}

		rules = append(rules, mycasbin.PolicyRule{
			PType: policy.PType,
			V0:    policy.V0,
			V1:    policy.V1,
			V2:    policy.V2,
			V3:    policy.V3,
			V4:    policy.V4,
			V5:    policy.V5,
		})
	}

	if invalidCount > 0 {
		logger.Warnf("租户 %d 跳过 %d 条非法策略，有效策略 %d 条", tenantID, invalidCount, len(rules))
	}

	logger.Infof("租户 %d 加载策略 %d 条", tenantID, len(rules))
	return rules, nil
}
