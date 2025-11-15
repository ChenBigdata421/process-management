package migration

import (
	"log"
	"path/filepath"
	"sort"
	"sync"

	"gorm.io/gorm"
)

var Migrate = &Migration{
	version: make(map[string]func(db *gorm.DB, version string) error),
}

type Migration struct {
	db      *gorm.DB
	version map[string]func(db *gorm.DB, version string) error
	mutex   sync.Mutex
}

func (e *Migration) GetDb() *gorm.DB {
	return e.db
}

func (e *Migration) SetDb(db *gorm.DB) {
	e.db = db
}

func (e *Migration) SetVersion(k string, f func(db *gorm.DB, version string) error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.version[k] = f
}

/*
Migrate方法负责按照版本号顺序执行数据库迁移。它通过查询sys_migration表
来跳过已经执行过的迁移，确保每个迁移只执行一次。如果在迁移过程中遇到任何错误，
程序将终止执行。这个方法是数据库版本控制和迁移管理的核心部分，确保数据库结构
的版本能够与应用程序的需求保持一致。
*/
func (e *Migration) Migrate() {
	versions := make([]string, 0)
	for k := range e.version { ////键是版本号，值是对应的迁移函数
		versions = append(versions, k)
	}
	if !sort.StringsAreSorted(versions) { //确保迁移是按照版本号顺序执行的。检查是否升序排序
		sort.Strings(versions) //升序排序
	}
	var err error
	var count int64
	for _, v := range versions {
		log.Printf("检查迁移版本: %s", v)
		err = e.db.Table("sys_migration").Where("version = ?", v).Count(&count).Error
		if err != nil {
			log.Fatalf("查询迁移版本失败: %v", err)
		}
		if count > 0 { //如果该版本已经存在（即count > 0），则跳过当前迁移。
			log.Printf("版本 %s 已存在，跳过迁移", v)
			count = 0
			continue
		}

		log.Printf("开始执行迁移版本: %s", v)
		//调用与该版本号关联的迁移函数
		err = (e.version[v])(e.db.Debug(), v)
		if err != nil {
			log.Fatalf("迁移版本 %s 执行失败: %v", v, err)
		}
		log.Printf("迁移版本 %s 执行成功", v)
	}
}

func GetFilename(s string) string {
	s = filepath.Base(s) //获取一个文件的基本名称（即不带路径的文件名）
	return s[:13]
}
