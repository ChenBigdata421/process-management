package database

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"jxt-evidence-system/process-management/cmd/migrate/migration"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/models"

	"github.com/ChenBigdata421/jxt-core/sdk"
	toolsConfig "github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	mylogger "github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	toolsDB "github.com/ChenBigdata421/jxt-core/tools/database"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// 数据库连接重试配置
const (
	maxRetries     = 10               // 最大重试次数
	initialBackoff = 1 * time.Second  // 初始退避时间
	maxBackoff     = 30 * time.Second // 最大退避时间
)

// connectWithRetry 带重试机制的数据库连接函数
func connectWithRetry(
	dbName string,
	driverStr string,
	resolverConfig toolsDB.Configure,
	gormConfig *gorm.Config,
) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = resolverConfig.Init(gormConfig, opens[driverStr])
		if err == nil {
			if attempt > 1 {
				log.Printf(pkg.Green("[%s] 数据库连接成功 (重试 %d 次后)\n"), dbName, attempt-1)
			}
			return db, nil
		}

		if attempt < maxRetries {
			log.Printf(pkg.Yellow("[%s] 数据库连接失败 (尝试 %d/%d): %v, %v 后重试...\n"),
				dbName, attempt, maxRetries, err, backoff)
			time.Sleep(backoff)
			// 指数退避，但不超过最大值
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return nil, fmt.Errorf("[%s] 数据库连接失败，已重试 %d 次: %w", dbName, maxRetries, err)
}

func ProcessDbSetup() {
	// 如果未配置多租户，只初始化默认数据库，唯一的tenantID设置为"*"
	if !toolsConfig.TenantsConfig.Enabled { //如果多租户为false，直接从config.Database获取
		setupProcessDatabase("*", toolsConfig.DatabaseConfig)
		return
	}

	// 如果多租户为true，则初始化每个租户的数据库
	for k := range toolsConfig.TenantsConfig.List {
		setupProcessDatabase(toolsConfig.TenantsConfig.List[k].ID, &toolsConfig.TenantsConfig.List[k].Database)
	}
}

/*
dbresolver 插件为 GORM 提供了数据库读写分离和分片的功能支持。通过配置该插件，
开发者可以轻松地实现数据库的读写分离和分片策略
*/
func setupProcessDatabase(tenantID string, c *toolsConfig.Database) {
	// 获取配置的优先级逻辑：
	// 1. 如果多租户为false，直接从config.Database获取
	// 2. 如果多租户为true：
	//    a. 先从Tenants.list[tenantID].Database获取
	//    b. 如果没有，再从Tenants.Defaults.Database获取
	//    c. 如果还没有，最后从config.Database获取
	getConfig := func(field string, required bool) interface{} {
		var result interface{}

		// 如果多租户为false，直接从config.Database获取
		if !toolsConfig.TenantsConfig.Enabled {
			result = getFieldValue(toolsConfig.DatabaseConfig, field)
		} else {
			// 多租户为true的情况
			// 1. 先从当前租户配置获取
			if c != nil {
				result = getFieldValue(c, field)
				if result != nil && result != "" && result != 0 {
					return result
				}
			}

			// 2. 然后从默认租户配置获取
			result = getFieldValue(&toolsConfig.TenantsConfig.Defaults.Database, field)
			if result != nil && result != "" && result != 0 {
				return result
			}

			// 3. 最后从全局数据库配置获取
			result = getFieldValue(toolsConfig.DatabaseConfig, field)
		}

		// 如果结果为nil或空值
		if result == nil || result == "" || result == 0 {
			// 如果是必需的配置项，报错
			if required {
				log.Fatalf("错误：无法从配置中获取 %s，请检查配置文件", field)
			}
			// 否则返回nil
			return nil
		}

		return result
	}

	// 1. 获取所有配置值
	// 获取数据库驱动（必需）
	driverValue := getConfig("ProcessDB.Driver", true)
	var driverStr string
	if str, ok := driverValue.(string); ok {
		driverStr = str
	} else {
		log.Fatalf("错误：ProcessDB.Driver 不是字符串类型：%v", driverValue)
	}

	// 设置全局驱动变量
	if global.ProcessDriver == "" {
		global.ProcessDriver = driverStr
	}

	// 获取数据库连接字符串（必需）
	sourceValue := getConfig("ProcessDB.Source", true)
	var sourceStr string
	if str, ok := sourceValue.(string); ok {
		sourceStr = str
	} else {
		log.Fatalf("错误：ProcessDB.Source 不是字符串类型：%v", sourceValue)
	}

	// 获取连接池配置（非必需，有默认值）
	maxIdleConnsValue := getConfig("ProcessDB.MaxIdleConns", false)
	maxOpenConnsValue := getConfig("ProcessDB.MaxOpenConns", false)
	connMaxIdleTimeValue := getConfig("ProcessDB.ConnMaxIdleTime", false)
	connMaxLifeTimeValue := getConfig("ProcessDB.ConnMaxLifeTime", false)

	// 转换为整数类型，并设置默认值
	maxIdleConns := 10 // 默认值
	if maxIdleConnsValue != nil {
		if val, ok := maxIdleConnsValue.(int); ok && val > 0 {
			maxIdleConns = val
		} else {
			log.Printf("警告：ProcessDB.MaxIdleConns 不是有效的整数类型，使用默认值：%d", maxIdleConns)
		}
	}

	maxOpenConns := 100 // 默认值
	if maxOpenConnsValue != nil {
		if val, ok := maxOpenConnsValue.(int); ok && val > 0 {
			maxOpenConns = val
		} else {
			log.Printf("警告：ProcessDB.MaxOpenConns 不是有效的整数类型，使用默认值：%d", maxOpenConns)
		}
	}

	connMaxIdleTime := 60 // 默认值
	if connMaxIdleTimeValue != nil {
		if val, ok := connMaxIdleTimeValue.(int); ok && val > 0 {
			connMaxIdleTime = val
		} else {
			log.Printf("警告：ProcessDB.ConnMaxIdleTime 不是有效的整数类型，使用默认值：%d", connMaxIdleTime)
		}
	}

	connMaxLifeTime := 3600 // 默认值
	if connMaxLifeTimeValue != nil {
		if val, ok := connMaxLifeTimeValue.(int); ok && val > 0 {
			connMaxLifeTime = val
		} else {
			log.Printf("警告：ProcessDB.ConnMaxLifeTime 不是有效的整数类型，使用默认值：%d", connMaxLifeTime)
		}
	}

	// 2. 检查配置值的有效性
	// 检查驱动类型是否支持
	if _, ok := opens[driverStr]; !ok {
		log.Printf("警告：不支持的数据库驱动类型 %s，将使用默认的 postgres", driverStr)
		driverStr = "postgres"
	}

	// 打印连接信息
	log.Printf("%s => %s", tenantID, pkg.Green(sourceStr))

	// 处理注册器
	var registers []toolsDB.ResolverConfigure
	if c != nil && len(c.ProcessDB.Registers) > 0 {
		registers = make([]toolsDB.ResolverConfigure, len(c.ProcessDB.Registers))
		for i := range c.ProcessDB.Registers {
			registers[i] = toolsDB.NewResolverConfigure(
				c.ProcessDB.Registers[i].Sources,
				c.ProcessDB.Registers[i].Replicas,
				c.ProcessDB.Registers[i].Policy,
				c.ProcessDB.Registers[i].Tables)
		}
	} else {
		registers = []toolsDB.ResolverConfigure{}
	}

	// 3. 使用配置值初始化数据库
	resolverConfig := toolsDB.NewConfigure(
		sourceStr,
		maxIdleConns,
		maxOpenConns,
		connMaxIdleTime,
		connMaxLifeTime,
		registers)

	// 使用带重试机制的连接函数
	db, err := connectWithRetry("ProcessDB", driverStr, resolverConfig, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: mylogger.NewGormLogger(mylogger.Logger, toolsConfig.LoggerConfig.GormLoggerLevel),
	})

	if err != nil {
		log.Fatalf(pkg.Red("%s connect error : %v\n"), driverStr, err)
	} else {
		log.Printf(pkg.Green("%s connect success ! \n"), driverStr)
	}

	sdk.Runtime.SetTenantDB(tenantID, db)

	// 只有在 dev 开发环境下才执行数据库迁移
	if toolsConfig.ApplicationConfig.Mode == "dev" {
		log.Println("数据库迁移开始（仅在 dev 环境执行）")
		if err := db.Debug().AutoMigrate(&models.Migration{}); err != nil {
			log.Println(pkg.Red("数据库迁移失败: %v\n"), err)
		}

		migration.Migrate.SetDb(db.Debug())
		migration.Migrate.Migrate()
		log.Println(`数据库基础数据初始化成功`)
	} else {
		log.Printf("当前环境为 %s，跳过数据库迁移（仅在 dev 环境执行迁移）\n", toolsConfig.ApplicationConfig.Mode)
	}
}

