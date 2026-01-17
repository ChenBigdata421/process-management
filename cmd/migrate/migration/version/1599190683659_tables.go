package version

import (
	"log"
	"runtime"

	"jxt-evidence-system/process-management/cmd/migrate/migration"
	inimodels "jxt-evidence-system/process-management/cmd/migrate/migration/models"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	models "jxt-evidence-system/process-management/shared/common/models"

	"gorm.io/gorm"
)

/*
这行代码使用了Go语言的runtime包来获取当前执行的文件名。runtime.Caller(0)函数返回当前调用栈
的某一层的程序计数器、文件名、行号和一个布尔值，表示是否成功获取信息。这里的参数0表示当前函数调
用本身的调用栈层级。这种做法通常用于获取当前执行代码的文件路径，例如在日志记录、错误报告或者是
像这里的数据库迁移脚本中自动获取版本号时。
*/
/*init函数有一个特殊的用途，它不需要被显式调用。被导入时会自动执行每个包中的所有init函数。
如果一个包导入了其他包，被导入包的init函数会先于导入它的包执行。
这里的init函数用于设置数据库迁移的版本号。*/
func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1599190683659Tables)
}

func _1599190683659Tables(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {

		// 1. 创建表结构
		err := tx.Migrator().AutoMigrate(
			&workflow_aggregate.Workflow{},
			&instance_aggregate.WorkflowInstance{},
			&task_aggregate.Task{},
			&task_aggregate.TaskHistory{},
		)
		log.Println(`数据表创建成功！！！ `)
		if err != nil {
			log.Println(`数据表失败: `, err.Error())
			return err
		}

		// 2. 初始化基础数据（执行db.sql）
		if err := inimodels.InitDb(tx); err != nil {
			log.Println(`数据表初始化失败: `, err.Error())
			return err
		}
		log.Println(`数据表初始化成功！！！ `)
		// 3. 最后插入版本记录（确保前面的步骤都成功）
		return tx.Create(&models.Migration{
			Version: version,
		}).Error
	})
}
