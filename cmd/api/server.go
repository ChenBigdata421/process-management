package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/eventbus"
	logger "github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	ws "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	localoutbox "jxt-evidence-system/process-management/internal/infrastructure/outbox"
	"jxt-evidence-system/process-management/shared/common/database"
	"jxt-evidence-system/process-management/shared/common/di"
	"jxt-evidence-system/process-management/shared/common/global"
	common "jxt-evidence-system/process-management/shared/common/middleware"
	"jxt-evidence-system/process-management/shared/common/middleware/handler"
)

var (
	configYml string
	apiCheck  bool
	StartCmd  = &cobra.Command{
		Use:          "server",
		Short:        "Start API server",
		Example:      "go-admin server -c config/settings.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

// 用来记录用户自定义路由，路由属于api，和application解耦是对的
var (
	AppRouters    = make([]func(), 0)
	Registrations = make([]func(), 0) // jiyuanjie添加：记录需要的依赖注入
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/settings.yml", "Start server with provided configuration file")
	StartCmd.PersistentFlags().BoolVarP(&apiCheck, "api", "a", false, "Start server with check api data")

	//注册路由 fixme 其他应用的路由，在本目录新建文件放在init方法
	//AppRouters = append(AppRouters, router.InitRouter)// 这里添加app/admin的router
	// 注意：router.InitRouter 已在 hexagon.go 中注册，不要在这里重复注册
}

func setup() {

	// 检查配置文件路径是否为空
	if configYml == "" {
		log.Fatal("配置文件路径不能为空，请使用-c参数指定配置文件路径")
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configYml); os.IsNotExist(err) {
		log.Fatalf("配置文件 %s 不存在", configYml)
	}

	// 读取配置，并写入配置结构体
	config.Setup(configYml)

	// 如果是开发模式，打印配置
	if config.ApplicationConfig.Mode == pkg.ModeDev.String() {
		printConfig()
	}

	// 初始化基础组件
	logger.Setup()
	database.ProcessDbSetup()

	// ⭐ 初始化 EventBus（支持 NATS/Kafka 切换）
	if err := setupEventBus(); err != nil {
		log.Fatalf("Failed to setup EventBus: %v", err)
	}

	usageStr := `starting process management command api server...`
	log.Println(usageStr)
}

func printConfig() {

	/*fmt：固定输出到标准输出(stdout)
	log：默认输出到标准错误(stderr)，可通过SetOutput()重定向
	fmt：原样输出内容
	log：自动添加时间前缀 2009/01/23 01:23:23 message
	fmt：非线程安全
	log：内部有锁机制保证线程安全
	临时调试用 fmt
	正式日志记录用 log
	需要结构化日志时推荐使用 zap/logrus 等专业日志库*/

	applicationConfig, errs := json.MarshalIndent(config.ApplicationConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("application:", string(applicationConfig))

	loggerConfig, errs := json.MarshalIndent(config.LoggerConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("logger:", string(loggerConfig))

	httpConfig, errs := json.MarshalIndent(config.HttpConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("http:", string(httpConfig))

	etcdConfig, errs := json.MarshalIndent(config.EtcdConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("etcd:", string(etcdConfig))

	grpcConfig, errs := json.MarshalIndent(config.GrpcConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("grpc:", string(grpcConfig))

	jwtConfig, errs := json.MarshalIndent(config.JwtConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("jwt:", string(jwtConfig))

	// todo 需要兼容
	databaseConfig, errs := json.MarshalIndent(config.DatabaseConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("database:", string(databaseConfig))

	queueConfig, errs := json.MarshalIndent(config.QueueConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("queue:", string(queueConfig))

	tenantConfig, errs := json.MarshalIndent(config.TenantsConfig, "", "   ") //转换成JSON返回的是byte[]
	if errs != nil {
		log.Println(errs.Error())
	}
	log.Println("tenant:", string(tenantConfig))

}

func run() error {

	if config.ApplicationConfig.Mode == pkg.ModeProd.String() {
		gin.SetMode(gin.DebugMode) // 调试阶段改为debugMode
	}

	// 确保程序退出前刷新日志
	defer logger.Logger.Sync()

	// jiyuanjie添加：初始化gin路由之前，先完成repo，service，api的依赖注入
	for _, f := range Registrations {
		f()
	}

	// 启动outbox调度器
	if err := startOutboxScheduler(); err != nil {
		log.Printf("启动outbox调度器失败: %v", err)
		return err
	}

	// 同时启动HTTP和gRPC服务
	errChan := make(chan error, 2)
	sigChan := make(chan os.Signal, 1)
	stopChan := make(chan struct{})
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动HTTP服务(如果启用)
	if config.HttpConfig.Enabled {
		go func() {
			errChan <- startHTTPServer()
		}()
	}

	// 等待信号或错误
	select {
	case err := <-errChan:
		log.Printf("服务错误: %v", err)
	case sig := <-sigChan:
		log.Printf("接收到信号: %v", sig)
	}

	// 发送停止信号并等待服务关闭
	close(stopChan)
	// 优雅关闭数据库等
	if err := gracefulShutdown(); err != nil {
		log.Printf("Error during graceful shutdown: %v\n", err)
	}
	time.Sleep(time.Second) // 给服务一些时间来完成关闭
	log.Println("服务已优雅退出")
	return nil

}

func startHTTPServer() error {
	initRouter()

	for _, f := range AppRouters {
		f()
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.HttpConfig.Host, config.HttpConfig.Port),
		Handler: sdk.Runtime.GetEngine(),
	}

	if apiCheck {
		performAPICheck()
	}

	go func() {
		// 服务连接
		if config.HttpConfig.SSL.Enabled {
			if err := srv.ListenAndServeTLS(config.HttpConfig.SSL.Pem, config.HttpConfig.SSL.KeyStr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("listen: ", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("listen: ", err)
			}
		}
	}()
	log.Println(pkg.Red(string(global.LogoContent)))
	tip()
	fmt.Println(pkg.Blue(string(global.JXTLogoContent)))
	JXTTip()

	log.Println(pkg.Green("HTTPServer run at:"))
	log.Printf("-  Local:   %s://localhost:%d/ \r\n", "http", config.HttpConfig.Port)
	log.Printf("-  Network: %s://%s:%d/ \r\n", "http", pkg.GetLocaHonst(), config.HttpConfig.Port)
	log.Printf("%s Enter Control + C Shutdown HTTPServer \r\n", pkg.GetCurrentTimeStr())

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("Shutdown HTTPServer ... ")

	// 关闭 WebSocket Hub
	var wsNotifier ws.WebSocketNotifier
	di.Invoke(func(notifier ws.WebSocketNotifier) {
		wsNotifier = notifier
	})
	if wsNotifier != nil {
		if err := wsNotifier.Close(); err != nil {
			log.Printf("Error closing WebSocket Hub: %v\n", err)
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("HTTPServer Shutdown:", err)
	}
	log.Println("HTTPServer shutdown completed")

	return nil
}

//var Router runtime.Router

func tip() {
	usageStr := `欢迎使用 ` + pkg.Green(`go-admin `+global.Version) + ` 可以使用 ` + pkg.Red(`-h`) + ` 查看命令`
	fmt.Printf("%s \n\n", usageStr)
}

func JXTTip() {
	usageStr := `欢迎使用 ` + pkg.Green(`JXT证据管理系统`) + ` 可以使用 ` + pkg.Red(`-h`) + ` 查看命令`
	fmt.Printf("%s \n\n", usageStr)
}

func initRouter() {
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		h = gin.New()
		sdk.Runtime.SetEngine(h)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
		//os.Exit(-1)
	}
	if config.SslConfig.Enable {
		r.Use(handler.TlsHandler())
	}
	//r.Use(middleware.Metrics())
	r.Use(common.Sentinel()).
		Use(logger.SetRequestLogger) //jiyuanjie 创建基于基础zapLogger的requestLogger

	common.InitMiddleware(r)

	// 注册健康检查端点（无需认证）
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "ok",
			"data": gin.H{
				"status": "healthy",
			},
		})
	})

}

func performAPICheck() {
	var routers = sdk.Runtime.GetRouter()
	q := sdk.Runtime.GetMemoryQueue("")
	mp := make(map[string]interface{})
	mp["List"] = routers
	message, err := sdk.Runtime.GetStreamMessage("", global.ApiCheck, mp)
	if err != nil {
		log.Printf("GetStreamMessage error, %s \n", err.Error())
		//日志报错错误，不中断请求
	} else {
		err = q.Append(message)
		if err != nil {
			log.Printf("Append message error, %s \n", err.Error())
		}
	}
}

func gracefulShutdown() error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// ⭐ 关闭 EventBus
	eventBus := sdk.Runtime.GetEventBus()
	if eventBus != nil {
		if err := eventBus.Close(); err != nil {
			log.Printf("EventBus Shutdown: %v\n", err)
		}
	}

	// 关闭 Database
	if err := database.Close(shutdownCtx); err != nil {
		log.Printf("Error closing database: %v\n", err)
	}

	return nil
}

// setupEventBus 初始化 EventBus（支持 NATS/Kafka 切换）
func setupEventBus() error {
	// 从配置文件加载 EventBus 配置
	eventBusConfig := config.AppConfig.EventBus
	if eventBusConfig == nil {
		logger.Warn("eventbus config not found in settings.yml, EventBus will not be initialized")
		return nil
	}

	// 初始化全局 EventBus
	if err := eventbus.InitializeFromConfig(eventBusConfig); err != nil {
		return fmt.Errorf("failed to initialize EventBus: %w", err)
	}

	// 设置到 SDK Runtime
	bus := eventbus.GetGlobal()
	sdk.Runtime.SetEventBus(bus)

	// ✅ Kafka 预订阅优化（避免 Consumer Group 重平衡）
	if eventBusConfig.Type == "kafka" {
		topics := []string{
			// Evidence 自身的 Topics
			"process.task.events",
		}
		//尝试将 bus 断言为具有 SetPreSubscriptionTopics 方法的接口
		if kafkaBus, ok := bus.(interface{ SetPreSubscriptionTopics([]string) }); ok {
			kafkaBus.SetPreSubscriptionTopics(topics)
			logger.Info("✅ Kafka 预订阅 topic 已设置", "count", len(topics))
		}
	}

	logger.Info("EventBus initialized successfully", "type", eventBusConfig.Type)

	// ========== ✅ Topic 预创建优化 ==========
	if err := preCreateTopics(bus); err != nil {
		// 预创建失败不影响启动，只记录警告
		logger.Warn("Topic pre-creation failed", "error", err)
	}

	logger.Info("EventBus initialized successfully", "type", eventBusConfig.Type)
	return nil
}

// startOutboxScheduler 启动 jxt-core outbox 事件调度器
// 根据多租户配置开关，创建单个（非多租户）或多个 Scheduler（每个租户一个）
func startOutboxScheduler() error {
	return di.Invoke(func(manager *localoutbox.OutboxSchedulerManager) {
		if manager == nil {
			log.Fatal("outbox scheduler manager is nil")
		}

		// 启动调度器管理器（会根据配置自动创建单个或多个 Scheduler）
		ctx := context.Background()
		if err := manager.Start(ctx); err != nil {
			log.Fatalf("Failed to start outbox scheduler manager: %v", err)
		}

		logger.Infof("Outbox scheduler manager started with %d scheduler(s)", manager.GetSchedulerCount())
	})
}

// preCreateTopics 预创建所有 Kafka Topics（可选，Kafka 支持自动创建）
//
// 优化原理：
// - 在应用启动时预先创建所有 Topic
// - 避免运行时自动创建导致的延迟
// - 可以预设 Topic 配置（分区数、压缩算法等）
//
// 参考文档：docs/消息中间件切换到kafka/02-设计方案.md
func preCreateTopics(bus eventbus.EventBus) error {
	ctx := context.Background()

	// 1. 定义所有需要使用的 Topic（与 TopicMapper 保持一致）
	topics := []string{
		// Evidence 自身的 Topics
		"process.task.events", // Media 聚合事件
	}

	logger.Infof("🚀 开始预创建 %d 个 Kafka Topics...", len(topics))
	startTime := time.Now()

	// 2. 预创建所有 Topics
	successCount := 0
	for _, topic := range topics {
		err := bus.ConfigureTopic(ctx, topic, eventbus.TopicOptions{
			PersistenceMode: eventbus.TopicPersistent, // 持久化模式
			RetentionTime:   168 * time.Hour,          // 7 天
			MaxSize:         1073741824,               // 1GB
			Partitions:      1,                        // 单分区保证顺序
			Compression:     "snappy",                 // Snappy 压缩
			Description:     fmt.Sprintf("Evidence management topic for %s", topic),
		})
		if err != nil {
			logger.Warnf("⚠️  预创建 Topic 失败: %s - %v", topic, err)
		} else {
			logger.Infof("  ✅ Topic 预创建成功: %s", topic)
			successCount++
		}
	}

	elapsed := time.Since(startTime)
	logger.Infof("✅ Topic 预创建完成！成功: %d/%d, 耗时: %v", successCount, len(topics), elapsed)

	if successCount == 0 {
		return fmt.Errorf("所有 Topic 预创建失败")
	}

	return nil
}
