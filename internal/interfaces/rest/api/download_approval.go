package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service"
	"jxt-evidence-system/process-management/internal/application/service/port"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/restapi"

	jwtuser "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DownloadApprovalHandler 下载审批HTTP处理器
type DownloadApprovalHandler struct {
	restapi.RestApi
	downloadApprovalService *service.DownloadApprovalService
	instanceService         port.InstanceService
	workflowService         port.WorkflowService
}

// NewDownloadApprovalHandler 创建下载审批处理器
func NewDownloadApprovalHandler(
	downloadApprovalService *service.DownloadApprovalService,
	instanceService port.InstanceService,
	workflowService port.WorkflowService,
) *DownloadApprovalHandler {
	return &DownloadApprovalHandler{
		downloadApprovalService: downloadApprovalService,
		instanceService:         instanceService,
		workflowService:         workflowService,
	}
}

// GetDownloadApprovalStatus 获取下载审批状态
// GET /api/v1/media/:mediaId/download-approval
func (h *DownloadApprovalHandler) GetDownloadApprovalStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var cmd command.GetDownloadApprovalStatusCommand
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

	// 设置租户ID
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")

	// 查询审批状态
	status, err := h.downloadApprovalService.GetApprovalStatus(ctx, cmd.MediaID, userID)
	if err != nil {
		logger.Error("查询下载审批状态失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "查询下载审批状态失败")
		return
	}

	h.OK(c, status, "查询成功")
}

// SubmitDownloadApproval 提交下载审批申请
// POST /api/v1/media/:mediaId/download-approval
func (h *DownloadApprovalHandler) SubmitDownloadApproval(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var cmd command.SubmitDownloadApprovalCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		h.GetLogger(c).Error("bind SubmitDownloadApprovalCommand err", zap.Error(err))
		h.Error(c, http.StatusBadRequest, err, fmt.Sprintf("提交下载审批申请失败，参数验证失败: %s", err.Error()))
		return
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.GetLogger(c).Error("bind SubmitDownloadApprovalCommand err", zap.Error(err))
		h.Error(c, http.StatusBadRequest, err, fmt.Sprintf("提交下载审批申请失败，参数验证失败: %s", err.Error()))
		return
	}

	// 获取当前用户ID
	userID := int64(jwtuser.GetUserId(c))
	if userID == 0 {
		h.Error(c, http.StatusUnauthorized, nil, "未授权")
		return
	}

	// 设置租户ID和用户ID
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	ctx = context.WithValue(ctx, global.UserIDKey, int(userID))

	// 提交审批申请
	_, err := h.downloadApprovalService.SubmitApproval(ctx, userID, &cmd)
	if err != nil {
		logger.Error("提交下载审批申请失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "提交下载审批申请失败")
		return
	}

	h.OK(c, "", "提交成功，请等待审批")
}

// BatchGetDownloadApprovalStatus 批量查询下载审批状态
// POST /api/v1/media/download-approval/batch
func (h *DownloadApprovalHandler) BatchGetDownloadApprovalStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 获取当前用户ID
	userID := int64(jwtuser.GetUserId(c))
	if userID == 0 {
		h.Error(c, http.StatusUnauthorized, nil, "未授权")
		return
	}

	// 解析请求体
	var cmd command.BatchGetDownloadApprovalStatusCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}

	if len(cmd.MediaIDs) == 0 {
		h.Error(c, http.StatusBadRequest, nil, "媒体ID列表不能为空")
		return
	}

	// 设置租户ID
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")

	// 批量查询
	results, err := h.downloadApprovalService.BatchGetApprovalStatus(ctx, cmd.MediaIDs, userID)
	if err != nil {
		logger.Error("批量查询下载审批状态失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "批量查询下载审批状态失败")
		return
	}

	h.OK(c, results, "查询成功")
}

// RecordDownload 记录下载（审批通过后调用）
// POST /api/v1/media/:mediaId/download-record
func (h *DownloadApprovalHandler) RecordDownload(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var cmd command.RecordDownloadCommand
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

	// 设置租户ID
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")

	// 记录下载
	if err := h.downloadApprovalService.RecordDownload(ctx, cmd.MediaID, userID); err != nil {
		logger.Error("记录下载失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "记录下载失败")
		return
	}

	h.OK(c, nil, "记录成功")
}

const mediaDownloadWorkflowName = "媒体下载申请流程"

// startDownloadApprovalWorkflow 启动下载审批工作流
func (h *DownloadApprovalHandler) startDownloadApprovalWorkflow(ctx context.Context, mediaID valueobject.MediaID, userID int64, reason string) (valueobject.InstanceID, error) {
	if h.workflowService == nil {
		return valueobject.InstanceID{}, fmt.Errorf("workflow service not initialized")
	}

	workflow, err := h.workflowService.GetWorkflowByName(ctx, mediaDownloadWorkflowName)
	if err != nil {
		return valueobject.InstanceID{}, fmt.Errorf("failed to find workflow %s: %w", mediaDownloadWorkflowName, err)
	}

	payload := map[string]interface{}{
		"mediaId":     mediaID,
		"reason":      reason,
		"applicantId": userID,
	}
	input, err := json.Marshal(payload)
	if err != nil {
		return valueobject.InstanceID{}, fmt.Errorf("failed to marshal workflow input: %w", err)
	}

	cmd := command.StartWorkflowInstanceCommand{
		ID:    workflow.WorkflowID,
		Input: input,
	}

	instanceID, err := h.instanceService.StartWorkflowInstance(ctx, &cmd)
	if err != nil {
		return valueobject.InstanceID{}, fmt.Errorf("failed to start workflow instance: %w", err)
	}

	return instanceID, nil
}
