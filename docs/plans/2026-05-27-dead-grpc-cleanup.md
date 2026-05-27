# Dead gRPC Code Cleanup — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove all dead gRPC client code and fix the Casbin DI wiring so the only runtime gRPC call actually works.

**Architecture:** Delete ~30 dead files across 3 layers (internal gRPC, shared service wrappers, shared gRPC clients). Add a `dependencies.go` aligned with evidence-management/file-storage-service pattern (`sync.Once` + separate register functions). Register in `cmd/api/hexagon.go`.

**Tech Stack:** Go, dig (DI), go-zero/zrpc, gRPC

---

### Task 1: Delete dead internal gRPC layer

**Files:**
- Delete: `internal/infrastructure/grpc/user_client.go`
- Delete: `internal/infrastructure/grpc/proto/user.proto`
- Delete: `internal/infrastructure/grpc/proto/user.pb.go`
- Delete: `internal/infrastructure/grpc/proto/user_grpc.pb.go`
- Delete: `internal/infrastructure/grpc/proto/org.proto`
- Delete: `internal/infrastructure/grpc/proto/org.pb.go`
- Delete: `internal/infrastructure/grpc/proto/org_grpc.pb.go`
- Delete: `internal/infrastructure/grpc/proto/generate.ps1`

**Step 1: Delete all files**

```bash
cd D:/JXT/jxt-evidence-system/process-management
rm -rf internal/infrastructure/grpc/
```

**Step 2: Verify build**

```bash
go build ./...
```

Expected: compiles with no errors. No code imports this package.

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: remove dead internal gRPC client layer"
```

---

### Task 2: Delete dead shared service wrappers

**Files:**
- Delete: `shared/infrastructure/service/user_info_service.go`
- Delete: `shared/infrastructure/service/organization_info_service.go`
- Delete: `shared/infrastructure/service/bwc_info_service.go`
- Delete: `shared/infrastructure/service/enforcement_type_info_service.go`
- Delete: `shared/infrastructure/service/storagesite_info_service.go`
- Delete: `shared/infrastructure/service/constants.go`
- Delete: `shared/infrastructure/service/dto/` (entire directory)
- Delete: `shared/infrastructure/service/port/` (entire directory)

Note: The entire `shared/infrastructure/service/` directory is dead — only self-references and references from the dead gRPC clients exist.

**Step 1: Delete the entire service directory**

```bash
rm -rf shared/infrastructure/service/
```

**Step 2: Verify build**

```bash
go build ./...
```

Expected: compiles with no errors.

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: remove dead shared service wrapper layer"
```

---

### Task 3: Delete dead gRPC client implementations and port interfaces

**Files:**
- Delete: `shared/infrastructure/grpc/client/userinfo_service_client.go`
- Delete: `shared/infrastructure/grpc/client/orginfo_service_client.go`
- Delete: `shared/infrastructure/grpc/client/bwc_info_service_client.go`
- Delete: `shared/infrastructure/grpc/client/enforcement_type_service_client.go`
- Delete: `shared/infrastructure/grpc/client/storagesiteinfo_service_client.go`
- Delete: `shared/infrastructure/grpc/client/auth_service_client.go`
- Delete: `shared/infrastructure/grpc/client/README.md`
- Delete: `shared/infrastructure/grpc/client/port/` (entire directory — all 5 port interfaces)

**Step 1: Delete files**

```bash
rm shared/infrastructure/grpc/client/userinfo_service_client.go
rm shared/infrastructure/grpc/client/orginfo_service_client.go
rm shared/infrastructure/grpc/client/bwc_info_service_client.go
rm shared/infrastructure/grpc/client/enforcement_type_service_client.go
rm shared/infrastructure/grpc/client/storagesiteinfo_service_client.go
rm shared/infrastructure/grpc/client/auth_service_client.go
rm shared/infrastructure/grpc/client/README.md
rm -rf shared/infrastructure/grpc/client/port/
```

**Step 2: Verify build**

```bash
go build ./...
```

Expected: compiles with no errors.

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: remove dead gRPC client implementations and port interfaces"
```

---

### Task 4: Add DI wiring for ConnectionManager and GrpcCasbinPolicyProvider

**Files:**
- Create: `shared/infrastructure/grpc/client/dependencies.go`
- Modify: `cmd/api/hexagon.go` — add registration

**Step 1: Create `shared/infrastructure/grpc/client/dependencies.go`**

对齐 file-storage-service 的模式：`sync.Once` 防重复 + 独立 register 函数。

```go
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
```

**Step 2: Modify `cmd/api/hexagon.go` — add import and registration**

Add the import:
```go
grpc_client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client"
```

Add registration in `init()`:
```go
Registrations = append(Registrations, grpc_client.RegisterDependencies)
```

The full `init()` should become:
```go
func init() {
	AppRouters = append(AppRouters, router.InitRouter)

	Registrations = append(Registrations, infra_ws.RegisterDependencies)
	Registrations = append(Registrations, persistence.RegisterDependencies)
	Registrations = append(Registrations, domain_service.RegisterDependencies)
	Registrations = append(Registrations, application.RegisterDependencies)
	Registrations = append(Registrations, api.RegisterDependencies)
	Registrations = append(Registrations, outbox.RegisterDependencies)
	Registrations = append(Registrations, grpc_client.RegisterDependencies)
}
```

**Step 3: Verify build**

```bash
go build ./...
```

Expected: compiles with no errors.

**Step 4: Commit**

```bash
git add shared/infrastructure/grpc/client/dependencies.go cmd/api/hexagon.go
git commit -m "feat: add DI wiring for ConnectionManager and GrpcCasbinPolicyProvider"
```

---

### Task 5: Verify and clean up

**Step 1: Full build check**

```bash
go build ./...
```

**Step 2: Verify no broken imports**

```bash
go vet ./...
```

**Step 3: Check that remaining files make sense**

The `shared/infrastructure/grpc/client/` directory should now contain:
- `connection_manager.go`
- `casbin_policy_provider.go`
- `dependencies.go`

The `shared/infrastructure/grpc/proto/` directory should contain:
- `casbin/` (in use)
- `user/` (for future use)
- `org/` (for future use)
- `bwc/` (for future use)
- `enforcement_type/` (for future use)
- `storagesite/` (for future use)

**Step 4: Final commit if any remaining cleanup needed**

```bash
git add -A
git commit -m "chore: final cleanup after dead gRPC code removal"
```
