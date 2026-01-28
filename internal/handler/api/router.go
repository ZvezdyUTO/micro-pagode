package api

import (
	"mp/internal/logic"
	"mp/internal/svc"
)

func initHandler(svc *svc.ServiceContext) []Handler {
	// new logics
	var (
		userLogic   = logic.NewUser(svc)
		fileLogic   = logic.NewFile(svc)
		systemLogic = logic.NewSystem(svc)
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
