# Dead gRPC Code Cleanup

**Date:** 2026-05-27
**Status:** Approved

## Problem

The process-management project defines 19 gRPC client methods across 6 service clients, but only `GrpcCasbinPolicyProvider.GetPolicies` is actually invoked at runtime. The remaining 18 methods across 5 service clients are dead code — implemented but never wired into DI or called by any business logic.

Additionally, the Casbin provider itself has a DI bug: `di.Invoke` is called for it in `cmd/api/server.go:103` but it was never registered with `di.Provide`, so it would fail at runtime.

## Decision

Delete all dead gRPC client implementations, service wrappers, and the internal gRPC layer. Keep proto files and ConnectionManager infrastructure for future use. Fix the Casbin DI wiring.

## Files to Delete

### Dead client implementations (6 files)
- `shared/infrastructure/grpc/client/userinfo_service_client.go`
- `shared/infrastructure/grpc/client/orginfo_service_client.go`
- `shared/infrastructure/grpc/client/bwc_info_service_client.go`
- `shared/infrastructure/grpc/client/enforcement_type_service_client.go`
- `shared/infrastructure/grpc/client/storagesiteinfo_service_client.go`
- `shared/infrastructure/grpc/client/auth_service_client.go`

### Dead port interfaces (5 files)
- `shared/infrastructure/grpc/client/port/userinfo_service_client.go`
- `shared/infrastructure/grpc/client/port/orginfo_service_client.go`
- `shared/infrastructure/grpc/client/port/bwc_info_service_client.go`
- `shared/infrastructure/grpc/client/port/enforcement_type_service_client.go`
- `shared/infrastructure/grpc/client/port/storagesiteinfo_service_client.go`

### Dead service wrappers (5 files)
- `shared/infrastructure/service/user_info_service.go`
- `shared/infrastructure/service/organization_info_service.go`
- `shared/infrastructure/service/bwc_info_service.go`
- `shared/infrastructure/service/enforcement_type_info_service.go`
- `shared/infrastructure/service/storagesite_info_service.go`

### Dead internal gRPC layer (8 files)
- `internal/infrastructure/grpc/user_client.go`
- `internal/infrastructure/grpc/proto/user.proto`
- `internal/infrastructure/grpc/proto/user.pb.go`
- `internal/infrastructure/grpc/proto/user_grpc.pb.go`
- `internal/infrastructure/grpc/proto/org.proto`
- `internal/infrastructure/grpc/proto/org.pb.go`
- `internal/infrastructure/grpc/proto/org_grpc.pb.go`
- `internal/infrastructure/grpc/proto/generate.ps1`

### Dead documentation
- `shared/infrastructure/grpc/client/README.md`

## Files to Keep

- `shared/infrastructure/grpc/client/connection_manager.go` — shared connection management
- `shared/infrastructure/grpc/client/casbin_policy_provider.go` — only working client
- `shared/infrastructure/grpc/proto/casbin/` — in use
- `shared/infrastructure/grpc/proto/user/` — for future use
- `shared/infrastructure/grpc/proto/org/` — for future use
- `shared/infrastructure/grpc/proto/bwc/` — for future use
- `shared/infrastructure/grpc/proto/enforcement_type/` — for future use
- `shared/infrastructure/grpc/proto/storagesite/` — for future use

## DI Fix

Create `shared/infrastructure/grpc/client/dependencies.go`:
- `di.Provide` for `*ConnectionManager` using `config.GrpcConfig.Client` + `config.EtcdConfig`
- `di.Provide` for `*GrpcCasbinPolicyProvider` depending on `*ConnectionManager`

Register in `cmd/api/hexagon.go`.

## Cleanup

Remove empty directories after deletion:
- `internal/infrastructure/grpc/` if empty
- `shared/infrastructure/service/` if empty
- `shared/infrastructure/grpc/client/port/` if empty
