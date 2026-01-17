package domain_service

import (
	"sync"

	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	"jxt-evidence-system/process-management/shared/common/di"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

var (
	registrations = make([]func(), 0)
	registerOnce  = sync.Once{}
)

// jiyuanjie 负责依赖注入所有的 application service 实例
func RegisterDependencies() {
	//遍历所有application service的依赖注入方法
	registerOnce.Do(func() {
		for _, f := range registrations {
			f()
		}
	})
}

func init() {
	registrations = append(registrations,
		registerWorkflowServiceDependencies,
	)
}

func registerWorkflowServiceDependencies() {
	err := di.Provide(func(taskRepo task_repository.TaskRepository) *WorkflowDomainService {
		return NewWorkflowDomainService(taskRepo)
	})
	if err != nil {
		logger.Fatalf("Failed to provide WorkflowDomainService: %v", err)
	}
}
