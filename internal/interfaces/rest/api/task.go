package api

import (
	"context"
	"net/http"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/status"

	"jxt-evidence-system/process-management/shared/common/restapi"

	jwtuser "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
)

// TaskHandler 任务HTTP处理器
type TaskHandler struct {
	restapi.RestApi
	taskService port.TaskService
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	cmd := command.CreateTaskCommand{}

	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.Error(c, http.StatusBadRequest, err, "参数验证失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	id, err := h.taskService.CreateTask(ctx, &cmd)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, err, "创建任务失败")
		return
	}

	h.OK(c, gin.H{"id": id}, "创建任务成功")
}

// GetPage 列出所有任务（支持筛选）
func (h *TaskHandler) GetPage(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	var query command.TaskPagedQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.GetLogger(c).Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	tasks, total, err := h.taskService.GetPage(ctx, &query)
	if err != nil {
		logger.Error("查询任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "查询任务失败")
		return
	}

	h.PageOK(c, tasks, int(total), query.GetPageIndex(), query.GetPageSize(), "查询成功")
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var cmd command.GetTaskByIDCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	task, err := h.taskService.GetTaskByID(ctx, cmd.ID)
	if err != nil {
		logger.Error("获取任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "获取任务失败")
		return
	}

	h.OK(c, task, "获取任务成功")
}

func (h *TaskHandler) GetRecentTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	var cmd command.GetRecentTaskCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取最近任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	task, err := h.taskService.GetRecentTask(ctx, cmd.ID)
	if err != nil {
		logger.Error("获取最近任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "获取最近任务失败")
		return
	}

	h.OK(c, task, "获取最近任务成功")
}

// GetTodoTasks 查询待办任务
func (h *TaskHandler) GetTodoTasks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	userID := jwtuser.GetUserId(c)
	if userID == 0 {
		logger.Error("获取用户ID失败")
		h.Error(c, http.StatusUnauthorized, nil, "获取当前用户ID失败")
		return
	}

	var query command.TodoTaskPagedQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.GetLogger(c).Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	tasks, total, err := h.taskService.GetTodoTasks(ctx, userID, &query)
	if err != nil {
		logger.Error("查询待办任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "查询待办任务失败")
		return
	}

	h.PageOK(c, tasks, int(total), query.GetPageIndex(), query.GetPageSize(), "查询成功")
}

// GetDoneTasks 查询已办任务
func (h *TaskHandler) GetDoneTasks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	userID := jwtuser.GetUserId(c)
	if userID == 0 {
		logger.Error("获取用户ID失败")
		h.Error(c, http.StatusUnauthorized, nil, "获取当前用户ID失败")
		return
	}

	var query command.DoneTaskPagedQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		h.GetLogger(c).Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}

	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	tasks, total, err := h.taskService.GetDoneTasks(ctx, userID, &query)
	if err != nil {
		logger.Error("查询已办任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "查询已办任务失败")
		return
	}

	h.PageOK(c, tasks, int(total), query.GetPageIndex(), query.GetPageSize(), "查询成功")
}

// CompleteTask 完成任务
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	userID := jwtuser.GetUserId(c)
	if userID == 0 {
		logger.Error("获取用户ID失败")
		h.Error(c, http.StatusInternalServerError, nil, "完成任务失败")
		return
	}
	var cmd command.CompleteTaskCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定完成任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logger.Error("绑定完成任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	cmd.UserID = int(userID)

	result := status.TaskResultCompleted
	if cmd.Result == "rejected" {
		result = status.TaskResultRejected
	}
	cmd.Result = result

	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.taskService.CompleteTask(ctx, &cmd); err != nil {
		logger.Error("完成任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "完成任务失败")
		return
	}

	h.OK(c, nil, "完成任务成功")
}

// ApproveTask 批准任务
func (h *TaskHandler) ApproveTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	userID := jwtuser.GetUserId(c)
	if userID == 0 {
		logger.Error("获取用户ID失败")
		h.Error(c, http.StatusInternalServerError, nil, "批准任务失败")
		return
	}

	var cmd command.CompleteTaskCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定批准任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logger.Error("绑定批准任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	cmd.UserID = int(userID)
	cmd.Result = status.TaskResultApproved
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	ctx = context.WithValue(ctx, global.UserIDKey, int(userID))
	if err := h.taskService.CompleteTask(ctx, &cmd); err != nil {
		logger.Error("批准任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "批准任务失败")
		return
	}

	h.OK(c, nil, "批准任务成功")
}

// RejectTask 驳回任务
func (h *TaskHandler) RejectTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	userID := jwtuser.GetUserId(c)
	if userID == 0 {
		logger.Error("获取用户ID失败")
		h.Error(c, http.StatusInternalServerError, nil, "驳回任务失败")
		return
	}

	var cmd command.CompleteTaskCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定批准任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logger.Error("绑定批准任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	cmd.UserID = int(userID)
	cmd.Result = status.TaskResultRejected

	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	ctx = context.WithValue(ctx, global.UserIDKey, int(userID))
	if err := h.taskService.CompleteTask(ctx, &cmd); err != nil {
		logger.Error("驳回任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "驳回任务失败")
		return
	}

	h.OK(c, nil, "驳回任务成功")
}

// DelegateTask 转办任务
func (h *TaskHandler) DelegateTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	userID := jwtuser.GetUserId(c)
	if userID == 0 {
		logger.Error("获取用户ID失败")
		h.Error(c, http.StatusInternalServerError, nil, "转办任务失败")
		return
	}

	var cmd command.DelegateTaskCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定转办任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logger.Error("绑定转办任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	cmd.UserID = int(userID)

	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.taskService.DelegateTask(ctx, &cmd); err != nil {
		logger.Error("转办任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "转办任务失败")
		return
	}

	h.OK(c, nil, "转办任务成功")
}

// GetTaskHistory 获取任务历史
func (h *TaskHandler) GetTaskHistory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var cmd command.TaskHistory
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取任务历史命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	histories, err := h.taskService.GetTaskHistory(ctx, cmd.ID)
	if err != nil {
		logger.Error("获取任务历史失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "获取任务历史失败")
		return
	}

	h.OK(c, histories, "获取任务历史成功")
}

// GetInstanceTaskHistory 获取实例的任务历史
func (h *TaskHandler) GetInstanceTaskHistory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var cmd command.GetTasksByInstanceID
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取实例任务历史命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	if err := c.ShouldBindQuery(&cmd); err != nil {
		h.GetLogger(c).Error(err.Error())
		h.Error(c, http.StatusBadRequest, err, "请求参数绑定失败")
		return
	}

	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	histories, err := h.taskService.GetInstanceTaskHistory(ctx, cmd.ID)
	if err != nil {
		logger.Error("获取实例任务历史失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "获取实例任务历史失败")
		return
	}

	h.OK(c, histories, "获取实例任务历史成功")
}

// GetInstanceTasks 获取实例的所有任务（包含当前状态）
func (h *TaskHandler) GetTasksByInstanceID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var cmd command.GetTasksByInstanceID
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定获取实例任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*")
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	tasks, err := h.taskService.GetTasksByInstanceID(ctx, cmd.ID)
	if err != nil {
		logger.Error("获取实例任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "获取实例任务失败")
		return
	}

	h.OK(c, tasks, "查询成功")
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	var cmd command.DeleteTaskCommand
	if err := c.ShouldBindUri(&cmd); err != nil {
		logger.Error("绑定删除任务命令的参数失败", "error", err)
		h.Error(c, http.StatusBadRequest, err, "请求参数错误")
		return
	}
	// 设置租户ID（单租户模式使用默认租户 "*"）
	ctx = context.WithValue(ctx, global.TenantIDKey, "*")
	if err := h.taskService.DeleteTask(ctx, &cmd); err != nil {
		logger.Error("删除任务失败", "error", err)
		h.Error(c, http.StatusInternalServerError, err, "删除任务失败")
		return
	}

	h.OK(c, nil, "删除任务成功")
}
