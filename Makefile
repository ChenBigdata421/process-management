.PHONY: help build run run-worker clean test docker-build docker-up docker-down docker-logs docker-ps

help:
	@echo "Process Management - Makefile Commands"
	@echo "========================================"
	@echo ""
	@echo "Docker Compose 命令:"
	@echo "  make docker-build   - 构建所有 Docker 镜像"
	@echo "  make docker-up      - 启动所有服务"
	@echo "  make docker-down    - 停止所有服务"
	@echo "  make docker-restart - 重启所有服务"
	@echo "  make docker-logs    - 查看所有服务日志"
	@echo "  make docker-ps      - 查看容器状态"
	@echo "  make docker-clean   - 停止并删除所有容器和卷"
	@echo ""
	@echo "本地开发命令:"
	@echo "  make build          - 构建本地二进制文件"
	@echo "  make run            - 本地运行 API 服务"
	@echo "  make run-worker     - 本地运行 Worker 服务"
	@echo "  make clean          - 清理构建文件"
	@echo "  make test           - 运行测试"
	@echo ""

# Docker Compose 命令
docker-network:
	@echo "创建共享网络..."
	docker network create security-management_jxt-web 2>/dev/null || echo "网络已存在"

docker-build: docker-network
	@echo "构建 Docker 镜像..."
	docker-compose build

docker-up: docker-network
	@echo "启动所有服务..."
	docker-compose up -d

docker-down:
	@echo "停止所有服务..."
	docker-compose down

docker-restart:
	@echo "重启所有服务..."
	docker-compose restart

docker-logs:
	docker-compose logs -f

docker-ps:
	docker-compose ps

docker-clean:
	@echo "清理所有容器和卷..."
	docker-compose down -v
	rm -f workflows.db

# 本地开发命令
build:
	@echo "构建本地二进制文件..."
	go build -o bin/api cmd/main.go
	go build -o bin/worker cmd/worker/main.go

run:
	@echo "运行 API 服务..."
	go run cmd/main.go

run-worker:
	@echo "运行 Worker 服务..."
	go run cmd/worker/main.go

clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f workflows.db

test:
	@echo "运行测试..."
	go test -v ./...

