package migrate

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"text/template"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"github.com/spf13/cobra"

	"jxt-evidence-system/process-management/cmd/migrate/migration"
	_ "jxt-evidence-system/process-management/cmd/migrate/migration/version"
	_ "jxt-evidence-system/process-management/cmd/migrate/migration/version-local"
	"jxt-evidence-system/process-management/shared/common/database"
	tenantdb "jxt-evidence-system/process-management/shared/common/tenantdb"

	"github.com/ChenBigdata421/jxt-core/sdk/config"
)

var (
	configYml string
	generate  bool
	goAdmin   bool
	tenantID  int
	StartCmd  = &cobra.Command{
		Use:     "migrate",
		Short:   "Initialize the database",
		Example: "go-admin migrate -c config/settings.yml",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

// fixme 在您看不见代码的时候运行迁移，我觉得是不安全的，所以编译后最好不要去执行迁移
func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/settings.yml", "Start server with provided configuration file")
	StartCmd.PersistentFlags().BoolVarP(&generate, "generate", "g", false, "generate migration file")
	StartCmd.PersistentFlags().BoolVarP(&goAdmin, "goAdmin", "a", false, "generate go-admin migration file")
	StartCmd.PersistentFlags().IntVarP(&tenantID, "domain", "d", 1, "select tenant id")
}

func run() {

	if !generate {
		fmt.Println(`start init`)
		//1. 读取配置
		config.Setup(configYml)
		//2.初始化日志
		logger.Setup()
		//3.初始化数据库
		initDB()
	} else {
		fmt.Println(`generate migration file`)
		_ = genFile()
	}
}

func migrateModel() error {
	if tenantID == 0 {
		tenantID = 1
	}
	db := sdk.Runtime.GetTenantDB(tenantID)

	if db == nil {
		return fmt.Errorf("未找到租户 %d 的数据库", tenantID)
	}

	// 从实际数据库连接获取驱动类型
	var driverName string
	if db.Dialector != nil {
		driverName = db.Dialector.Name()
	}

	// MySQL 特定配置
	if driverName == "mysql" {
		db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4")
	}

	// 使用多租户迁移接口
	migration.Migrate.SetTenantDb(tenantID, db.Debug())
	if err := migration.Migrate.MigrateTenant(tenantID); err != nil {
		return fmt.Errorf("租户 %d 迁移失败: %w", tenantID, err)
	}
	return nil
}
func initDB() {
	// 初始化租户缓存
	cache, err := tenantdb.Setup("process-management")
	if err != nil {
		log.Printf("警告：租户缓存初始化失败: %v", err)
	}
	if cache != nil {
		if err := tenantdb.StartWatcher(cache, nil); err != nil {
			log.Printf("警告：动态租户监听启动失败: %v", err)
		}
	}

	// 初始化命令数据库连接（内部会从 tenantdb 缓存获取配置）
	database.ProcessDbSetup()

	// 数据库迁移
	fmt.Println("数据库迁移开始")
	_ = migrateModel()
	fmt.Println(`数据库基础数据初始化成功`)
}

func genFile() error {
	t1, err := template.ParseFiles("template/migrate.template")
	if err != nil {
		return err
	}
	m := map[string]string{}
	m["GenerateTime"] = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	m["Package"] = "version_local"
	if goAdmin {
		m["Package"] = "version"
	}
	var b1 bytes.Buffer
	err = t1.Execute(&b1, m)
	if goAdmin {
		pkg.FileCreate(b1, "./cmd/migrate/migration/version/"+m["GenerateTime"]+"_migrate.go")
	} else {
		pkg.FileCreate(b1, "./cmd/migrate/migration/version-local/"+m["GenerateTime"]+"_migrate.go")
	}
	return nil
}
