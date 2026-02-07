package persistence

import (
	"sync"

	delete_approval_repository "jxt-evidence-system/process-management/internal/domain/aggregate/delete_approval/repository"
	download_approval_repository "jxt-evidence-system/process-management/internal/domain/aggregate/download_approval/repository"
	instance_repository "jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
	"jxt-evidence-system/process-management/shared/common/di"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

var (
	registrations = make([]func(), 0)
	registerOnce  sync.Once
)

// jiyuanjie 负责依赖注入所有的 Repo 实例
func RegisterDependencies() {
	//遍历所有api的依赖注入方法
	registerOnce.Do(func() {
		for _, f := range registrations {
			f()
		}
	})
}

// workflowInstanceRepository的依赖注入
func registerWorkflowInstanceRepoDependencies() {
	if err := di.Provide(func() instance_repository.WorkflowInstanceRepository {
		return &workflowInstanceRepository{}
	}); err != nil {
		logger.Fatalf("failed to provide workflowInstanceRepository: %v", err)
	}
}

func registerWorkflowRepoDependencies() {
	if err := di.Provide(func() workflow_repository.WorkflowRepository {
		return &workflowRepository{}
	}); err != nil {
		logger.Fatalf("failed to provide workflowRepository: %v", err)
	}
}

func registerTaskRepoDependencies() {
	if err := di.Provide(func() task_repository.TaskRepository {
		return &taskRepository{}
	}); err != nil {
		logger.Fatalf("failed to provide taskRepository: %v", err)
	}
}

func registerTaskHistoryRepoDependencies() {
	if err := di.Provide(func() task_repository.TaskHistoryRepository {
		return &taskHistoryRepository{}
	}); err != nil {
		logger.Fatalf("failed to provide taskHistoryRepository: %v", err)
	}
}

func registerDownloadApprovalRepoDependencies() {
	if err := di.Provide(func() download_approval_repository.MediaDownloadApprovalRepository {
		return &GormMediaDownloadApprovalRepository{}
	}); err != nil {
		logger.Fatalf("failed to provide MediaDownloadApprovalRepository: %v", err)
	}
}

func registerDeleteApprovalRepoDependencies() {
	if err := di.Provide(func() delete_approval_repository.MediaDeleteApprovalRepository {
		return &GormMediaDeleteApprovalRepository{}
	}); err != nil {
		logger.Fatalf("failed to provide MediaDeleteApprovalRepository: %v", err)
	}
}

func init() {
	registrations = append(registrations,
		registerWorkflowInstanceRepoDependencies,
		registerWorkflowRepoDependencies,
		registerTaskRepoDependencies,
		registerTaskHistoryRepoDependencies,
		registerDownloadApprovalRepoDependencies,
		registerDeleteApprovalRepoDependencies,
	)
}
