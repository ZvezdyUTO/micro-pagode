package api

import (
	"mp/internal/infra/collector"
	"mp/internal/logic"
	"mp/internal/svc"
)

// initHandler 实例化 Logic 层和 Handler 层，执行路由分发与依赖装配。
func initHandler(svc *svc.ServiceContext) []Handler {
	// 实例化 Logic，拆分一类并且使用构造函数注入，使每个模块值依赖自身接口，并且保持边界清晰。
	var (
		userLogic = logic.NewUser(svc.UsersModel)
		fileLogic = logic.NewFile(
			svc.FileRepo, svc.Config.FileBasePath,
			svc.FileSearch,
		)
		systemLogic = logic.NewSystem(
			svc.Monitor,
			svc.SystemMonitorConfig,
			svc.SystemMonitorWarning,
			collector.NewGopsutilCollector(),
		)
	)

	// 实例化 Handler，将创建好的 Logic 实例注入
	var (
		userSelf   = NewUserSelf(svc, userLogic)
		adminUser  = NewAdminUser(svc, userLogic)
		file       = NewFile(svc, fileLogic)
		system     = NewSystem(svc, systemLogic)
		userPublic = NewUserPublic(svc, userLogic)
	)

	// 将所有实例化的 Handler 放入切片中返回
	// 这样组装，明确了项目中各个组件之间的依赖关系，通过模块化管理，将功能进行拆分，且为上层提供了一个整齐的处理器列表
	return []Handler{
		userSelf,
		adminUser,
		file,
		system,
		userPublic,
	}
}
