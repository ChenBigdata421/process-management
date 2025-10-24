package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/domain/workflow"
	"github.com/jxt/process-management/interfaces/http/handler"
	"github.com/jxt/process-management/interfaces/http/router"
)

// MockWorkflowRepository 模拟工作流仓储
type MockWorkflowRepository struct {
	workflows map[string]*workflow.Workflow
}

func NewMockWorkflowRepository() *MockWorkflowRepository {
	return &MockWorkflowRepository{
		workflows: make(map[string]*workflow.Workflow),
	}
}

func (m *MockWorkflowRepository) Save(ctx context.Context, wf *workflow.Workflow) error {
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MockWorkflowRepository) FindByID(ctx context.Context, id string) (*workflow.Workflow, error) {
	if wf, ok := m.workflows[id]; ok {
		return wf, nil
	}
	return nil, workflow.ErrWorkflowNotFound
}

func (m *MockWorkflowRepository) FindAll(ctx context.Context, limit, offset int) ([]*workflow.Workflow, error) {
	var result []*workflow.Workflow
	for _, wf := range m.workflows {
		result = append(result, wf)
	}
	return result, nil
}

func (m *MockWorkflowRepository) Update(ctx context.Context, wf *workflow.Workflow) error {
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MockWorkflowRepository) Delete(ctx context.Context, id string) error {
	delete(m.workflows, id)
	return nil
}

func (m *MockWorkflowRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.workflows)), nil
}

// MockWorkflowInstanceRepository 模拟实例仓储
type MockWorkflowInstanceRepository struct {
	instances map[string]*workflow.WorkflowInstance
}

func NewMockWorkflowInstanceRepository() *MockWorkflowInstanceRepository {
	return &MockWorkflowInstanceRepository{
		instances: make(map[string]*workflow.WorkflowInstance),
	}
}

func (m *MockWorkflowInstanceRepository) Save(ctx context.Context, instance *workflow.WorkflowInstance) error {
	m.instances[instance.ID] = instance
	return nil
}

func (m *MockWorkflowInstanceRepository) FindByID(ctx context.Context, id string) (*workflow.WorkflowInstance, error) {
	if instance, ok := m.instances[id]; ok {
		return instance, nil
	}
	return nil, workflow.ErrInstanceNotFound
}

