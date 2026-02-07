package api

import (
	"jxt-evidence-system/process-management/internal/application/service"
	"jxt-evidence-system/process-management/internal/application/service/port"
	"jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	domain_websocket "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	"jxt-evidence-system/process-management/shared/common/di"
	"sync"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

var (
	registrations = make([]func(), 0)
	registerOnce  sync.Once
)

// 基于 jiyuanjie 负责依赖注入所有的 api 实例
func RegisterDependencies() {
	//遍历所有api的依赖注入方法
	registerOnce.Do(func() {
		for _, f := range registrations {
			f()
		}
	})
}

func init() {
	registrations = append(registrations,
		registerWorkflowApiDependencies,
		registerInstanceApiDependencies,
		registerTaskApiDependencies,
		registerWebSocketApiDependencies,
		registerDownloadApprovalApiDependencies,
		registerDeleteApprovalApiDependencies,
	)
}

func registerDownloadApprovalApiDependencies() {
	err := di.Provide(func(
		downloadApprovalService *service.DownloadApprovalService,
		instanceService port.InstanceService,
		workflowService port.WorkflowService,
	) *DownloadApprovalHandler {
		return NewDownloadApprovalHandler(downloadApprovalService, instanceService, workflowService)
	})
	if err != nil {
		logger.Fatalf("Failed to provide DownloadApprovalHandler: %v", err)
	}
}

func registerDeleteApprovalApiDependencies() {
	err := di.Provide(func(
		deleteApprovalService *service.DeleteApprovalService,
		instanceService port.InstanceService,
		workflowService port.WorkflowService,
	) *DeleteApprovalHandler {
		return NewDeleteApprovalHandler(deleteApprovalService, instanceService, workflowService)
	})
	if err != nil {
		logger.Fatalf("Failed to provide DeleteApprovalHandler: %v", err)
	}
}

func registerInstanceApiDependencies() {
	err := di.Provide(func(instanceService port.InstanceService) *InstanceHandler {
		return &InstanceHandler{
			instanceService: instanceService,
		}
	})
	if err != nil {
		logger.Fatalf("Failed to provide InstanceHandler: %v", err)
	}
}

func registerTaskApiDependencies() {
	err := di.Provide(func(taskService port.TaskService) *TaskHandler {
		return &TaskHandler{
			taskService: taskService,
		}
	})
	if err != nil {
		logger.Fatalf("Failed to provide TaskHandler: %v", err)
	}
}

func registerWorkflowApiDependencies() {
	err := di.Provide(func(
		workflowService port.WorkflowService,
		instanceService port.InstanceService,
		instanceRepo repository.WorkflowInstanceRepository,
	) *WorkflowHandler {
		return &WorkflowHandler{
			workflowService: workflowService,
			instanceService: instanceService,
			instanceRepo:    instanceRepo,
		}
	})
	if err != nil {
		logger.Fatalf("Failed to provide WorkflowHandler: %v", err)
	}
}

func registerWebSocketApiDependencies() {
	err := di.Provide(func(wsNotifier domain_websocket.WebSocketNotifier) *WebSocketHandler {
		return NewWebSocketHandler(wsNotifier)
	})
	if err != nil {
		logger.Fatalf("Failed to provide WebSocketHandler: %v", err)
	}
}