// getFieldValue 通用获取字段值函数，支持Database和DatabaseDefaults类型
func getFieldValue(config interface{}, field string) interface{} {
	if config == nil {
		return 0 // 返回0而不是空字符串
	}

	r := reflect.ValueOf(config)
	if r.Kind() == reflect.Ptr {
		if r.IsNil() {
			return 0
		}
		r = r.Elem()
	}

	// 处理嵌套字段，如 "ProcessDB.Driver"
	fields := strings.Split(field, ".")
	current := r

	for _, f := range fields {
		if !current.IsValid() {
			log.Printf("警告: 字段路径 %s 无效", field)
			return 0
		}

		if current.Kind() == reflect.Struct {
			current = current.FieldByName(f)
		} else {
			log.Printf("警告: 字段路径 %s 不是结构体", field)
			return 0
		}
	}

	if !current.IsValid() {
		log.Printf("警告: 字段 %s 不存在", field)
		return 0
	}

	return current.Interface()
}

// getIntConfig 获取整数类型配置
func getIntConfig(c *toolsConfig.Database, field string, defaultValue int) int {
	// 使用与setupSimpleDatabase中相同的逻辑获取配置
	// 1. 如果多租户为false，直接从config.Database获取
	if !toolsConfig.TenantsConfig.Enabled {
		val := getFieldValue(toolsConfig.DatabaseConfig, field)
		if intVal, ok := val.(int); ok && intVal != 0 {
			return intVal
		}
		return defaultValue
	}

	// 2. 多租户为true的情况
	// a. 先从当前租户配置获取
	if c != nil {
		val := getFieldValue(c, field)
		if intVal, ok := val.(int); ok && intVal != 0 {
			return intVal
		}
	}

	// b. 然后从默认租户配置获取
	val := getFieldValue(&toolsConfig.TenantsConfig.Defaults.Database, field)
	if intVal, ok := val.(int); ok && intVal != 0 {
		return intVal
	}

	// c. 最后从全局数据库配置获取
	val = getFieldValue(toolsConfig.DatabaseConfig, field)
	if intVal, ok := val.(int); ok && intVal != 0 {
		return intVal
	}

	return defaultValue
}
