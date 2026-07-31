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
	infra_eventbus "jxt-evidence-system/process-management/internal/infrastructure/eventbus"
	localoutbox "jxt-evidence-system/process-management/internal/infrastructure/outbox"
	"jxt-evidence-system/process-management/shared/common/database"
	"jxt-evidence-system/process-management/shared/common/di"
	"jxt-evidence-system/process-management/shared/common/global"
	common "jxt-evidence-system/process-management/shared/common/middleware"
	"jxt-evidence-system/process-management/shared/common/middleware/handler"
	tenantdb "jxt-evidence-system/process-management/shared/common/tenantdb"
	grpc_client "jxt-evidence-system/process-management/shared/infrastructure/grpc/client"
)

var (
	configYml   string
	apiCheck    bool
	tenantCache *tenantdb.Cache
	StartCmd    = &cobra.Command{
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
	// 新增：初始化租户数据库缓存（仅 process DB）
	var err error
	tenantCache, err = tenantdb.Setup("process-management")
	if err != nil {
		log.Printf("警告：租户缓存初始化失败: %v", err)
	}

	// 启动租户监听器（设置 Casbin 初始化回调）
	if tenantCache != nil {
		watcherConfig := tenantdb.DefaultWatcherConfig()
		// 设置新租户 Casbin 初始化回调
		if config.GrpcConfig.Client.Enabled {
			watcherConfig.OnTenantAdded = func(tenantID int) error {
				log.Printf("[Casbin] 动态租户 %d 开始初始化", tenantID)
				var setupErr error
				if err := di.Invoke(func(provider *grpc_client.GrpcCasbinPolicyProvider) {
					setupErr = database.SetupTenantCasbin(provider, tenantID)
				}); err != nil {
					log.Printf("[Casbin] 动态租户 %d 初始化失败(DI): %v", tenantID, err)
					return err
				}
				if setupErr != nil {
					log.Printf("[Casbin] 动态租户 %d 初始化失败: %v", tenantID, setupErr)
					return setupErr
				}
				log.Printf("[Casbin] 动态租户 %d 初始化成功", tenantID)
				return nil
			}
		}

		if err := tenantdb.StartWatcher(tenantCache, watcherConfig); err != nil {
			log.Printf("警告：动态租户监听启动失败: %v", err)
		}
	}
	database.ProcessDbSetup() // 初始化process数据库(存放流程信息)

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

	// 注册健康检查端点（无需认证、无需租户解析）。
	// 必须在 common.InitMiddleware（含域名租户回退中间件）之前注册：
	// Gin 路由只会绑定其注册之前通过 r.Use 挂载的中间件；若放在后面，
	// /api/health 会被 host 解析的租户回退中间件拦截（域名无法解析即返回
	// 400 "tenant ID missing"），导致 Docker healthcheck 探针失败、容器被判为 unhealthy。
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "ok",
			"data": gin.H{
				"status": "healthy",
			},
		})
	})

	if config.SslConfig.Enable {
		r.Use(handler.TlsHandler())
	}
	//r.Use(middleware.Metrics())
	r.Use(common.Sentinel()).
		Use(logger.SetRequestLogger) //jiyuanjie 创建基于基础zapLogger的requestLogger

	common.InitMiddleware(r, tenantCache)

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

	// 关闭 gRPC 连接
	_ = di.Invoke(func(cm *grpc_client.ConnectionManager) {
		if err := cm.Close(); err != nil {
			log.Printf("Error closing gRPC connection: %v\n", err)
		}
	})

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

	// ========== v4 动作1:注册表驱动 ==========
	reg := buildEventRegistry()

	// ✅ Kafka 预订阅优化（避免 Consumer Group 重平衡）
	if eventBusConfig.Type == "kafka" {
		//尝试将 bus 断言为具有 SetPreSubscriptionTopics 方法的接口
		if kafkaBus, ok := bus.(interface{ SetPreSubscriptionTopics([]string) }); ok {
			kafkaBus.SetPreSubscriptionTopics(reg.Topics())
			logger.Info("✅ Kafka 预订阅 topic 已设置", "count", len(reg.Topics()))
		}
	}

	logger.Info("EventBus initialized successfully", "type", eventBusConfig.Type)

	// ========== ✅ 就绪门禁 + §七 存在性断言（topic 配置由 infra bootstrap 落地，v2 §十七） ==========
	if err := waitTopologyReady(bus, reg.Topics()); err != nil {
		// fail-fast：topology 未就绪即拒启动——由 restart: on-failure 重试
		return fmt.Errorf("waitTopologyReady failed: %w", err)
	}

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

// waitTopologyReady 在订阅之前【单次】检查 redpanda topology 就绪哨兵 jxt.topology.ready，
// 通过后对 owned topic 做 §七 存在性断言。topic 存在性/分区数/留存期/压缩配置全部由 infra
// bootstrap 收敛（docs/analysis/redpanda主题创建优化方案_v2.md 改动 4/6 + §十七）：标志存在 ⟹
// 最近一次 bootstrap exit 0 ⟹ 全部主题已建好并配置 → 本服务不再断言分区数、不再 ConfigureTopic。
// 标志缺失即返回错误 → setupEventBus → log.Fatalf → restart: on-failure 重试。
func waitTopologyReady(bus eventbus.EventBus, topics []string) error {
	ctx := context.Background()

	// 【就绪门禁 / 形态乙】检查 jxt.topology.ready 哨兵：标志存在 ⟹ 最近一次 bootstrap 成功
	// ⟹ 全部主题已建好并配置。(E3) 走 metadata 只读查询，禁止 produce/consume 触发式探测。
	if err := eventbus.WaitForTopologyReady(ctx, bus); err != nil {
		return fmt.Errorf("redpanda topology not ready: %w", err)
	}
	logger.Info("✅ redpanda topology 就绪（jxt.topology.ready 存在）")

	// §七 启动期存在性断言：代码要订阅的 topic 必须已由 bootstrap 建好，缺失即 fail-fast
	// 并指名道姓给出修复指令。断言逻辑已上提 jxt-core WaitForTopicsExist（消除 4 服务重复）。
	logger.Infof("🔍 存在性断言 %d 个 topics...", len(topics))
	return eventbus.WaitForTopicsExist(ctx, bus, topics)
}

// buildEventRegistry 构造 process-management 的 topic 注册表(v4 动作1)。
// process.task.events 为本服务拥有(发布侧,单分区保序)。
func buildEventRegistry() *infra_eventbus.Registry {
	reg := infra_eventbus.NewRegistry()
	reg.Register("process.task.events")
	return reg
}
