package service

import (
	"sync"

	"jxt-evidence-system/process-management/internal/application/service/port"
	instance_repository "jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	websocket "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
	domain_service "jxt-evidence-system/process-management/internal/domain/service"
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
		registerTaskServiceDependencies,
		registerWorkflowServiceDependencies,
		registerInstanceServiceDependencies,
		registerNotificationServiceDependencies,
		registerWorkflowEngineServiceDependencies,
	)
}

func registerTaskServiceDependencies() {
	err := di.Provide(func(
		taskRepo task_repository.TaskRepository,
		historyRepo task_repository.TaskHistoryRepository,
		workflowRepo workflow_repository.WorkflowRepository,
		engineService port.WorkflowEngineService,
	) port.TaskService {
		return &taskService{
			taskRepo:      taskRepo,
			historyRepo:   historyRepo,
			workflowRepo:  workflowRepo,
			engineService: engineService,
		}
	})
	if err != nil {
		logger.Fatalf("Failed to provide TaskService: %v", err)
	}
}

func registerWorkflowServiceDependencies() {
	err := di.Provide(func(
		workflowRepo workflow_repository.WorkflowRepository,
	) port.WorkflowService {
		return &workflowService{
			repo: workflowRepo,
		}
	})
	if err != nil {
		logger.Fatalf("Failed to provide WorkflowService: %v", err)
	}
}

func registerInstanceServiceDependencies() {
	err := di.Provide(func(
		workflowService port.WorkflowService,
		instanceRepo instance_repository.WorkflowInstanceRepository,
		engineService port.WorkflowEngineService,
		taskService port.TaskService,
		domainService *domain_service.WorkflowDomainService,
	) port.InstanceService {
		return &instanceService{
			workflowService: workflowService,
			instanceRepo:    instanceRepo,
			engineService:   engineService,
			taskService:     taskService,
			domainService:   *domainService,
		}
	})
	if err != nil {
		logger.Fatalf("Failed to provide InstanceService: %v", err)
	}
}

func registerNotificationServiceDependencies() {
	err := di.Provide(func(wsHub websocket.WebSocketNotifier) port.NotificationService {
		return NewNotificationService(wsHub)
	})
	if err != nil {
		logger.Fatalf("Failed to provide NotificationService: %v", err)
	}
}

func registerWorkflowEngineServiceDependencies() {
	err := di.Provide(func(
		workflowRepo workflow_repository.WorkflowRepository,
		instanceRepo instance_repository.WorkflowInstanceRepository,
		taskRepo task_repository.TaskRepository,
		domainService *domain_service.WorkflowDomainService,
	) port.WorkflowEngineService {
		return NewWorkflowEngineService(workflowRepo, instanceRepo, taskRepo, *domainService)
	})
	if err != nil {
		logger.Fatalf("Failed to provide WorkflowEngineService: %v", err)
	}
}
