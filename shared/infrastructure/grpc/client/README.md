# gRPC 客户端

证据管理系统的 gRPC 客户端实现，提供与外部微服务通信的统一接口。采用现代的微服务架构设计模式，具备高可用性、自动重连、健康监控等企业级特性。

> **🎯 设计理念**  
> 本客户端采用职责分离的设计原则，将连接管理与业务逻辑完全分离，符合 SOLID 原则，提供高度可扩展和可维护的架构。

## 📋 目录

- [设计演进](#design-evolution)
- [架构设计](#architecture)
- [功能特性](#features)
- [快速开始](#quick-start)
- [配置说明](#configuration)
- [服务客户端](#service-clients)
- [连接管理](#connection-management)
- [网络重试机制](#network-retry)
- [监控与日志](#monitoring)
- [最佳实践](#best-practices)
- [故障排查](#troubleshooting)
- [开发指南](#development)
- [依赖注入](#dependency-injection)

## 🎯 设计演进：从职责混乱到职责分离 {#design-evolution}

### 问题起源

**核心问题**：传统的 gRPC 客户端设计往往将连接管理和业务逻辑混合在一起，违反了单一职责原则。

**关键洞察**：如果客户端只提供 gRPC 连接管理，不实现具体的业务方法，会不会更好？

**答案**：完全正确！这正是优秀架构设计的体现。

### ❌ 传统设计的问题

```go
// 职责混乱的设计
type SharedGrpcClient struct {
    // 连接管理职责
    client zrpc.Client
    isHealthy bool
    
    // 业务方法职责 - 这里有问题！
    GetUserById(...)     
    GetUserByPoliceNo(...)
    GetOrgById(...)      
    ValidateToken(...)   
    // 每添加新服务都要在这里加方法
}
```

**问题症状**：
1. **违反单一职责原则 (SRP)**：一个类承担了连接管理和业务逻辑两个职责
2. **难以扩展和维护**：添加新服务需要修改核心客户端类
3. **职责边界模糊**：不清楚哪些功能属于连接管理，哪些属于业务逻辑
4. **测试困难**：无法单独测试连接逻辑和业务逻辑
5. **接口混乱**：一个类暴露所有服务的方法

### ✅ 改进后的设计

```go
// 1. 连接管理器 - 专注连接管理
type ConnectionManager struct {
    client zrpc.Client
    config config.GRPCClientConfig
    // 只负责连接相关功能
}

// 2. 服务适配器 - 专注业务逻辑
type UserInfoServiceClient struct {
    connectionManager *ConnectionManager
    userClient        userProto.UserInfoServiceClient
    // 只负责用户相关业务
}

type OrgInfoServiceClient struct {
    connectionManager *ConnectionManager
    orgClient         orgProto.OrgInfoServiceClient
    // 只负责组织相关业务
}
```

## 🏗️ 架构设计 {#architecture}

### 详细架构图

```
应用层 (Application Layer)
┌─────────────────────────┐    ┌─────────────────────────┐
│    MediaService         │    │   PermissionService     │
│                         │    │                         │
│ userInfoClient ────────┐│    │ userInfoClient ────────┐│
│ lawCameraClient ───────┼│    │ orgInfoClient ─────────┼│
└────────────────────────┼┘    └────────────────────────┼┘
                         │                              │
接口层 (Domain Ports)    │                              │
┌─────────────────────────┼─────────────────────────────┼┘
│ UserInfoServiceClient   │     OrgInfoServiceClient    │
│ ├─ GetUserById()        │     ├─ GetOrgById()         │
│ └─ GetUserByPoliceNo()  │     └─ GetOrgFullName()     │
│                         │                             │
│ BWCInfoServiceClient   AuthServiceClient        │
│ ├─ GetBWCById()   │     ├─ ValidateToken()      │
│ └─ GetBWCByNo()   │     └─ RefreshToken()       │
└─────────────────────────┼─────────────────────────────┼┘
                         │                              │
适配器层 (Adapters)      │                              │
┌─────────────────────────┼─────────────────────────────┼┐
│                         ▼                             ▼│
│ UserInfoServiceClient   OrgInfoServiceClient           │
│ BWCInfoServiceClient   AuthServiceClient        │
│ │                    │                                 │
│ └─────────────┐      └──────────────┐                  │
│               ▼                     ▼                  │
│           ┌─────────────────────────────────────────┐  │
│           │      ConnectionManager                  │  │
│           │  ┌─────────────────────────────────────┐│  │
│           │  │       zrpc.Client                   ││  │
│           │  │   (单一gRPC连接)                   ││  │
│           │  │   + 自动重连机制                    ││  │
│           │  │   + 健康检查                        ││  │
│           │  │   + 网络重试                        ││  │
│           │  └─────────────────────────────────────┘│  │
│           └─────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────┘
                         │
底层连接 (Infrastructure) │
                         ▼
         etcd://security-management/services
                (共享 serviceKey)
```

### 设计原则体现

#### 1. 单一职责原则 (SRP)
- `ConnectionManager`：只负责连接管理
- `UserInfoServiceClient`：只负责用户业务逻辑
- `OrgInfoServiceClient`：只负责组织业务逻辑
- `BWCInfoServiceClient`：只负责执法记录仪业务逻辑
- **效果**：每个类都有且仅有一个变更的理由

#### 2. 开闭原则 (OCP)
- 添加新服务时，不需要修改现有代码
- 只需要创建新的适配器实现新接口
- 连接管理器对扩展开放，对修改封闭
- **效果**：系统易于扩展，稳定性高

#### 3. 接口隔离原则 (ISP)
- 每个服务有独立的接口
- 应用服务只依赖需要的接口
- 避免了大而全的接口
- **效果**：客户端不依赖它们不使用的接口

#### 4. 依赖倒置原则 (DIP)
- 应用服务依赖接口抽象
- 适配器实现具体接口
- 高层不依赖低层具体实现
- **效果**：系统灵活性和可替换性高

### 架构模式

- **六边形架构（Hexagonal Architecture）**：通过端口和适配器模式实现业务逻辑与基础设施分离
- **CQRS（Command Query Responsibility Segregation）**：命令和查询分离，提高系统的可维护性
- **适配器模式（Adapter Pattern）**：每个服务客户端都是一个适配器，将外部服务接口适配为内部接口
- **依赖注入（Dependency Injection）**：通过构造函数注入依赖，便于测试和扩展

## ✨ 功能特性 {#features}

### 🔄 自动重连机制
- **智能重连**：网络故障时自动重连，默认最多重试 5 次
- **退避策略**：采用固定间隔重连策略，默认 10 秒间隔
- **健康检查**：定期检查连接状态，默认 30 秒间隔
- **优雅降级**：重连期间提供明确的错误信息
- **连接重建**：自动关闭旧连接并重新创建，确保连接质量

### 🔍 连接监控
- **实时状态**：实时监控连接状态（Ready、Idle、Connecting、TransientFailure、Shutdown）
- **重连状态**：追踪重连尝试次数和状态
- **健康指标**：提供连接健康状态查询接口
- **状态变更通知**：连接状态变化时自动触发相应处理

### 🛡️ 多层错误处理
- **连接级错误**：ConnectionManager 处理连接相关错误
- **网络级重试**：`ExecuteWithNetworkRetry` 提供透明的网络重试
- **业务级错误**：服务适配器处理业务逻辑错误
- **错误分类**：智能识别网络错误、业务错误和系统错误
- **详细日志**：提供详细的调用日志和错误信息

### 🚀 高性能特性
- **连接复用**：多个服务客户端共享同一个 gRPC 连接
- **懒加载**：客户端懒加载初始化，减少启动时间
- **并发安全**：所有客户端都是线程安全的
- **服务发现**：集成 ETCD 服务发现机制
- **网络重试**：自动处理网络级别的临时故障

### 🎯 职责分离特性
- **接口隔离**：每个服务有独立的接口定义
- **依赖注入**：完整的 DI 支持，便于测试和扩展
- **模块化设计**：各服务客户端可独立开发和维护
- **易于扩展**：添加新服务无需修改现有代码

## 🚀 快速开始 {#quick-start}

### 1. 依赖注入方式使用（推荐）

```go
package main

import (
    "context"
    "log"
    
    client "jxt-evidence-system/evidence-management/shared/common/grpc/client/port"
    grpc_client "jxt-evidence-system/evidence-management/shared/common/grpc/client"
    "github.com/ChenBigdata421/jxt-core/sdk/di"
)

// 应用服务
type MediaService struct {
    userInfoClient client.UserInfoServiceClient
}

func NewMediaService(userClient client.UserInfoServiceClient) *MediaService {
    return &MediaService{
        userInfoClient: userClient,
    }
}

func (s *MediaService) ProcessMedia(ctx context.Context, policeNo string) error {
    // 使用注入的客户端
    user, err := s.userInfoClient.GetUserByPoliceNo(ctx, "tenant1", policeNo)
    if err != nil {
        return err
    }
    
    log.Printf("处理用户 %s 的媒体", user.UserName)
    return nil
}

func main() {
    // 1. 注册依赖
    grpc_client.RegisterDependencies()
    
    // 2. 注册应用服务
    di.Provide(NewMediaService)
    
    // 3. 获取应用服务实例
    var mediaService *MediaService
    if err := di.Invoke(func(service *MediaService) {
        mediaService = service
    }); err != nil {
        log.Fatalf("获取服务实例失败: %v", err)
    }
    
    // 4. 使用服务
    ctx := context.Background()
    if err := mediaService.ProcessMedia(ctx, "001234"); err != nil {
        log.Fatalf("处理媒体失败: %v", err)
    }
}
```

### 2. 手动创建方式（测试或特殊场景）

```go
package main

import (
    "context"
    "log"
    
    grpc_client "jxt-evidence-system/evidence-management/shared/common/grpc/client"
    "github.com/ChenBigdata421/jxt-core/sdk/config"
)

func main() {
    // 1. 配置 gRPC 客户端
    grpcConfig := config.GRPCClientConfig{
        ServiceKey: "security-management/services",
        Timeout:    5000, // 5秒超时
    }
    
    etcdConfig := config.ETCDConfig{
        Hosts: []string{"localhost:2379"},
    }
    
    // 2. 创建连接管理器
    connManager, err := grpc_client.NewConnectionManager(grpcConfig, etcdConfig)
    if err != nil {
        log.Fatalf("创建连接管理器失败: %v", err)
    }
    defer connManager.Close()
    
    // 3. 创建用户信息客户端
    userClient := grpc_client.NewUserInfoServiceClient(connManager)
    
    // 4. 调用服务
    ctx := context.Background()
    user, err := userClient.GetUserById(ctx, "tenant1", 123)
    if err != nil {
        log.Fatalf("查询用户失败: %v", err)
    }
    
    log.Printf("用户信息: ID=%d, 姓名=%s", user.UserId, user.UserName)
}
```

### 3. 完整集成示例

```go
// main.go
package main

import (
    "context"
    "fmt"
    "time"
    
    grpc_client "jxt-evidence-system/evidence-management/shared/common/grpc/client"
    "github.com/ChenBigdata421/jxt-core/sdk/config"
)

type GRPCClients struct {
    UserInfo     grpc_client.UserInfoServiceClient
    OrgInfo      grpc_client.OrgInfoServiceClient
    LawCamera    grpc_client.BWCInfoServiceClient
    connManager  *grpc_client.ConnectionManager
}

func NewGRPCClients() (*GRPCClients, error) {
    // 配置信息
    etcdConfig := config.ETCDConfig{
        Hosts: []string{"localhost:2379"},
    }
    
    // 创建连接管理器
    grpcConfig := config.GRPCClientConfig{
        ServiceKey: "external.services.rpc",
        Timeout:    5000,
    }
    
    connManager, err := grpc_client.NewConnectionManager(grpcConfig, etcdConfig)
    if err != nil {
        return nil, fmt.Errorf("创建连接管理器失败: %w", err)
    }
    
    // 创建各个服务客户端
    clients := &GRPCClients{
        UserInfo:    grpc_client.NewUserInfoServiceClient(connManager),
        OrgInfo:     grpc_client.NewOrgInfoServiceClient(connManager),
        LawCamera:   grpc_client.NewBWCInfoServiceClient(connManager),
        connManager: connManager,
    }
    
    return clients, nil
}

func (c *GRPCClients) Close() error {
    return c.connManager.Close()
}

func (c *GRPCClients) HealthCheck() {
    fmt.Printf("连接健康状态: %v\n", c.connManager.IsHealthy())
    fmt.Printf("连接状态: %v\n", c.connManager.GetConnectionState())
    fmt.Printf("重连次数: %d\n", c.connManager.GetReconnectAttempts())
}

func main() {
    clients, err := NewGRPCClients()
    if err != nil {
        panic(err)
    }
    defer clients.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 使用示例
    tenantID := "tenant1"
    
    // 查询用户信息
    user, err := clients.UserInfo.GetUserById(ctx, tenantID, 123)
    if err != nil {
        fmt.Printf("查询用户失败: %v\n", err)
    } else {
        fmt.Printf("用户: %s\n", user.UserName)
    }
    
    // 查询组织信息
    org, err := clients.OrgInfo.GetOrgById(ctx, tenantID, 456)
    if err != nil {
        fmt.Printf("查询组织失败: %v\n", err)
    } else {
        fmt.Printf("组织: %s\n", org.OrgName)
    }
    
    // 健康检查
    clients.HealthCheck()
}
```

## ⚙️ 配置说明 {#configuration}

### 实际配置结构

基于 SDK 的配置结构和实际代码实现：

```go
// 实际使用的gRPC客户端配置（来自SDK）
type config.GRPCClientConfig struct {
    ServiceKey string `json:"service_key" yaml:"service_key"`
    Timeout    int64  `json:"timeout" yaml:"timeout"`
}

// ETCD配置（来自SDK）
type config.ETCDConfig struct {
    Hosts []string `json:"hosts" yaml:"hosts"`
}

// 重连专用配置（在ConnectionManager中定义）
type ReconnectConfig struct {
    MaxRetries     int32         `json:"max_retries" yaml:"max_retries" default:"5"`
    RetryInterval  time.Duration `json:"retry_interval" yaml:"retry_interval" default:"10s"`
    HealthInterval time.Duration `json:"health_interval" yaml:"health_interval" default:"30s"`
}
```

### 配置文件示例

```yaml
# config/settings.yml - SDK配置
grpc:
  client:
    service_key: "security-management/services"  # 服务发现Key
    timeout: 5000                                # 请求超时时间(ms)

etcd:
  hosts:
    - "etcd-1:2379"
    - "etcd-2:2379" 
    - "etcd-3:2379"

# 重连配置（当前使用硬编码默认值，未来可考虑配置化）
# reconnect:
#   max_retries: 5        # 最大重连次数
#   retry_interval: 10s   # 重连间隔
#   health_interval: 30s  # 健康检查间隔

# 环境特定配置
production:
  grpc:
    client:
      timeout: 3000       # 生产环境更短超时

development:
  grpc:
    client:
      timeout: 10000      # 开发环境更长超时
```

### 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `grpc.client.service_key` | string | - | ETCD 中的服务发现键 |
| `grpc.client.timeout` | int64 | 5000 | 请求超时时间（毫秒） |
| `etcd.hosts` | []string | - | ETCD 集群地址列表 |
| `reconnect.max_retries` | int32 | 5 | 最大重连次数（硬编码） |
| `reconnect.retry_interval` | duration | 10s | 重连间隔时间（硬编码） |
| `reconnect.health_interval` | duration | 30s | 健康检查间隔（硬编码） |

### 配置加载实现

```go
// 配置加载（由SDK负责）
func NewConnectionManager(cfg config.GRPCClientConfig, etcdConfig config.ETCDConfig) (*ConnectionManager, error) {
    logger.Infof("创建gRPC连接管理器，配置: ServiceKey=%s, Timeout=%d",
        cfg.ServiceKey, cfg.Timeout)

    // 构建zrpc配置 - 使用服务发现模式
    conf := zrpc.RpcClientConf{
        Etcd: discov.EtcdConf{
            Hosts: etcdConfig.Hosts,  // 从SDK全局配置获取
            Key:   cfg.ServiceKey,
        },
        Timeout: cfg.Timeout,
    }

    client, err := zrpc.NewClient(conf)
    if err != nil {
        logger.Errorf("创建zrpc客户端失败: %v", err)
        return nil, err
    }

    // 获取重连配置（当前使用默认值）
    reconnectCfg := getDefaultReconnectConfig()

    manager := &ConnectionManager{
        client:              client,
        config:              cfg,
        isHealthy:           true,
        maxReconnectCount:   reconnectCfg.MaxRetries,
        reconnectInterval:   reconnectCfg.RetryInterval,
        healthCheckInterval: reconnectCfg.HealthInterval,
        stopCh:              make(chan struct{}),
    }

    // 启动健康检查
    go manager.startHealthCheck()

    logger.Infof("gRPC连接管理器创建完成，重连配置: 最大重试%d次, 重试间隔%v, 健康检查间隔%v", 
        reconnectCfg.MaxRetries, reconnectCfg.RetryInterval, reconnectCfg.HealthInterval)
    return manager, nil
}
```

## 🔧 服务客户端 {#service-clients}

### UserInfoServiceClient - 用户信息服务

```go
type UserInfoServiceClient interface {
    // GetUserById 根据用户ID查询用户信息
    GetUserById(ctx context.Context, tenantId string, userId int32) (*proto.UserInfoReply, error)
    
    // GetUserByPoliceNo 根据警号查询用户信息  
    GetUserByPoliceNo(ctx context.Context, tenantId string, policeNo string) (*proto.UserInfoReply, error)
}
```

**使用示例：**
```go
// 根据ID查询用户
user, err := userClient.GetUserById(ctx, "tenant1", 123)
if err != nil {
    return fmt.Errorf("查询用户失败: %w", err)
}

// 根据警号查询用户
user, err := userClient.GetUserByPoliceNo(ctx, "tenant1", "001234")
if err != nil {
    return fmt.Errorf("查询用户失败: %w", err)
}
```

### OrgInfoServiceClient - 组织信息服务

```go
type OrgInfoServiceClient interface {
    // GetOrgById 根据组织ID查询组织信息
    GetOrgById(ctx context.Context, tenantId string, orgId int32) (*proto.OrgInfoReply, error)
    
    // GetOrgByCode 根据组织编码查询组织信息
    GetOrgByCode(ctx context.Context, tenantId string, orgCode string) (*proto.OrgInfoReply, error)
    
    // GetOrgByName 根据组织名称查询组织信息
    GetOrgByName(ctx context.Context, tenantId string, orgName string) (*proto.OrgInfoReply, error)
    
    // GetOrgFullName 获取组织全名（包含上级路径）
    GetOrgFullName(ctx context.Context, tenantId string, orgId int32) (*proto.OrgFullNameReply, error)
}
```

**使用示例：**
```go
// 查询组织信息
org, err := orgClient.GetOrgById(ctx, "tenant1", 456)
if err != nil {
    return fmt.Errorf("查询组织失败: %w", err)
}

// 获取组织全路径名称
fullName, err := orgClient.GetOrgFullName(ctx, "tenant1", 456)
if err != nil {
    return fmt.Errorf("查询组织全名失败: %w", err)
}
```

### BWCInfoServiceClient - 执法记录仪服务

```go
type BWCInfoServiceClient interface {
    // GetBWCById 根据执法记录仪ID查询信息
    GetBWCById(ctx context.Context, tenantId string, id int32) (*proto.BWCInfoReply, error)
    
    // GetBWCByNo 根据执法记录仪编号查询信息
    GetBWCByNo(ctx context.Context, tenantId string, no string) (*proto.BWCInfoReply, error)
    
    // GetBWCsByManagerId 根据管理员ID查询执法记录仪列表
    GetBWCsByManagerId(ctx context.Context, tenantId string, managerId int32) (*proto.BWCListReply, error)
    
    // GetBWCsByRequisitionerId 根据领用人ID查询执法记录仪列表
    GetBWCsByRequisitionerId(ctx context.Context, tenantId string, requisitionerId int32) (*proto.BWCListReply, error)
}
```

**使用示例：**
```go
// 查询执法记录仪信息
camera, err := lawCameraClient.GetBWCByNo(ctx, "tenant1", "DV001234")
if err != nil {
    return fmt.Errorf("查询执法记录仪失败: %w", err)
}

// 查询管理员的执法记录仪列表
cameras, err := lawCameraClient.GetBWCsByManagerId(ctx, "tenant1", 123)
if err != nil {
    return fmt.Errorf("查询管理员设备失败: %w", err)
}
```

## 🔗 连接管理 {#connection-management}

### ConnectionManager - 连接管理器

ConnectionManager 是 gRPC 客户端的核心组件，负责连接的创建、维护、监控和自动重连。

#### 主要功能

1. **连接创建与管理**
   - 基于 ETCD 服务发现创建连接
   - 懒加载连接初始化
   - 连接复用和线程安全

2. **健康监控**
   - 定期检查连接状态
   - 实时健康状态查询
   - 连接状态变更通知

3. **自动重连**
   - 网络故障自动检测
   - 智能重连策略
   - 重连状态跟踪

4. **容错处理**
   - 网络重试机制
   - 超时等待机制
   - 优雅错误处理

#### 核心方法

```go
type ConnectionManager struct {
    // ... 内部字段
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(cfg config.GRPCClientConfig, etcdConfig config.ETCDConfig) (*ConnectionManager, error)

// GetConnection 获取gRPC连接（懒加载）
func (m *ConnectionManager) GetConnection() *grpc.ClientConn

// IsHealthy 检查连接是否健康
func (m *ConnectionManager) IsHealthy() bool

// GetConnectionState 获取连接状态
func (m *ConnectionManager) GetConnectionState() connectivity.State

// CheckHealth 主动触发健康检查
func (m *ConnectionManager) CheckHealth()

// ExecuteWithNetworkRetry 带网络重试的操作执行
func (m *ConnectionManager) ExecuteWithNetworkRetry(ctx context.Context, operation func() error) error

// IsReconnecting 检查是否正在重连
func (m *ConnectionManager) IsReconnecting() bool

// GetReconnectAttempts 获取重连尝试次数
func (m *ConnectionManager) GetReconnectAttempts() int32

// Close 关闭连接管理器
func (m *ConnectionManager) Close() error
```

#### 连接状态

| 状态 | 说明 |
|------|------|
| `IDLE` | 空闲状态，可以接受请求 |
| `CONNECTING` | 正在连接中 |
| `READY` | 连接就绪，可以正常服务 |
| `TRANSIENT_FAILURE` | 暂时故障，会自动重连 |
| `SHUTDOWN` | 连接已关闭 |

#### 重连策略

```go
// 重连配置
type ReconnectConfig struct {
    MaxRetries     int32         // 最大重连次数，默认5次
    RetryInterval  time.Duration // 重连间隔，默认10秒
    HealthInterval time.Duration // 健康检查间隔，默认30秒
}
```

**重连流程：**

1. 健康检查检测到连接不健康
2. 触发自动重连机制
3. 等待重连间隔时间
4. 尝试重新创建连接
5. 检查连接状态
6. 如果失败且未达到最大次数，继续重试
7. 达到最大次数后停止重连，记录错误

## 🔄 网络重试机制 {#network-retry}

### ExecuteWithNetworkRetry - 透明网络重试

ConnectionManager 提供了 `ExecuteWithNetworkRetry` 方法，为业务层提供透明的网络级别重试机制。

#### 工作原理

```go
// 网络重试流程
func (m *ConnectionManager) ExecuteWithNetworkRetry(ctx context.Context, operation func() error) error {
    maxNetworkRetries := 3 // 网络级别最多重试3次

    for attempt := 0; attempt < maxNetworkRetries; attempt++ {
        // 1. 等待连接健康（如果正在重连中）
        if err := m.waitForHealthyConnection(ctx, 5*time.Second); err != nil {
            // 连接不可用，继续下一次重试
            continue
        }

        // 2. 执行业务操作
        err := operation()
        if err == nil {
            return nil // 成功，直接返回
        }

        // 3. 判断是否为网络错误
        if !m.isNetworkError(err) {
            return err // 非网络错误，直接返回
        }

        // 4. 网络错误，触发健康检查和重试
        m.CheckHealth()
        
        // 5. 等待后重试（指数退避：1s, 2s, 3s）
        if attempt < maxNetworkRetries-1 {
            retryDelay := time.Duration(attempt+1) * time.Second
            time.Sleep(retryDelay)
        }
    }

    return fmt.Errorf("网络重试失败，已达最大重试次数 %d", maxNetworkRetries)
}
```

#### 网络错误识别

系统能够智能识别以下网络错误模式：

```go
networkErrorPatterns := []string{
    "connection refused",
    "connection reset by peer", 
    "connection closed",
    "the client connection is closing",
    "transport is closing",
    "rpc error: code = Unavailable",
    "rpc error: code = DeadlineExceeded",
    "rpc error: code = Canceled desc = grpc: the client connection is closing",
    "dial tcp",
    "no such host",
    "network is unreachable", 
    "timeout",
    "context deadline exceeded",
}
```

#### 使用示例

```go
// 在服务适配器中使用网络重试
func (c *UserInfoServiceClient) GetUserById(ctx context.Context, tenantId string, userId int32) (*userProto.UserInfoReply, error) {
    c.initOnce()

    req := &userProto.GetUserByIdReq{
        TenantId: tenantId,
        UserId:   userId,
    }

    var resp *userProto.UserInfoReply
    err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
        var err error
        resp, err = c.userClient.GetUserById(ctx, req)
        return err
    })

    if err != nil {
        return nil, fmt.Errorf("查询用户信息失败: %w", err)
    }

    return resp, nil
}
```

#### 重试策略

| 重试次数 | 等待时间 | 说明 |
|---------|---------|------|
| 第1次 | 立即 | 首次调用 |
| 第2次 | 1秒 | 第一次重试 |
| 第3次 | 2秒 | 第二次重试 |
| 第4次 | 3秒 | 第三次重试 |

#### 与自动重连的关系

```
网络重试 (ExecuteWithNetworkRetry)
    ↓
检测到网络错误
    ↓
触发健康检查 (CheckHealth)
    ↓
发现连接不健康
    ↓
启动自动重连 (attemptReconnect)
    ↓
重建连接
    ↓
网络重试继续执行
```

#### 优势特性

1. **对业务透明**：业务代码无需关心网络重试逻辑
2. **智能错误识别**：只对网络错误进行重试，避免无意义重试
3. **与重连协同**：网络重试会触发连接重建，提高成功率
4. **可配置策略**：重试次数和间隔可以根据需要调整
5. **上下文感知**：支持 context 取消和超时

## 📊 监控与日志 {#monitoring}

### 日志级别

- **DEBUG**：详细的调用信息，包括请求参数和响应数据
- **INFO**：重要的操作信息，如连接创建、重连成功等
- **WARN**：警告信息，如连接不健康、查询结果为空等
- **ERROR**：错误信息，如连接失败、查询异常等

### 日志示例

```log
2024-01-15 14:30:25.123 INFO  创建gRPC连接管理器，配置: ServiceKey=external.services.rpc, Timeout=5000
2024-01-15 14:30:25.456 INFO  gRPC连接管理器创建完成，重连配置: 最大重试5次, 重试间隔10s, 健康检查间隔30s
2024-01-15 14:30:25.789 INFO  用户服务客户端初始化完成
2024-01-15 14:30:26.012 DEBUG 调用GetUserById，tenantId: tenant1, userId: 123
2024-01-15 14:30:26.345 DEBUG GetUserById成功，userId: 123, 用户名: 张三
2024-01-15 14:35:30.678 WARN  gRPC连接不健康，状态: TRANSIENT_FAILURE
2024-01-15 14:35:30.901 INFO  检测到连接状态从健康变为不健康，启动重连机制
2024-01-15 14:35:30.902 INFO  开始尝试重连gRPC连接
2024-01-15 14:35:30.903 INFO  第 1/5 次重连尝试
2024-01-15 14:35:41.234 INFO  重连成功，连接状态: READY
2024-01-15 14:35:41.235 INFO  gRPC连接已恢复健康
```

### 监控指标

你可以通过以下方法获取监控指标：

```go
func monitorGRPCClients(connManager *grpc_client.ConnectionManager) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // 连接健康状态
            isHealthy := connManager.IsHealthy()
            
            // 连接状态
            state := connManager.GetConnectionState()
            
            // 重连状态
            isReconnecting := connManager.IsReconnecting()
            reconnectAttempts := connManager.GetReconnectAttempts()
            
            // 记录监控指标
            logger.Info("gRPC连接监控",
                "healthy", isHealthy,
                "state", state.String(),
                "reconnecting", isReconnecting,
                "reconnect_attempts", reconnectAttempts,
            )
            
            // 可以将这些指标发送到监控系统（如 Prometheus）
            // metrics.GRPCConnectionHealthy.Set(boolToFloat64(isHealthy))
            // metrics.GRPCReconnectAttempts.Set(float64(reconnectAttempts))
        }
    }
}
```

## 🎯 最佳实践 {#best-practices}

### 1. 连接管理

```go
// ✅ 推荐：单个服务使用一个连接管理器
func NewServiceClients() (*ServiceClients, error) {
    // 为每个外部服务创建独立的连接管理器
    userConnManager, err := grpc_client.NewConnectionManager(userGRPCConfig, etcdConfig)
    if err != nil {
        return nil, err
    }
    
    orgConnManager, err := grpc_client.NewConnectionManager(orgGRPCConfig, etcdConfig)
    if err != nil {
        userConnManager.Close()
        return nil, err
    }
    
    return &ServiceClients{
        UserClient: grpc_client.NewUserInfoServiceClient(userConnManager),
        OrgClient:  grpc_client.NewOrgInfoServiceClient(orgConnManager),
        userConn:   userConnManager,
        orgConn:    orgConnManager,
    }, nil
}

// ❌ 不推荐：为每个客户端创建新的连接管理器
func badExample() {
    userConn1, _ := grpc_client.NewConnectionManager(config1, etcdConfig)
    userConn2, _ := grpc_client.NewConnectionManager(config1, etcdConfig) // 重复连接
}
```

### 2. 错误处理

```go
// ✅ 推荐：详细的错误处理
func getUserInfo(ctx context.Context, client grpc_client.UserInfoServiceClient, userID int32) (*proto.UserInfoReply, error) {
    user, err := client.GetUserById(ctx, "tenant1", userID)
    if err != nil {
        // 检查错误类型
        if status.Code(err) == codes.NotFound {
            return nil, fmt.Errorf("用户不存在: userID=%d", userID)
        }
        if status.Code(err) == codes.Unavailable {
            return nil, fmt.Errorf("用户服务暂时不可用，请稍后重试: %w", err)
        }
        return nil, fmt.Errorf("查询用户信息失败: userID=%d, error=%w", userID, err)
    }
    
    // 检查返回结果
    if user == nil {
        return nil, fmt.Errorf("用户信息查询结果为空: userID=%d", userID)
    }
    
    return user, nil
}

// ❌ 不推荐：简单的错误传递
func badErrorHandling(ctx context.Context, client grpc_client.UserInfoServiceClient, userID int32) (*proto.UserInfoReply, error) {
    return client.GetUserById(ctx, "tenant1", userID) // 直接返回，没有上下文
}
```

### 3. 超时控制

```go
// ✅ 推荐：合理的超时控制
func queryUserWithTimeout(client grpc_client.UserInfoServiceClient, userID int32) (*proto.UserInfoReply, error) {
    // 为每个请求设置合适的超时时间
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return client.GetUserById(ctx, "tenant1", userID)
}

// ❌ 不推荐：没有超时控制
func badTimeoutHandling(client grpc_client.UserInfoServiceClient, userID int32) (*proto.UserInfoReply, error) {
    return client.GetUserById(context.Background(), "tenant1", userID) // 可能无限等待
}
```

### 4. 资源清理

```go
// ✅ 推荐：正确的资源清理
func main() {
    clients, err := NewServiceClients()
    if err != nil {
        log.Fatal(err)
    }
    
    // 确保资源被正确清理
    defer func() {
        if err := clients.Close(); err != nil {
            log.Printf("关闭客户端连接失败: %v", err)
        }
    }()
    
    // 使用客户端...
}

// 优雅关闭处理
func gracefulShutdown(clients *ServiceClients) {
    // 停止接受新请求
    // ...
    
    // 等待现有请求完成
    time.Sleep(2 * time.Second)
    
    // 关闭连接
    if err := clients.Close(); err != nil {
        log.Printf("关闭连接失败: %v", err)
    }
    
    log.Println("服务已优雅关闭")
}
```

### 5. 并发使用

```go
// ✅ 推荐：客户端是并发安全的
func concurrentUsage(client grpc_client.UserInfoServiceClient) {
    var wg sync.WaitGroup
    
    // 并发查询多个用户
    userIDs := []int32{1, 2, 3, 4, 5}
    for _, userID := range userIDs {
        wg.Add(1)
        go func(id int32) {
            defer wg.Done()
            
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            user, err := client.GetUserById(ctx, "tenant1", id)
            if err != nil {
                log.Printf("查询用户 %d 失败: %v", id, err)
                return
            }
            
            log.Printf("用户 %d: %s", id, user.UserName)
        }(userID)
    }
    
    wg.Wait()
}
```

## 🐛 故障排查 {#troubleshooting}

### 常见问题

#### 1. 连接超时

**症状：**
```
Error: context deadline exceeded
Error: connection timeout
```

**解决方案：**
```go
// 检查网络连接
ping etcd-server-ip

// 检查ETCD服务状态
etcdctl endpoint health

// 调整超时配置
grpcConfig := config.GRPCClientConfig{
    ServiceKey: "service.rpc",
    Timeout:    10000, // 增加到10秒
}
```

#### 2. 服务发现失败

**症状：**
```
Error: no available servers
Error: service not found in etcd
```

**解决方案：**
```bash
# 检查ETCD中的服务注册
etcdctl get --prefix service.rpc

# 检查服务key是否正确
# 确认目标服务是否已注册到ETCD
```

#### 3. 重连失败

**症状：**
```
WARN: gRPC连接不健康，状态: TRANSIENT_FAILURE
INFO: 第 5/5 次重连尝试
ERROR: 重连失败，已达到最大重试次数
```

**解决方案：**
```go
// 检查目标服务状态
// 调整重连配置
reconnectConfig := ReconnectConfig{
    MaxRetries:     10,           // 增加重连次数
    RetryInterval:  15 * time.Second, // 增加重连间隔
    HealthInterval: 60 * time.Second, // 增加健康检查间隔
}
```

#### 4. 内存泄漏

**症状：**
```
内存使用持续增长
Too many open files
```

**解决方案：**
```go
// 确保正确关闭连接
defer connManager.Close()

// 检查是否有协程泄漏
go func() {
    // 定期输出协程数量
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            log.Printf("当前协程数量: %d", runtime.NumGoroutine())
        }
    }
}()
```

### 调试技巧

#### 1. 启用调试日志

```go
// 在初始化时启用调试级别日志
logger.SetLevel(logger.DebugLevel)
```

#### 2. 连接状态监控

```go
func debugConnectionState(connManager *grpc_client.ConnectionManager) {
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                log.Printf("连接状态: healthy=%v, state=%v, reconnecting=%v, attempts=%d",
                    connManager.IsHealthy(),
                    connManager.GetConnectionState(),
                    connManager.IsReconnecting(),
                    connManager.GetReconnectAttempts(),
                )
            }
        }
    }()
}
```

#### 3. 网络诊断

```bash
# 检查网络连通性
telnet target-service-ip target-service-port

# 检查DNS解析
nslookup service-name

# 检查防火墙规则
iptables -L

# 检查端口占用
netstat -tlnp | grep :port
```

## 👨‍💻 开发指南 {#development}

### 添加新的服务客户端

#### 1. 定义接口（port包）

```go
// shared/common/grpc/client/port/new_service_client.go
package client

import (
    "context"
    proto "your-project/proto/new_service"
)

type NewServiceClient interface {
    GetSomething(ctx context.Context, req *proto.GetSomethingRequest) (*proto.GetSomethingResponse, error)
    // 添加其他方法...
}
```

#### 2. 实现客户端

```go
// shared/common/grpc/client/new_service_client.go
package grpc_client

import (
    "context"
    "fmt"
    "sync"
    
    client "your-project/shared/common/grpc/client/port"
    proto "your-project/proto/new_service"
    
    "github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

type NewServiceClient struct {
    connectionManager *ConnectionManager
    protoClient       proto.NewServiceClient
    initOnce          func()
}

func NewNewServiceClient(connManager *ConnectionManager) client.NewServiceClient {
    client := &NewServiceClient{
        connectionManager: connManager,
    }
    
    var once sync.Once
    client.initOnce = func() {
        once.Do(func() {
            conn := connManager.GetConnection()
            client.protoClient = proto.NewNewServiceClient(conn)
            logger.Info("新服务客户端初始化完成")
        })
    }
    
    return client
}

func (c *NewServiceClient) GetSomething(ctx context.Context, req *proto.GetSomethingRequest) (*proto.GetSomethingResponse, error) {
    c.initOnce()
    
    logger.Debug(fmt.Sprintf("调用GetSomething，参数: %+v", req))
    
    var resp *proto.GetSomethingResponse
    err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
        var err error
        resp, err = c.protoClient.GetSomething(ctx, req)
        return err
    })
    
    if err != nil {
        logger.Error(fmt.Sprintf("GetSomething失败: %v", err))
        return nil, fmt.Errorf("查询失败: %w", err)
    }
    
    logger.Debug(fmt.Sprintf("GetSomething成功，响应: %+v", resp))
    return resp, nil
}
```

#### 3. 编写测试

```go
// shared/common/grpc/client/new_service_client_test.go
package grpc_client_test

import (
    "context"
    "testing"
    "time"
    
    grpc_client "your-project/shared/common/grpc/client"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestNewServiceClient_GetSomething(t *testing.T) {
    // 创建模拟连接管理器
    mockConnManager := &MockConnectionManager{}
    
    // 创建客户端
    client := grpc_client.NewNewServiceClient(mockConnManager)
    
    // 执行测试
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    req := &proto.GetSomethingRequest{
        Id: 123,
    }
    
    resp, err := client.GetSomething(ctx, req)
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, int32(123), resp.Id)
}
```

### 代码规范

#### 1. 命名规范

- **接口名称**：以 `Client` 结尾，如 `UserInfoServiceClient`
- **实现名称**：与接口同名，如 `UserInfoServiceClient`
- **构造函数**：以 `New` 开头，如 `NewUserInfoServiceClient`

#### 2. 错误处理

```go
// ✅ 推荐的错误处理模式
func (c *ServiceClient) DoSomething(ctx context.Context, req *proto.Request) (*proto.Response, error) {
    c.initOnce()
    
    logger.Debug(fmt.Sprintf("调用DoSomething，参数: %+v", req))
    
    var resp *proto.Response
    err := c.connectionManager.ExecuteWithNetworkRetry(ctx, func() error {
        var err error
        resp, err = c.protoClient.DoSomething(ctx, req)
        return err
    })
    
    if err != nil {
        logger.Error(fmt.Sprintf("DoSomething失败: %v", err))
        return nil, fmt.Errorf("操作失败: %w", err)
    }
    
    if resp == nil {
        logger.Warn("DoSomething返回空结果")
        return nil, fmt.Errorf("响应结果为空")
    }
    
    logger.Debug(fmt.Sprintf("DoSomething成功，响应: %+v", resp))
    return resp, nil
}
```

#### 3. 日志规范

- **DEBUG**：记录详细的请求和响应信息
- **INFO**：记录重要的操作和状态变更
- **WARN**：记录警告信息，如空结果
- **ERROR**：记录错误信息

```go
// 日志格式示例
logger.Debug(fmt.Sprintf("调用%s，参数: %+v", methodName, request))
logger.Info(fmt.Sprintf("%s客户端初始化完成", serviceName))
logger.Warn(fmt.Sprintf("%s返回空结果，参数: %+v", methodName, request))
logger.Error(fmt.Sprintf("%s失败: %v", methodName, err))
```

---

## 📞 支持与反馈

如果你在使用过程中遇到问题或有改进建议，请：

1. 查看本文档的[故障排查](#troubleshooting)部分
2. 检查日志输出以获取详细错误信息
3. 提交 Issue 或联系开发团队

## 🔧 依赖注入 {#dependency-injection}

### 分离架构的依赖注入配置

本项目采用模块化的依赖注入设计，每个组件都有独立的注册函数。

#### 1. ConnectionManager 独立注册

```go
// connection_manager.go
func init() {
    registrations = append(registrations, registerConnectionManagerDependencies)
}

func registerConnectionManagerDependencies() {
    // 注册连接管理器 - 单例，所有服务适配器共享
    if err := di.Provide(func() (*ConnectionManager, error) {
        // 使用SDK中的GRPCClientConfig配置
        clientConfig := config.GrpcConfig.Client

        logger.Infof("从配置中获取gRPC客户端配置: ServiceKey=%s, Timeout=%d",
            clientConfig.ServiceKey, clientConfig.Timeout)

        // 创建连接管理器
        connManager, err := NewConnectionManager(clientConfig, config.EtcdConfig)
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
```

#### 2. 服务客户端独立注册

```go
// userinfo_service_client.go
func init() {
    registrations = append(registrations, registerUserInfoServiceClientDependencies)
}

func registerUserInfoServiceClientDependencies() {
    // 注册用户服务适配器 - 依赖已注册的ConnectionManager
    if err := di.Provide(func(connManager *ConnectionManager) client.UserInfoServiceClient {
        logger.Info("创建用户服务适配器")
        return NewUserInfoServiceClient(connManager)
    }); err != nil {
        logger.Fatalf("注册用户服务客户端失败: %v", err)
    }
}

// orginfo_service_client.go
func init() {
    registrations = append(registrations, registerOrgInfoServiceClientDependencies)
}

func registerOrgInfoServiceClientDependencies() {
    if err := di.Provide(func(connManager *ConnectionManager) client.OrgInfoServiceClient {
        logger.Info("创建组织服务适配器")
        return NewOrgInfoServiceClient(connManager)
    }); err != nil {
        logger.Fatalf("注册组织服务客户端失败: %v", err)
    }
}

// bwc_info_service_client.go
func init() {
    registrations = append(registrations, registerBWCInfoServiceClientDependencies)
}

func registerBWCInfoServiceClientDependencies() {
    if err := di.Provide(func(connManager *ConnectionManager) client.BWCInfoServiceClient {
        logger.Info("创建执法记录仪服务适配器")
        return NewBWCInfoServiceClient(connManager)
    }); err != nil {
        logger.Fatalf("注册执法记录仪服务客户端失败: %v", err)
    }
}
```

#### 3. 统一注册机制

```go
// dependencies.go
var (
    registrations = make([]func(), 0)
)

func RegisterDependencies() {
    // 遍历所有的依赖注入方法
    for _, f := range registrations {
        f()
    }
}
```

#### 4. 应用服务中的使用

```go
// 应用服务中注入需要的客户端
type MediaService struct {
    userInfoClient    client.UserInfoServiceClient    // 只注入用户服务
    lawCameraClient   client.BWCInfoServiceClient // 只注入执法记录仪服务
}

type PermissionService struct {
    userInfoClient client.UserInfoServiceClient  // 注入用户服务
    orgInfoClient  client.OrgInfoServiceClient   // 注入组织服务
}

// 依赖注入配置
func init() {
    di.Provide(func(
        userClient client.UserInfoServiceClient,
        lawCameraClient client.BWCInfoServiceClient,
    ) *MediaService {
        return &MediaService{
            userInfoClient:  userClient,
            lawCameraClient: lawCameraClient,
        }
    })

    di.Provide(func(
        userClient client.UserInfoServiceClient,
        orgClient client.OrgInfoServiceClient,
    ) *PermissionService {
        return &PermissionService{
            userInfoClient: userClient,
            orgInfoClient:  orgClient,
        }
    })
}
```

### 扩展新服务的依赖注入

当添加新服务时，只需要创建对应的注册函数：

```go
// new_service_client.go
func init() {
    registrations = append(registrations, registerNewServiceClientDependencies)
}

func registerNewServiceClientDependencies() {
    if err := di.Provide(func(connManager *ConnectionManager) client.NewServiceClient {
        logger.Info("创建新服务适配器")
        return NewNewServiceClient(connManager)
    }); err != nil {
        logger.Fatalf("注册新服务客户端失败: %v", err)
    }
}
```

### 依赖注入的优势

1. **模块化**：每个服务客户端有独立的注册函数
2. **自动装配**：依赖关系自动解析和注入
3. **单例共享**：ConnectionManager 作为单例被所有客户端共享
4. **易于测试**：可以轻松替换依赖进行单元测试
5. **配置集中**：所有配置都从统一的配置源获取

---

## 📞 支持与反馈

如果你在使用过程中遇到问题或有改进建议，请：

1. 查看本文档的[故障排查](#troubleshooting)部分
2. 检查日志输出以获取详细错误信息
3. 提交 Issue 或联系开发团队

---

**最后更新：** 2024-12-15
**版本：** v2.0.0
**架构版本：** 职责分离架构