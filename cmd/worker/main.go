package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jxt/process-management/config"
	"github.com/jxt/process-management/infrastructure/database"
	"github.com/jxt/process-management/infrastructure/workflow"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	log.Printf("Config loaded: %v", cfg)

	// 创建数据库
	adminDSN := os.Getenv("ADMIN_DATABASE_URL")
	if adminDSN == "" {
		adminDSN = "postgres://root:123456@127.0.0.1:5432/postgres?sslmode=disable&connect_timeout=1&TimeZone=Asia/Shanghai"
	}
	if err := database.CreateDatabaseIfNotExists(adminDSN, "processdb"); err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	// 连接数据库
	dbConn, err := database.NewPostgresConnection(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// 检查数据库连接
	if err := dbConn.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database connected successfully")

	// 创建工作流执行器
	executor, err := workflow.NewWorkflowExecutor("./workflows.db")
	if err != nil {
		log.Fatalf("Failed to create workflow executor: %v", err)
	}
	defer executor.Close()

	// 注册工作流和活动
	executor.RegisterWorkflows()
	executor.RegisterActivities()

	// 启动执行器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := executor.Start(ctx); err != nil {
			log.Fatalf("Workflow executor error: %v", err)
		}
	}()

	log.Println("Workflow executor started successfully")

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down workflow executor...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := executor.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping executor: %v", err)
	}

	log.Println("Workflow executor stopped")
}
