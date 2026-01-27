package api

import (
	application "jxt-evidence-system/process-management/internal/application/service"
	domain_service "jxt-evidence-system/process-management/internal/domain/service"
	persistence "jxt-evidence-system/process-management/internal/infrastructure/persistence/gorm"
	infra_ws "jxt-evidence-system/process-management/internal/infrastructure/websocket"
	"jxt-evidence-system/process-management/internal/interfaces/rest/api"
	"jxt-evidence-system/process-management/internal/interfaces/rest/router"
)

// jiyuanjie 注意import目录为正确的六边形架构的目录

func init() {
	//注册路由 fixme 其他应用的路由，在本目录新建文件放在init方法
	AppRouters = append(AppRouters, router.InitRouter)

	// jiyuanjie 添加依赖注入
	Registrations = append(Registrations, infra_ws.RegisterDependencies)
	Registrations = append(Registrations, persistence.RegisterDependencies)
	Registrations = append(Registrations, domain_service.RegisterDependencies)
	Registrations = append(Registrations, application.RegisterDependencies)
	Registrations = append(Registrations, api.RegisterDependencies)

}