func (m *MockWorkflowInstanceRepository) FindByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*workflow.WorkflowInstance, error) {
	var result []*workflow.WorkflowInstance
	for _, instance := range m.instances {
		if instance.WorkflowID == workflowID {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (m *MockWorkflowInstanceRepository) Update(ctx context.Context, instance *workflow.WorkflowInstance) error {
	m.instances[instance.ID] = instance
	return nil
}

func (m *MockWorkflowInstanceRepository) Delete(ctx context.Context, id string) error {
	delete(m.instances, id)
	return nil
}

func (m *MockWorkflowInstanceRepository) CountByWorkflowID(ctx context.Context, workflowID string) (int64, error) {
	count := int64(0)
	for _, instance := range m.instances {
		if instance.WorkflowID == workflowID {
			count++
		}
	}
	return count, nil
}

func (m *MockWorkflowInstanceRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*workflow.WorkflowInstance, int, error) {
	var result []*workflow.WorkflowInstance
	for _, instance := range m.instances {
		result = append(result, instance)
	}
	return result, len(result), nil
}

func setupTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// 初始化仓储
	workflowRepo := NewMockWorkflowRepository()
	instanceRepo := NewMockWorkflowInstanceRepository()

	// 初始化命令处理器
	createHandler := command.NewCreateWorkflowHandler(workflowRepo)
	updateHandler := command.NewUpdateWorkflowHandler(workflowRepo)
	deleteHandler := command.NewDeleteWorkflowHandler(workflowRepo)
	activateHandler := command.NewActivateWorkflowHandler(workflowRepo)
	freezeHandler := command.NewFreezeWorkflowHandler(workflowRepo)
	startInstanceHandler := command.NewStartWorkflowInstanceHandler(workflowRepo, instanceRepo, nil)

	// 初始化查询服务
	workflowQueryService := query.NewWorkflowQueryService(workflowRepo)
	instanceQueryService := query.NewWorkflowInstanceQueryService(instanceRepo)

	// 初始化HTTP处理器
	workflowHandler := handler.NewWorkflowHandler(
		createHandler,
		updateHandler,
		deleteHandler,
		activateHandler,
		freezeHandler,
		workflowQueryService,
		instanceRepo,
	)
	instanceHandler := handler.NewInstanceHandler(
		startInstanceHandler,
		instanceQueryService,
	)

	// 创建任务仓储模拟
	mockTaskRepo := NewMockTaskRepository()
	mockTaskHistoryRepo := NewMockTaskHistoryRepository()

	// 初始化任务命令处理器
	createTaskHandler := command.NewCreateTaskHandler(mockTaskRepo)
	claimTaskHandler := command.NewClaimTaskHandler(mockTaskRepo, mockTaskHistoryRepo)
	completeTaskHandler := command.NewCompleteTaskHandler(mockTaskRepo, mockTaskHistoryRepo, nil)
	delegateTaskHandler := command.NewDelegateTaskHandler(mockTaskRepo, mockTaskHistoryRepo)
	deleteTaskHandler := command.NewDeleteTaskHandler(mockTaskRepo)

	// 初始化任务查询服务
	taskQueryService := query.NewTaskQueryService(mockTaskRepo, mockTaskHistoryRepo, workflowRepo)

	// 初始化任务处理器
	taskHandler := handler.NewTaskHandler(
		createTaskHandler,
		claimTaskHandler,
		completeTaskHandler,
		delegateTaskHandler,
		deleteTaskHandler,
		taskQueryService,
	)

	// 设置路由
	router.SetupRoutes(engine, workflowHandler, instanceHandler, taskHandler)

	return engine
}

func TestCreateWorkflowAPI(t *testing.T) {
	engine := setupTestEngine()

	body := map[string]string{
		"name":        "订单处理流程",
		"description": "处理订单的业务流程",
		"definition":  `{"steps": ["validate", "process", "notify"]}`,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/workflows", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}
}

func TestListWorkflowsAPI(t *testing.T) {
	engine := setupTestEngine()

	req := httptest.NewRequest("GET", "/api/workflows?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 200 {
		t.Errorf("Expected code 200, got %v", response["code"])
	}
}

func TestHealthCheckAPI(t *testing.T) {
	engine := setupTestEngine()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// MockTaskRepository 模拟任务仓储
type MockTaskRepository struct {
	tasks map[string]*workflow.Task
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks: make(map[string]*workflow.Task),
	}
}

func (m *MockTaskRepository) Save(ctx context.Context, task *workflow.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*workflow.Task, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, nil
}

func (m *MockTaskRepository) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*workflow.Task, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.InstanceID == instanceID {
			result = append(result, task)
		}
	}
	return result, nil
}

func (m *MockTaskRepository) FindTodoByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.Assignee == assignee && (task.Status == workflow.TaskStatusPending || task.Status == workflow.TaskStatusClaimed) {
			result = append(result, task)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) FindDoneByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.Assignee == assignee && (task.Status == workflow.TaskStatusCompleted || task.Status == workflow.TaskStatusRejected) {
			result = append(result, task)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) FindClaimable(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.Status == workflow.TaskStatusPending {
			if task.Assignee == "" || task.Assignee == userID {
				result = append(result, task)
			}
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		result = append(result, task)
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *workflow.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

// MockTaskHistoryRepository 模拟任务历史仓储
type MockTaskHistoryRepository struct {
	histories map[string]*workflow.TaskHistory
}

func NewMockTaskHistoryRepository() *MockTaskHistoryRepository {
	return &MockTaskHistoryRepository{
		histories: make(map[string]*workflow.TaskHistory),
	}
}

func (m *MockTaskHistoryRepository) Save(ctx context.Context, history *workflow.TaskHistory) error {
	m.histories[history.ID] = history
	return nil
}

func (m *MockTaskHistoryRepository) FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]*workflow.TaskHistory, error) {
	var result []*workflow.TaskHistory
	for _, history := range m.histories {
		if history.TaskID == taskID {
			result = append(result, history)
		}
	}
	return result, nil
}

func (m *MockTaskHistoryRepository) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*workflow.TaskHistory, error) {
	var result []*workflow.TaskHistory
	for _, history := range m.histories {
		if history.InstanceID == instanceID {
			result = append(result, history)
		}
	}
	return result, nil
}
