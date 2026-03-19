# 生成 gRPC 代码
# 需要先安装 protoc 和 protoc-gen-go, protoc-gen-go-grpc

Write-Host "Generating gRPC code from proto files..." -ForegroundColor Green

# 生成用户服务代码
protoc --go_out=. --go_opt=paths=source_relative `
       --go-grpc_out=. --go-grpc_opt=paths=source_relative `
       user.proto

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ user.proto generated successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to generate user.proto" -ForegroundColor Red
    exit 1
}

# 生成组织服务代码
protoc --go_out=. --go_opt=paths=source_relative `
       --go-grpc_out=. --go-grpc_opt=paths=source_relative `
       org.proto

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ org.proto generated successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to generate org.proto" -ForegroundColor Red
    exit 1
}

Write-Host "`nAll proto files generated successfully!" -ForegroundColor Green

