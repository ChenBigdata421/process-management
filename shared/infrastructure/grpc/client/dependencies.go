package grpc_client

import (
	"sync"

	"jxt-evidence-system/process-management/shared/common/di"

	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

var registerOnce sync.Once

// RegisterDependencies 注册所有 gRPC 客户端依赖
func RegisterDependencies() {
	registerOnce.Do(func() {
		registerConnectionManager()
		registerCasbinPolicyProvider()
	})
}

// registerConnectionManager 注册 gRPC 连接管理器（单例）
func registerConnectionManager() {
	if err := di.Provide(func() (*ConnectionManager, error) {
		clientConfig := config.GrpcConfig.Client
		etcdConfig := config.EtcdConfig

		logger.Infof("从配置中获取gRPC客户端配置: ServiceKey=%s, Timeout=%d",
			clientConfig.ServiceKey, clientConfig.Timeout)

		connManager, err := NewConnectionManager(clientConfig, *etcdConfig)
		if err != nil {
			logger.Errorf("创建gRPC连接管理器失败: %v", err)
			return nil, err
		}

		logger.Info("gRPC连接管理器创建成功")
		return connManager, nil
	}); err != nil {
		logger.Fatalf("注册gRPC连接管理器失败: %v", err)
	}
}

// registerCasbinPolicyProvider 注册 Casbin 策略提供者
func registerCasbinPolicyProvider() {
	if err := di.Provide(func(connManager *ConnectionManager) *GrpcCasbinPolicyProvider {
		logger.Info("创建Casbin策略提供者")
		return NewGrpcCasbinPolicyProvider(connManager)
	}); err != nil {
		logger.Fatalf("注册Casbin策略提供者失败: %v", err)
	}
}
