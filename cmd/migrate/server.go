package migrate

import (
	"bytes"
	"fmt"
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
	"jxt-evidence-system/process-management/shared/common/models"

	"github.com/ChenBigdata421/jxt-core/sdk/config"
)

var (
	configYml string
	generate  bool
	goAdmin   bool
	tenantID  string
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
	StartCmd.PersistentFlags().StringVarP(&tenantID, "domain", "d", "*", "select tenant id")
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
	if tenantID == "" {
		tenantID = "*"
	}
	db := sdk.Runtime.GetTenantCommandDB(tenantID)

	if db == nil {
		return fmt.Errorf("未找到任何租户数据库")
	}

	// 获取数据库驱动配置
	var driver string
	if !config.TenantsConfig.Enabled {
		// 非多租户模式，直接使用全局配置
		driver = config.DatabaseConfig.CommandDB.Driver
	} else {
		// 多租户模式
		if config.TenantsConfig.List == nil {
			return fmt.Errorf("租户配置未初始化")
		}

		// 查找租户配置
		var tc *config.TenantConfig
		for i := range config.TenantsConfig.List {
			if config.TenantsConfig.List[i].ID == tenantID {
				tc = &config.TenantsConfig.List[i]
				break
			}
		}

		if tc == nil {
			return fmt.Errorf("租户ID未配置: %s", tenantID)
		}

		// 配置层级回退：租户配置 > 默认配置 > 全局配置
		if tc.Database.CommandDB.Driver != "" {
			driver = tc.Database.CommandDB.Driver
		} else if config.TenantsConfig.Defaults.Database.Driver != "" {
			driver = config.TenantsConfig.Defaults.Database.Driver
		} else {
			driver = config.DatabaseConfig.CommandDB.Driver
		}

		if driver == "" {
			return fmt.Errorf("无法确定租户%s的数据库驱动配置", tenantID)
		}
	}

	// MySQL特定配置
	if driver == "mysql" {
		db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4")
	}

	err := db.Debug().AutoMigrate(&models.Migration{})
	if err != nil {
		return err
	}
	migration.Migrate.SetDb(db.Debug())
	migration.Migrate.Migrate()
	return err
}
func initDB() {
	//3. 初始化工作流数据库链接
	database.ProcessDbSetup()
	//4. 数据库迁移
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
