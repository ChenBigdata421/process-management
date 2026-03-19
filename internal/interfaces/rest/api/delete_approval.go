package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service"
	"jxt-evidence-system/process-management/internal/application/service/port"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/restapi"

	jwtuser "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeleteApprovalHandler 删除审批HTTP处理器
type DeleteApprovalHandler struct {
	restapi.RestApi
	downloadApprovalService *service.DeleteApprovalService
	instanceService         port.InstanceService
	workflowService         port.WorkflowService
}

// NewDeleteApprovalHandler 创建删除审批处理器
func NewDeleteApprovalHandler(
	downloadApprovalService *service.DeleteApprovalService,
	instanceService port.InstanceService,
	workflowService port.WorkflowService,
) *DeleteApprovalHandler {
	return &DeleteApprovalHandler{
		downloadApprovalService: downloadApprovalService,
		instanceService:         instanceService,
		workflowService:         workflowService,
	}
}

// GetDeleteApprovalStatus 获取删除审批状态
// GET /api/v1/media/:mediaId/delete-approval
func (h *DeleteApprovalHandler) GetDeleteApprovalStatus(c *gin.Context) {
	// c.Request.Context() 已被租户中间件设置了租户ID，直接使用即可
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var cmd command.GetDeleteApprovalStatusCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}

	// 获取当前用户ID
	userID := int64(jwtuser.GetUserId(c))
	if userID == 0 {
		h.Error(c, http.StatusUnauthorized, nil, "未授权")
		return
	}

	// 查询审批状态
	status, err := h.downloadApprovalService.GetApprovalStatus(ctx, cmd.MediaID, userID)
	if err != nil {
		logger.Error("查询删除审批状态失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "查询删除审批状态失败")
		return
	}

	h.OK(c, status, "查询成功")
}

// SubmitDeleteApproval 提交删除审批申请
// POST /api/v1/media/:mediaId/delete-approval
func (h *DeleteApprovalHandler) SubmitDeleteApproval(c *gin.Context) {
	var cmd command.SubmitDeleteApprovalCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		h.GetLogger(c).Error("bind SubmitDeleteApprovalCommand err", zap.Error(err))
		h.Error(c, http.StatusBadRequest, err, fmt.Sprintf("提交删除审批申请失败，参数验证失败: %s", err.Error()))
		return
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.GetLogger(c).Error("bind SubmitDeleteApprovalCommand err", zap.Error(err))
		h.Error(c, http.StatusBadRequest, err, fmt.Sprintf("提交删除审批申请失败，参数验证失败: %s", err.Error()))
		return
	}

	// 获取当前用户ID
	userID := int64(jwtuser.GetUserId(c))
	if userID == 0 {
		h.Error(c, http.StatusUnauthorized, nil, "未授权")
		return
	}

	// 先设置UserID到context（基于c.Request.Context()，它已包含租户ID）
	baseCtx := context.WithValue(c.Request.Context(), global.UserIDKey, int(userID))

	// 再创建超时context（继承租户ID和UserID）
	ctx, cancel := context.WithTimeout(baseCtx, 10*time.Second)
	defer cancel()

	// 提交审批申请
	_, err := h.downloadApprovalService.SubmitApproval(ctx, userID, &cmd)
	if err != nil {
		logger.Error("提交删除审批申请失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "提交删除审批申请失败")
		return
	}

	h.OK(c, "", "提交成功，请等待审批")
}
