package api

import (
	"context"
	"net/http"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	"jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/status"

	"jxt-evidence-system/process-management/shared/common/restapi"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WorkflowHandler 工作流HTTP处理器
type WorkflowHandler struct {
	restapi.RestApi
	workflowService port.WorkflowService
	instanceService port.InstanceService
	instanceRepo    repository.WorkflowInstanceRepository
}

// CreateWorkflow 创建工作流
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	cmd := command.CreateWorkflowCommand{}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	cmd.SetCreateBy(user.GetUserId(c))
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	id, err := h.workflowService.CreateWorkflow(ctx, &cmd)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "创建工作流失败")
		return
	}
	h.OK(c, gin.H{"id": id}, "创建工作流成功")
}

// GetWorkflow 获取工作流
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	cmd := command.GetWorkflowByIDCommand{}
	if err := c.ShouldBindUri(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	workflow, err := h.workflowService.GetWorkflowByID(ctx, cmd.ID)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "获取工作流失败")
		return
	}

	h.OK(c, workflow, "获取工作流成功")
}

// GetWorkflow 获取工作流
func (h *WorkflowHandler) GetWorkflowByName(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	cmd := command.GetWorkflowByNameCommand{}
	if err := c.ShouldBindUri(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	workflow, err := h.workflowService.GetWorkflowByName(ctx, cmd.Name)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "获取工作流失败")
		return
	}

	h.OK(c, workflow, "获取工作流成功")
}

// GetPage 列出所有工作流（支持筛选）
func (h *WorkflowHandler) GetPage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	var query command.WorkflowPagedQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.GetLogger(c).Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	workflows, total, err := h.workflowService.GetPage(ctx, &query)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "查询工作流失败")
		return
	}

	h.PageOK(c, workflows, int(total), query.GetPageIndex(), query.GetPageSize(), "查询成功")
}

func (h *WorkflowHandler) GetAllWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	workflows, err := h.workflowService.GetAllWorkflow(ctx)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "查询工作流失败")
		return
	}

	h.OK(c, workflows, "查询成功")
}

// UpdateWorkflow 更新工作流
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	cmd := command.UpdateWorkflowCommand{}
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定更新工作流命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}

	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.GetLogger(c).Error("bind UpdateWorkflowCommand err", zap.Error(err))
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}

	cmd.SetUpdateBy(user.GetUserId(c))
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.workflowService.UpdateWorkflow(ctx, &cmd); err != nil {
		h.Error(c, http.StatusInternalServerError, err, "更新工作流失败")
		return
	}

	h.OK(c, nil, "更新工作流成功")
}

// DeleteWorkflow 删除工作流
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var cmd command.DeleteWorkflowCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定删除工作流命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}

	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.workflowService.DeleteWorkflow(ctx, &cmd); err != nil {
		logger.Error("删除工作流失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "删除工作流失败")
		return
	}

	h.OK(c, nil, "删除工作流成功")
}

// ActivateWorkflow 激活工作流
func (h *WorkflowHandler) ActivateWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	cmd := &command.ActivateWorkflowCommand{}
	err := c.ShouldBindUri(&cmd)
	log := h.GetLogger(c)
	if err != nil {
		log.Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.workflowService.ActivateWorkflow(ctx, cmd); err != nil {
		h.Error(c, 500, err, "激活工作流失败")
		return
	}

	h.OK(c, nil, "激活工作流成功")
}

// FreezeWorkflow 冻结工作流
func (h *WorkflowHandler) FreezeWorkflow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	cmd := &command.FreezeWorkflowCommand{}
	err := c.ShouldBindUri(&cmd)
	log := h.GetLogger(c)
	if err != nil {
		log.Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.workflowService.FreezeWorkflow(ctx, cmd); err != nil {
		log.Error(err.Error())
		h.Error(c, http.StatusInternalServerError, err, "冻结工作流失败")
		return
	}

	h.OK(c, nil, "冻结工作流成功")
}

// CheckCanFreeze 检查工作流是否可以冻结
func (h *WorkflowHandler) CheckCanFreeze(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var cmd command.CheckCanFreezeCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定检查工作流可否冻结命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")

	// 1. 先调用 CountInstanceByWorkflow 统计总实例数
	totalCount, err := h.instanceService.CountInstanceByWorkflow(ctx, cmd.ID)
	if err != nil {
		logger.Error("统计工作流实例数量失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "统计实例数量失败")
		return
	}

	// 2. 循环调用 GetInstancesByWorkflow 获取所有实例并统计状态
	const pageSize = 100
	var allInstances []*instance_aggregate.WorkflowInstance

	for offset := 0; ; offset += pageSize {
		query := &command.GetInstancesByWorkflowPagedQuery{
			ID: cmd.ID,
		}
		query.PageIndex = (offset / pageSize) + 1
		query.PageSize = pageSize

		instances, total, err := h.instanceService.GetInstancesByWorkflow(ctx, query)
		if err != nil {
			logger.Error("查询工作流实例失败", "error", err)
			h.Error(c, http.StatusInternalServerError, err, "查询实例失败")
			return
		}

		allInstances = append(allInstances, instances...)

		// 如果已经获取了所有实例，退出循环
		if len(allInstances) >= total {
			break
		}
	}

	// 3. 统计各种状态的实例数量
	runningCount := 0
	completedCount := 0
	failedCount := 0
	cancelledCount := 0
	canFreeze := true
	reason := ""

	for _, inst := range allInstances {
		switch inst.Status {
		case status.InstanceStatusCompleted:
			completedCount++
		case status.InstanceStatusFailed:
			failedCount++
		case status.InstanceStatusCancelled:
			cancelledCount++
		default:
			// 运行中或其他未完成状态
			runningCount++
			canFreeze = false
		}
	}

	if !canFreeze {
		reason = "存在未完成的实例"
	}

	// 4. 调用 h.OK 输出结果
	h.OK(c, gin.H{
		"can_freeze":          canFreeze,
		"total_instances":     totalCount,
		"running_instances":   runningCount,
		"completed_instances": completedCount,
		"failed_instances":    failedCount,
		"cancelled_instances": cancelledCount,
		"reason":              reason,
	}, "检查工作流冻结状态成功")
}
