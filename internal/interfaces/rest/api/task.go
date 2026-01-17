package api

import (
	"fmt"
	"net/http"
	"strconv"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
	"jxt-evidence-system/process-management/shared/common/status"

	jwtuser "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/gin-gonic/gin"
)

// TaskHandler 任务HTTP处理器
type TaskHandler struct {
	taskService port.TaskService
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	cmd := command.CreateTaskCommand{}

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	id, err := h.taskService.CreateTask(c.Request.Context(), &cmd)
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

	tasks, total, err := h.taskService.ListAllTasks(c.Request.Context(), filters, limit, offset)
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

	task, err := h.taskService.GetTaskByID(c.Request.Context(), taskID)
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
	userID := fmt.Sprintf("%d", jwtuser.GetUserId(c))
	if userID == "0" {
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

	tasks, total, err := h.taskService.ListTodoTasks(c.Request.Context(), userID, limit, offset)
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
	userID := fmt.Sprintf("%d", jwtuser.GetUserId(c))
	if userID == "0" {
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

	tasks, total, err := h.taskService.ListDoneTasks(c.Request.Context(), userID, limit, offset)
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
	userID := jwtuser.GetUserId(c)
	if userID == 0 {
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

	tasks, total, err := h.taskService.ListClaimableTasks(c.Request.Context(), strconv.Itoa(userID), userGroups, limit, offset)
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
	userID := jwtuser.GetUserId(c)

	if userID == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}

	cmd := &command.ClaimTaskCommand{
		TaskID: taskID,
		UserID: strconv.Itoa(userID),
	}

	if err := h.taskService.ClaimTask(c.Request.Context(), cmd); err != nil {
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
	userID := jwtuser.GetUserId(c)

	if userID == 0 {
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

	result := status.TaskResultCompleted
	if req.Result == "rejected" {
		result = status.TaskResultRejected
	}

	cmd := &command.CompleteTaskCommand{
		TaskID:  taskID,
		UserID:  strconv.Itoa(userID),
		Output:  req.Output,
		Comment: req.Comment,
		Result:  result,
	}

	if err := h.taskService.CompleteTask(c.Request.Context(), cmd); err != nil {
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
	userID := jwtuser.GetUserId(c)

	if userID == 0 {
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
		UserID:  strconv.Itoa(userID),
		Output:  req.Output,
		Comment: req.Comment,
		Result:  status.TaskResultCompleted,
	}

	if err := h.taskService.CompleteTask(c.Request.Context(), cmd); err != nil {
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
	userID := jwtuser.GetUserId(c)

	if userID == 0 {
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
		UserID:  strconv.Itoa(userID),
		Output:  req.Output,
		Comment: req.Comment,
		Result:  status.TaskResultRejected,
	}

	if err := h.taskService.CompleteTask(c.Request.Context(), cmd); err != nil {
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
	userID := jwtuser.GetUserId(c)

	if userID == 0 {
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
		UserID:   strconv.Itoa(userID),
		TargetID: req.TargetID,
		Comment:  req.Comment,
	}

	if err := h.taskService.DelegateTask(c.Request.Context(), cmd); err != nil {
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

	histories, err := h.taskService.GetTaskHistory(c.Request.Context(), taskID, limit, offset)
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

	histories, err := h.taskService.GetInstanceTaskHistory(c.Request.Context(), instanceID, limit, offset)
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

	tasks, total, err := h.taskService.GetInstanceTasks(c.Request.Context(), instanceID, limit, offset)
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

	if err := h.taskService.DeleteTask(c.Request.Context(), cmd); err != nil {
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
