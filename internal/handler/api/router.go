package api

import (
	"mp/internal/infra/collector"
	"mp/internal/logic"
	"mp/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
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

	// new handlers
	var (
		userSelf   = NewUserSelf(svc, userLogic)
		adminUser  = NewAdminUser(svc, userLogic)
		file       = NewFile(svc, fileLogic)
		system     = NewSystem(svc, systemLogic)
		userPublic = NewUserPublic(svc, userLogic)
	)

	return []Handler{
		userSelf,
		adminUser,
		file,
		system,
		userPublic,
	}
}
