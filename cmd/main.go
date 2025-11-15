package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/gin-gonic/gin"
	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/cmd/migration"
	_ "github.com/jxt/process-management/cmd/migration/version"
	"github.com/jxt/process-management/config"
	"github.com/jxt/process-management/domain/workflow"
	"github.com/jxt/process-management/infrastructure/database"
	"github.com/jxt/process-management/infrastructure/persistence"
	"github.com/jxt/process-management/infrastructure/websocket"
	"github.com/jxt/process-management/interfaces/http/handler"
	"github.com/jxt/process-management/interfaces/http/router"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	log.Printf("Config loaded: %v", cfg)

	// 创建数据库
	// 从 DATABASE_URL 中提取主机信息
	adminDSN := os.Getenv("ADMIN_DATABASE_URL")
	if adminDSN == "" {
		// 默认使用本地连接
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

	db := dbConn.GetDB()

	log.Println("数据库迁移开始")
	if err := db.Debug().AutoMigrate(&workflow.Migration{}); err != nil {
		log.Println(pkg.Red("数据库迁移失败: %v\n"), err)
	}

	migration.Migrate.SetDb(db.Debug())
	migration.Migrate.Migrate()
	log.Println(`数据库基础数据初始化成功`)

	// 初始化仓储
	workflowRepo := persistence.NewWorkflowRepository(db)
	instanceRepo := persistence.NewWorkflowInstanceRepository(db)
	taskRepo := persistence.NewTaskRepository(db)
	taskHistoryRepo := persistence.NewTaskHistoryRepository(db)

	// 🆕 初始化WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	log.Println("WebSocket Hub started")

	// 🆕 初始化通知服务
	notificationService := workflow.NewNotificationService(wsHub)
	log.Println("Notification service initialized")

	// 🆕 初始化工作流引擎服务
	engineService := workflow.NewWorkflowEngineService(workflowRepo, instanceRepo, taskRepo)
	engineService.SetNotificationService(notificationService)
	log.Println("Workflow engine service initialized")

	// 初始化命令处理器
	createHandler := command.NewCreateWorkflowHandler(workflowRepo)
	updateHandler := command.NewUpdateWorkflowHandler(workflowRepo)
	deleteHandler := command.NewDeleteWorkflowHandler(workflowRepo)
	activateHandler := command.NewActivateWorkflowHandler(workflowRepo)
	freezeHandler := command.NewFreezeWorkflowHandler(workflowRepo)
	startInstanceHandler := command.NewStartWorkflowInstanceHandler(workflowRepo, instanceRepo, engineService)
	deleteInstanceHandler := command.NewDeleteInstanceHandler(instanceRepo)
	createTaskHandler := command.NewCreateTaskHandler(taskRepo)
	claimTaskHandler := command.NewClaimTaskHandler(taskRepo, taskHistoryRepo)
	completeTaskHandler := command.NewCompleteTaskHandler(taskRepo, taskHistoryRepo, engineService)
	delegateTaskHandler := command.NewDelegateTaskHandler(taskRepo, taskHistoryRepo)
	deleteTaskHandler := command.NewDeleteTaskHandler(taskRepo)

	// 初始化查询服务
	workflowQueryService := query.NewWorkflowQueryService(workflowRepo)
	instanceQueryService := query.NewWorkflowInstanceQueryService(instanceRepo)
	taskQueryService := query.NewTaskQueryService(taskRepo, taskHistoryRepo, workflowRepo)

	// 初始化HTTP处理器
	workflowHandler := handler.NewWorkflowHandler(
		createHandler,
		updateHandler,
		deleteHandler,
		activateHandler,
		freezeHandler,
		workflowQueryService,
		instanceRepo,
	)
	instanceHandler := handler.NewInstanceHandler(
		startInstanceHandler,
		deleteInstanceHandler,
		instanceQueryService,
	)
	taskHandler := handler.NewTaskHandler(
		createTaskHandler,
		claimTaskHandler,
		completeTaskHandler,
		delegateTaskHandler,
		deleteTaskHandler,
		taskQueryService,
	)

	// 🆕 初始化WebSocket处理器
	wsHandler := handler.NewWebSocketHandler(wsHub)
	log.Println("WebSocket handler initialized")

	// 设置Gin引擎
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// 设置路由
	router.SetupRoutes(engine, workflowHandler, instanceHandler, taskHandler)

	// 🆕 添加WebSocket路由
	engine.GET("/ws", wsHandler.HandleWebSocket)
	engine.GET("/api/ws/online-users", wsHandler.GetOnlineUsers)
	engine.GET("/api/ws/user/:user_id/online", wsHandler.CheckUserOnline)
	engine.POST("/api/ws/test-message", wsHandler.SendTestMessage)
	log.Println("WebSocket routes registered")

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    cfg.GetServerPort(),
		Handler: engine,
	}

	go func() {
		log.Printf("Starting server on %s", cfg.GetServerPort())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
