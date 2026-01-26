package router

import (
	"jxt-evidence-system/process-management/internal/interfaces/rest/api"
	"jxt-evidence-system/process-management/shared/common/di"
	"jxt-evidence-system/process-management/shared/common/middleware"
	"log"

	jwt "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/gin-gonic/gin"
)

func init() {
	routerCheckRole = append(routerCheckRole,
		registerWorkflowRouter,
		registerInstanceRouter,
		registerTaskRouter,
	)
}

func registerWorkflowRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 通过依赖注入创建API处理器
	err := di.Invoke(func(handler *api.WorkflowHandler) {
		if handler != nil {
			r := v1.Group("/workflows").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
			{
				r.POST("", handler.CreateWorkflow)
				r.GET("", handler.GetPage)
				r.GET("/all", handler.GetAllWorkflow)
				r.GET("/:id", handler.GetWorkflow)
				r.GET("/name/:name", handler.GetWorkflowByName)
				r.PUT("/:id", handler.UpdateWorkflow)
				r.DELETE("/:id", handler.DeleteWorkflow)
				r.POST("/:id/activate", handler.ActivateWorkflow)
				r.POST("/:id/freeze", handler.FreezeWorkflow)
				r.GET("/:id/can-freeze", handler.CheckCanFreeze)
			}
		} else {
			logger.Fatal("WorkflowHandler is nil after resolution")
		}
	})

	if err != nil {
		logger.Fatalf("Failed to resolve WorkflowHandler: %v", err)
	}
}

func registerInstanceRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 通过依赖注入创建API处理器
	err := di.Invoke(func(handler *api.InstanceHandler) {
		if handler != nil {
			r := v1.Group("/instances").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
			{
				r.GET("", handler.GetPage)
				r.POST("", handler.StartInstance)
				r.GET("/:id", handler.GetInstance)
				r.GET("/:id/cancel", handler.CancelInstance)
				r.GET("/:id/detail", handler.GetInstanceDetail)
				r.DELETE("/:id", handler.DeleteInstance)
				r.GET("/workflow/:workflow_id", handler.GetInstancesByWorkflow)
			}
		} else {
			logger.Fatal("InstanceHandler is nil after resolution")
		}
	})

	if err != nil {
		logger.Fatalf("Failed to resolve InstanceHandler: %v", err)
	}
}

func registerTaskRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 通过依赖注入创建API处理器
	err := di.Invoke(func(handler *api.TaskHandler) {
		if handler != nil {
			r := v1.Group("/tasks").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
			{
				r.POST("", handler.CreateTask)                                         // 创建任务
				r.GET("", handler.GetPage)                                             // 查询所有任务
				r.GET("/todo", handler.GetTodoTasks)                                   // 我的待办
				r.GET("/done", handler.GetDoneTasks)                                   // 我的已办
				r.GET("/:id", handler.GetTask)                                         // 任务详情
				r.POST("/:id/complete", handler.CompleteTask)                          // 完成任务
				r.POST("/:id/approve", handler.ApproveTask)                            // 批准任务
				r.POST("/:id/reject", handler.RejectTask)                              // 驳回任务
				r.POST("/:id/delegate", handler.DelegateTask)                          // 转办任务
				r.DELETE("/:id", handler.DeleteTask)                                   // 删除任务
				r.GET("/:id/history", handler.GetTaskHistory)                          // 任务历史
				r.GET("/instance/:instanceId/recent", handler.GetRecentTask)           // 实例最近任务
				r.GET("/instance/:instanceId/history", handler.GetInstanceTaskHistory) // 实例任务历史
				r.GET("/instance/:instanceId", handler.GetTasksByInstanceID)           // 实例所有任务
			}
		} else {
			logger.Fatal("TaskHandler is nil after resolution")
		}
	})

	if err != nil {
		logger.Fatalf("Failed to resolve TaskHandler: %v", err)
	}
}

func registerGinWebSocketRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	// 通过依赖注入创建API处理器
	err := di.Invoke(func(handler *api.WebSocketHandler) {
		if handler != nil {
			r := v1.Group("/instances").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
			{
				// 🆕 添加WebSocket路由
				r.GET("/ws", handler.HandleWebSocket)
				r.GET("/api/ws/online-users", handler.GetOnlineUsers)
				r.GET("/api/ws/user/:user_id/online", handler.CheckUserOnline)
				r.POST("/api/ws/test-message", handler.SendTestMessage)
				log.Println("WebSocket routes registered")
			}
		} else {
			logger.Fatal("InstanceHandler is nil after resolution")
		}
	})

	if err != nil {
		logger.Fatalf("Failed to resolve InstanceHandler: %v", err)
	}
}
