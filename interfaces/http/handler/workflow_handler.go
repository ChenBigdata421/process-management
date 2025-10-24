package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/domain/workflow"
)

// WorkflowHandler 工作流HTTP处理器
type WorkflowHandler struct {
	createHandler   *command.CreateWorkflowHandler
	updateHandler   *command.UpdateWorkflowHandler
	deleteHandler   *command.DeleteWorkflowHandler
	activateHandler *command.ActivateWorkflowHandler
	freezeHandler   *command.FreezeWorkflowHandler
	queryService    *query.WorkflowQueryService
	instanceRepo    workflow.WorkflowInstanceRepository
}

// NewWorkflowHandler 创建处理器
func NewWorkflowHandler(
	createHandler *command.CreateWorkflowHandler,
	updateHandler *command.UpdateWorkflowHandler,
	deleteHandler *command.DeleteWorkflowHandler,
	activateHandler *command.ActivateWorkflowHandler,
	freezeHandler *command.FreezeWorkflowHandler,
	queryService *query.WorkflowQueryService,
	instanceRepo workflow.WorkflowInstanceRepository,
) *WorkflowHandler {
	return &WorkflowHandler{
		createHandler:   createHandler,
		updateHandler:   updateHandler,
		deleteHandler:   deleteHandler,
		activateHandler: activateHandler,
		freezeHandler:   freezeHandler,
		queryService:    queryService,
		instanceRepo:    instanceRepo,
	}
}

// CreateWorkflow 创建工作流
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Definition  string `json:"definition" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	cmd := &command.CreateWorkflowCommand{
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.Definition,
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

// GetWorkflow 获取工作流
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")

	dto, err := h.queryService.GetWorkflowByID(c.Request.Context(), id)
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
		"data": dto,
	})
}

// ListWorkflows 列出工作流
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
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

	dtos, err := h.queryService.ListWorkflows(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	count, err := h.queryService.CountWorkflows(c.Request.Context())
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
		"data": gin.H{
			"items": dtos,
			"total": count,
		},
	})
}

// UpdateWorkflow 更新工作流
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Definition  string `json:"definition" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	cmd := &command.UpdateWorkflowCommand{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.Definition,
	}

	if err := h.updateHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// DeleteWorkflow 删除工作流
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")

	cmd := &command.DeleteWorkflowCommand{ID: id}

	if err := h.deleteHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// ActivateWorkflow 激活工作流
func (h *WorkflowHandler) ActivateWorkflow(c *gin.Context) {
	id := c.Param("id")

	cmd := &command.ActivateWorkflowCommand{ID: id}

	if err := h.activateHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// FreezeWorkflow 冻结工作流
func (h *WorkflowHandler) FreezeWorkflow(c *gin.Context) {
	id := c.Param("id")

	cmd := &command.FreezeWorkflowCommand{ID: id}

	if err := h.freezeHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})
}

// CheckCanFreeze 检查工作流是否可以冻结
func (h *WorkflowHandler) CheckCanFreeze(c *gin.Context) {
	id := c.Param("id")

	// 查询该工作流的所有实例
	instances, err := h.instanceRepo.FindByWorkflowID(c.Request.Context(), id, 10000, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "查询实例失败: " + err.Error(),
		})
		return
	}

	// 统计运行中和已完成的实例数量
	runningCount := 0
	completedCount := 0
	canFreeze := true
	reason := ""

	for _, inst := range instances {
		if inst.Status == workflow.InstanceStatusCompleted {
			completedCount++
		} else {
			runningCount++
			canFreeze = false
		}
	}

	if !canFreeze {
		reason = "存在未完成的实例"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"can_freeze":          canFreeze,
			"running_instances":   runningCount,
			"completed_instances": completedCount,
			"reason":              reason,
		},
	})
}
