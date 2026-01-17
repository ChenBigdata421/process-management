package router

import (
	"os"

	common "jxt-evidence-system/process-management/shared/common/middleware"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/gin-gonic/gin"
)

// InitRouter 路由初始化，不要怀疑，这里用到了
func InitRouter() {
	println("🔧 [DEBUG] InitRouter() 被调用")
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		logger.Fatal("not found engine...")
		os.Exit(-1)
	}
	switch engine := h.(type) {
	case *gin.Engine:
		r = engine
	default:
		logger.Fatal("not support other engine")
		os.Exit(-1)
	}
	// the jwt middleware
	authMiddleware, err := common.AuthInit()
	if err != nil {
		logger.Fatalf("JWT Init Error, %s", err.Error())
	}

	// 注册业务路由
	// TODO: 这里可存放业务路由，里边并无实际路由只有演示代码
	println("🔧 [DEBUG] 开始调用 initRouter")
	initRouter(r, authMiddleware)
	println("🔧 [DEBUG] initRouter 调用完成")
}
