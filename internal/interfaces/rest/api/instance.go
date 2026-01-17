package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"

	"github.com/gin-gonic/gin"
)

// InstanceHandler 工作流实例HTTP处理器
type InstanceHandler struct {
	instanceService port.InstanceService
}

// StartInstance 启动工作流实例
func (h *InstanceHandler) StartInstance(c *gin.Context) {
	var req struct {
		WorkflowID string          `json:"workflow_id" binding:"required"`
		Input      json.RawMessage `json:"input"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 将 JSON 对象转换为字符串，并移除多余的空格和换行符
	inputStr := string(req.Input)
	if inputStr == "" || inputStr == "null" {
		inputStr = "{}"
	} else {
		// 解析 JSON 并重新编码，以移除多余的空格和换行符
		var jsonData interface{}
		if err := json.Unmarshal([]byte(inputStr), &jsonData); err == nil {
			// 重新编码为紧凑的 JSON 字符串
			if compactJSON, err := json.Marshal(jsonData); err == nil {
				inputStr = string(compactJSON)
			}
		}
	}

	cmd := &command.StartWorkflowInstanceCommand{
		WorkflowID: req.WorkflowID,
		Input:      inputStr,
	}

	id, err := h.instanceService.StartWorkflowInstance(c.Request.Context(), cmd)
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

// GetInstance 获取工作流实例
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")

	dto, err := h.instanceService.GetInstanceByID(c.Request.Context(), id)
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

// ListInstances 列出工作流实例
func (h *InstanceHandler) ListInstances(c *gin.Context) {
	workflowID := c.Param("workflow_id")
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

	dtos, err := h.instanceService.ListInstancesByWorkflowID(c.Request.Context(), workflowID, limit, offset)
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
		"data": dtos,
	})
}

// ListAllInstances 列出所有工作流实例（支持筛选）
func (h *InstanceHandler) ListAllInstances(c *gin.Context) {
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
	if workflowID := c.Query("workflow_id"); workflowID != "" {
		filters["workflow_id"] = workflowID
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	dtos, total, err := h.instanceService.ListAllInstances(c.Request.Context(), filters, limit, offset)
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
			"total": total,
		},
	})
}

// DeleteInstance 删除工作流实例
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	id := c.Param("id")

	cmd := &command.DeleteInstanceCommand{
		ID: id,
	}

	if err := h.instanceService.DeleteInstance(c.Request.Context(), cmd); err != nil {
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
