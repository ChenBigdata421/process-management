package version

import (
	"log"
	"runtime"

	"jxt-evidence-system/process-management/cmd/migrate/migration"
	delete_approval "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval"
	download_approval "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval"
	models "jxt-evidence-system/process-management/shared/common/models"

	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1706345600000DownloadApproval)
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

		// 插入版本记录
		return tx.Create(&models.Migration{
			Version: version,
		}).Error
	})
}
