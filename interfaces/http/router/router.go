package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jxt/process-management/interfaces/http/handler"
	"github.com/jxt/process-management/interfaces/http/middleware"
)

// SetupRoutes 设置路由
func SetupRoutes(
	engine *gin.Engine,
	workflowHandler *handler.WorkflowHandler,
	instanceHandler *handler.InstanceHandler,
	taskHandler *handler.TaskHandler,
) {
	// 工作流路由
	workflows := engine.Group("/api/workflows")
	{
		workflows.POST("", workflowHandler.CreateWorkflow)
		workflows.GET("", workflowHandler.ListWorkflows)
		workflows.GET("/:id", workflowHandler.GetWorkflow)
		workflows.PUT("/:id", workflowHandler.UpdateWorkflow)
		workflows.DELETE("/:id", workflowHandler.DeleteWorkflow)
		workflows.POST("/:id/activate", workflowHandler.ActivateWorkflow)
		workflows.POST("/:id/freeze", workflowHandler.FreezeWorkflow)
		workflows.GET("/:id/can-freeze", workflowHandler.CheckCanFreeze)
	}

	// 工作流实例路由
	instances := engine.Group("/api/instances")
	{
		instances.GET("", instanceHandler.ListAllInstances)
		instances.POST("", instanceHandler.StartInstance)
		instances.GET("/:id", instanceHandler.GetInstance)
		instances.DELETE("/:id", instanceHandler.DeleteInstance)
		instances.GET("/workflow/:workflow_id", instanceHandler.ListInstances)
	}

	// 任务路由 - 需要认证
	tasks := engine.Group("/api/tasks")
	tasks.Use(middleware.AuthMiddleware())
	{
		tasks.POST("", taskHandler.CreateTask)                                          // 创建任务
		tasks.GET("", taskHandler.ListAllTasks)                                         // 查询所有任务
		tasks.GET("/todo", taskHandler.ListTodoTasks)                                   // 我的待办
		tasks.GET("/done", taskHandler.ListDoneTasks)                                   // 我的已办
		tasks.GET("/claimable", taskHandler.ListClaimableTasks)                         // 待领任务
		tasks.GET("/:id", taskHandler.GetTask)                                          // 任务详情
		tasks.POST("/:id/claim", taskHandler.ClaimTask)                                 // 认领任务
		tasks.POST("/:id/complete", taskHandler.CompleteTask)                           // 完成任务
		tasks.POST("/:id/approve", taskHandler.ApproveTask)                             // 批准任务
		tasks.POST("/:id/reject", taskHandler.RejectTask)                               // 驳回任务
		tasks.POST("/:id/delegate", taskHandler.DelegateTask)                           // 转办任务
		tasks.DELETE("/:id", taskHandler.DeleteTask)                                    // 删除任务
		tasks.GET("/:id/history", taskHandler.GetTaskHistory)                           // 任务历史
		tasks.GET("/instance/:instance_id/history", taskHandler.GetInstanceTaskHistory) // 实例任务历史
		tasks.GET("/instance/:instance_id", taskHandler.GetInstanceTasks)               // 实例所有任务
	}

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "ok",
		})
	})
}
