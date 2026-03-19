package api

import (
	"context"
	"net/http"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
	sharederrors "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/restapi"

	jwtuser "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
)

// InstanceHandler 工作流实例HTTP处理器
type InstanceHandler struct {
	restapi.RestApi
	instanceService port.InstanceService
}

// CancelInstance 取消工作流实例（将状态标记为取消）
func (h *InstanceHandler) CancelInstance(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var cmd command.CancelInstanceCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定取消工作流实例命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}

	if err := h.instanceService.CancelInstance(ctx, &cmd); err != nil {
		logger.Error("取消工作流实例失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "取消工作流实例失败")
		return
	}

	h.OK(c, nil, "取消工作流实例成功")
}

// StartInstance 启动工作流实例
func (h *InstanceHandler) StartInstance(c *gin.Context) {
	cmd := command.StartWorkflowInstanceCommand{}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "")
		return
	}

	// 先设置UserID到context（基于c.Request.Context()，它已包含租户ID）
	userID := jwtuser.GetUserId(c)
	baseCtx := context.WithValue(c.Request.Context(), global.UserIDKey, int(userID))

	// 再创建超时context（继承租户ID和UserID）
	ctx, cancel := context.WithTimeout(baseCtx, 2*time.Second)
	defer cancel()

	id, err := h.instanceService.StartWorkflowInstance(ctx, &cmd)
	if err != nil {
		// 根据错误类型判断返回的 HTTP 状态码
		switch err {
		case sharederrors.ErrWorkflowNotFound,
			 sharederrors.ErrInvalidStatusTransition,
			 sharederrors.ErrInvalidWorkflowDefinition:
			h.Error(c, http.StatusBadRequest, err, "启动工作流实例失败")
		default:
			h.Error(c, http.StatusInternalServerError, err, "启动工作流实例失败")
		}
		return
	}

	h.OK(c, gin.H{"id": id}, "启动工作流实例成功")
}

// GetInstance 获取工作流实例
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var cmd command.GetInstanceCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取工作流实例命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	dto, err := h.instanceService.GetInstanceByID(ctx, cmd.ID)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "获取工作流实例失败")
		return
	}
	h.OK(c, dto, "获取工作流实例成功")
}

func (h *InstanceHandler) GetInstanceDetail(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var cmd command.GetInstanceCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取工作流实例详情命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	dto, err := h.instanceService.GetInstanceDetailByID(ctx, cmd.ID)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "获取工作流实例详情失败")
		return
	}
	h.OK(c, dto, "获取工作流实例详情成功")
}

// ListInstances 列出工作流实例
func (h *InstanceHandler) GetInstancesByWorkflow(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var query command.GetInstancesByWorkflowPagedQuery
	if err := c.ShouldBindUri(&query); err != nil {
		logger.Error("绑定获取工作流实例命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	instances, total, err := h.instanceService.GetInstancesByWorkflow(ctx, &query)
	if err != nil {

		h.Error(c, http.StatusInternalServerError, err, "查询工作流实例失败")
		return
	}

	h.PageOK(c, instances, int(total), query.GetPageIndex(), query.GetPageSize(), "查询成功")
}

// GetPage 列出所有工作流实例（支持筛选）
func (h *InstanceHandler) GetPage(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	var query command.InstancePagedQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.GetLogger(c).Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	instances, total, err := h.instanceService.GetPage(ctx, &query)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "查询工作流实例失败")
		return
	}

	h.PageOK(c, instances, int(total), query.GetPageIndex(), query.GetPageSize(), "查询成功")
}

// DeleteInstance 删除工作流实例
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var cmd command.DeleteInstanceCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定删除工作流实例命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	if cmd.GetId().IsEmpty() {
		logger.Error("绑定删除工作流实例命令的参数失败,实例ID不能为空")
		h.Error(c, http.StatusBadRequest, nil, "请求参数错误")
		return
	}
	if err := h.instanceService.DeleteInstance(ctx, &cmd); err != nil {
		logger.Error("删除工作流实例失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "删除工作流实例失败")
		return
	}

	h.OK(c, nil, "删除工作流实例成功")
}
