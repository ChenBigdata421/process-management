package version

import (
	"log"
	"runtime"

	"jxt-evidence-system/process-management/cmd/migrate/migration"

	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
	"gorm.io/gorm"
)

// 背景：jxt-core v1.1.69 的 OutboxEventModel 新增了发布侧死信字段 DeadLetteredAt / DlqNotifiedAt
// （见 jxt-core sdk/pkg/outbox/adapters/gorm/model.go），outbox GORM 仓储的 INSERT 会写入
// dead_lettered_at / dlq_notified_at 两列。
//
// 本服务 1706345600000 迁移虽已 AutoMigrate 过 OutboxEventModel，但当时模型尚无这两列；该版本
// 一旦记入 sys_migration 便不会重跑，导致升级 jxt-core 后新列缺失——所有 MediaDeleted 等事件
// 落库报 SQLSTATE 42703，删除审批"流程走完却不删文档"。
//
// 本版本重新 AutoMigrate OutboxEventModel 以补齐 dead_lettered_at / dlq_notified_at 列与
// idx_outbox_dlq_notify 复合索引。AutoMigrate 幂等：对已手动补列的库为 no-op，对全新库按当前
// 模型定义建表，从此与 jxt-core 模型演进保持一致。
func init() {
	_, fileName, _, _ := runtime.Caller(0)
	// 添加 _process 后缀区分 Process 服务的迁移版本
	migration.Migrate.RegisterVersion(migration.GetFilename(fileName)+"_process", _1785464640000OutboxDeadLetterColumns)
}

func _1785464640000OutboxDeadLetterColumns(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().AutoMigrate(&gormadapter.OutboxEventModel{}); err != nil {
			log.Println(`补齐 outbox_events 死信列失败: `, err.Error())
			return err
		}
		log.Println(`outbox_events 死信列补齐成功（dead_lettered_at / dlq_notified_at + idx_outbox_dlq_notify）`)
		// 注意：Migration 记录由 jxt-core 框架自动创建，无需手动插入
		return nil
	})
}
