package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/domain/workflow"
)

// TaskHandler 任务HTTP处理器
type TaskHandler struct {
	createHandler   *command.CreateTaskHandler
	claimHandler    *command.ClaimTaskHandler
	completeHandler *command.CompleteTaskHandler
	delegateHandler *command.DelegateTaskHandler
	deleteHandler   *command.DeleteTaskHandler
	queryService    *query.TaskQueryService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(
	createHandler *command.CreateTaskHandler,
	claimHandler *command.ClaimTaskHandler,
	completeHandler *command.CompleteTaskHandler,
	delegateHandler *command.DelegateTaskHandler,
	deleteHandler *command.DeleteTaskHandler,
	queryService *query.TaskQueryService,
) *TaskHandler {
	return &TaskHandler{
		createHandler:   createHandler,
		claimHandler:    claimHandler,
		completeHandler: completeHandler,
		delegateHandler: delegateHandler,
		deleteHandler:   deleteHandler,
		queryService:    queryService,
	}
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		InstanceID      string   `json:"instance_id" binding:"required"`
		WorkflowID      string   `json:"workflow_id" binding:"required"`
		TaskName        string   `json:"task_name" binding:"required"`
		TaskKey         string   `json:"task_key" binding:"required"`
		Description     string   `json:"description"`
		Assignee        string   `json:"assignee"`
		CandidateUsers  []string `json:"candidate_users"`
		CandidateGroups []string `json:"candidate_groups"`
		Priority        string   `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	cmd := &command.CreateTaskCommand{
		InstanceID:      req.InstanceID,
		WorkflowID:      req.WorkflowID,
		TaskName:        req.TaskName,
		TaskKey:         req.TaskKey,
		Description:     req.Description,
		Assignee:        req.Assignee,
		CandidateUsers:  req.CandidateUsers,
		CandidateGroups: req.CandidateGroups,
		Priority:        req.Priority,
	}

	id, err := h.createHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{"id": id},
	})
}

// ListAllTasks 查询所有任务
func (h *TaskHandler) ListAllTasks(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if taskName := c.Query("task_name"); taskName != "" {
		filters["task_name"] = taskName
	}
	if workflowName := c.Query("workflow_name"); workflowName != "" {
		filters["workflow_name"] = workflowName
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if assignee := c.Query("assignee"); assignee != "" {
		filters["assignee"] = assignee
	}

	tasks, total, err := h.queryService.ListAllTasks(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items": tasks,
			"total": total,
		},
		"msg": "success",
	})
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.queryService.GetTaskByID(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	if task == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 404,
			"msg":  "task not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": task,
		"msg":  "success",
	})
}

// ListTodoTasks 查询待办任务
func (h *TaskHandler) ListTodoTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	tasks, total, err := h.queryService.ListTodoTasks(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items": tasks,
			"total": total,
		},
		"msg": "success",
	})
}

// ListDoneTasks 查询已办任务
func (h *TaskHandler) ListDoneTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	tasks, total, err := h.queryService.ListDoneTasks(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items": tasks,
			"total": total,
		},
		"msg": "success",
	})
}

// ListClaimableTasks 查询可认领的任务
func (h *TaskHandler) ListClaimableTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	// 从上下文获取用户组
	userGroups := []string{}
	if groups, exists := c.Get("user_groups"); exists {
		if g, ok := groups.([]string); ok {
			userGroups = g
		}
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	tasks, total, err := h.queryService.ListClaimableTasks(c.Request.Context(), userID, userGroups, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items": tasks,
			"total": total,
		},
		"msg": "success",
	})
}

// ClaimTask 认领任务
func (h *TaskHandler) ClaimTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	cmd := &command.ClaimTaskCommand{
		TaskID: taskID,
		UserID: userID,
	}

	if err := h.claimHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// CompleteTask 完成任务
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	var req struct {
		Output  string `json:"output"`
		Comment string `json:"comment"`
		Result  string `json:"result"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	result := workflow.TaskResultCompleted
	if req.Result == "rejected" {
		result = workflow.TaskResultRejected
	}

	cmd := &command.CompleteTaskCommand{
		TaskID:  taskID,
		UserID:  userID,
		Output:  req.Output,
		Comment: req.Comment,
		Result:  result,
	}

	if err := h.completeHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// ApproveTask 批准任务
func (h *TaskHandler) ApproveTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	var req struct {
		Output  string `json:"output"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	cmd := &command.CompleteTaskCommand{
		TaskID:  taskID,
		UserID:  userID,
		Output:  req.Output,
		Comment: req.Comment,
		Result:  workflow.TaskResultCompleted,
	}

	if err := h.completeHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// RejectTask 驳回任务
func (h *TaskHandler) RejectTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	var req struct {
		Output  string `json:"output"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	cmd := &command.CompleteTaskCommand{
		TaskID:  taskID,
		UserID:  userID,
		Output:  req.Output,
		Comment: req.Comment,
		Result:  workflow.TaskResultRejected,
	}

	if err := h.completeHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// DelegateTask 转办任务
func (h *TaskHandler) DelegateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	var req struct {
		TargetID string `json:"target_id"`
		Comment  string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 验证 TargetID 不为空
	if req.TargetID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "target_id is required",
		})
		return
	}

	cmd := &command.DelegateTaskCommand{
		TaskID:   taskID,
		UserID:   userID,
		TargetID: req.TargetID,
		Comment:  req.Comment,
	}

	if err := h.delegateHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// GetTaskHistory 获取任务历史
func (h *TaskHandler) GetTaskHistory(c *gin.Context) {
	taskID := c.Param("id")

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	histories, err := h.queryService.GetTaskHistory(c.Request.Context(), taskID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": histories,
		"msg":  "success",
	})
}

// GetInstanceTaskHistory 获取实例的任务历史
func (h *TaskHandler) GetInstanceTaskHistory(c *gin.Context) {
	instanceID := c.Param("instance_id")

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	histories, err := h.queryService.GetInstanceTaskHistory(c.Request.Context(), instanceID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": histories,
		"msg":  "success",
	})
}

// GetInstanceTasks 获取实例的所有任务（包含当前状态）
func (h *TaskHandler) GetInstanceTasks(c *gin.Context) {
	instanceID := c.Param("instance_id")

	limit := 1000
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	tasks, total, err := h.queryService.GetInstanceTasks(c.Request.Context(), instanceID, limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"items": tasks,
			"total": total,
		},
		"msg": "success",
	})
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")

	cmd := &command.DeleteTaskCommand{
		TaskID: taskID,
	}

	if err := h.deleteHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}
