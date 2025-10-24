package workflow

import (
	"context"
	"log"

	"github.com/cschleiden/go-workflows/backend"
	"github.com/cschleiden/go-workflows/backend/sqlite"
	"github.com/cschleiden/go-workflows/worker"
)

// WorkflowExecutor 工作流执行器
type WorkflowExecutor struct {
	backend backend.Backend
	worker  *worker.Worker
}

// NewWorkflowExecutor 创建新的工作流执行器
func NewWorkflowExecutor(dbPath string) (*WorkflowExecutor, error) {
	// 创建 SQLite 后端
	b := sqlite.NewSqliteBackend(dbPath, sqlite.WithApplyMigrations(true))

	// 创建 worker
	w := worker.New(b, &worker.Options{
		SingleWorkerMode: true,
	})

	return &WorkflowExecutor{
		backend: b,
		worker:  w,
	}, nil
}

// RegisterWorkflows 注册工作流
func (e *WorkflowExecutor) RegisterWorkflows() {
	e.worker.RegisterWorkflow(ProcessWorkflow)
	log.Println("Workflow registered: ProcessWorkflow")

	e.worker.RegisterWorkflow(OrderProcessingWorkflow)
	log.Println("Workflow registered: OrderProcessingWorkflow")

	e.worker.RegisterWorkflow(DynamicWorkflow)
	log.Println("Workflow registered: DynamicWorkflow")
}

// RegisterActivities 注册活动
func (e *WorkflowExecutor) RegisterActivities() {
	// 通用活动
	e.worker.RegisterActivity(ValidateActivity)
	e.worker.RegisterActivity(ProcessActivity)
	e.worker.RegisterActivity(NotifyActivity)
	e.worker.RegisterActivity(CompleteActivity)
	log.Println("Activities registered: ValidateActivity, ProcessActivity, NotifyActivity, CompleteActivity")

	// 订单处理活动
	e.worker.RegisterActivity(OrderValidateActivity)
	e.worker.RegisterActivity(OrderProcessActivity)
	e.worker.RegisterActivity(OrderNotifyActivity)
	e.worker.RegisterActivity(OrderCompleteActivity)
	log.Println("Activities registered: OrderValidateActivity, OrderProcessActivity, OrderNotifyActivity, OrderCompleteActivity")
}

// Start 启动工作流执行器
func (e *WorkflowExecutor) Start(ctx context.Context) error {
	log.Println("Starting workflow executor...")
	return e.worker.Start(ctx)
}

// Stop 停止工作流执行器
func (e *WorkflowExecutor) Stop(ctx context.Context) error {
	log.Println("Stopping workflow executor...")
	// worker 没有 Stop 方法，只需关闭后端
	return e.backend.Close()
}

// Close 关闭执行器
func (e *WorkflowExecutor) Close() error {
	log.Println("Closing workflow executor...")
	return e.backend.Close()
}
