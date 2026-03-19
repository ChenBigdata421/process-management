package version

import (
	"log"
	"runtime"

	"jxt-evidence-system/process-management/cmd/migrate/migration"
	delete_approval "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval"
	download_approval "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval"

	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	// 添加 _process 后缀区分 Process 服务的迁移版本
	migration.Migrate.RegisterVersion(migration.GetFilename(fileName)+"_process", _1706345600000DownloadApproval)
}

func _1706345600000DownloadApproval(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {

		// 创建媒体下载审批表
		var modelsToMigrate []interface{}
		modelsToMigrate = append(modelsToMigrate, &download_approval.MediaDownloadApproval{})
		modelsToMigrate = append(modelsToMigrate, &delete_approval.MediaDeleteApproval{})
		modelsToMigrate = append(modelsToMigrate, &gormadapter.OutboxEventModel{})

		err := tx.Migrator().AutoMigrate(modelsToMigrate...)
		if err != nil {
			log.Println(`创建媒体下载审批表失败: `, err.Error())
			return err
		}
		log.Println(`媒体下载审批表创建成功！`)

		// 注意：Migration 记录由 jxt-core 框架自动创建，无需手动插入
		return nil
	})
}
