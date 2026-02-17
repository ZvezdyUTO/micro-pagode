package svc

import (
	"context"
	"errors"
	"mp/internal/config"
	"mp/internal/infra/monitorstore"
	"mp/internal/middleware"
	"mp/internal/model"
	repository "mp/internal/repository"
	"mp/internal/search"
	"mp/pkg/encrypt"
	"mp/pkg/jwt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config

	// 基础设施
	JWT                  *jwt.JWT
	UsersModel           model.UsersModel
	FileRepo             repository.FileRepository
	FileSearch           search.FileSearchService
	Monitor              *monitorstore.FileStore
	SystemMonitorConfig  model.SystemMonitorConfigModel
	SystemMonitorWarning model.SystemMonitorWarningModel

	// Middleware
	JwtMid     *middleware.JWTMid
	AdminMid   *middleware.AdminMid
	LoggingMid *middleware.LoggingMid
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	db, err := gorm.Open(mysql.Open(c.MySql.DataSource), &gorm.Config{}) // 这就是进行组装了
	if err != nil {
		return nil, err
	}

	jwtTool := jwt.NewJWT(
		c.JWT.Secret,
		c.JWT.Expire,
	)

	fileRepo := repository.NewLocalFileRepository()
	fileSearch := search.NewFileSearchService(
		func(ctx context.Context) ([]search.FileMeta, error) {
			return search.BuildSnapshot(ctx, fileRepo, c.FileBasePath)
		},
	)

	store := monitorstore.NewFileStore(c.Monitor.File.Path, c.Monitor.MaxRecord)

	res := &ServiceContext{
		Config:               c,
		UsersModel:           model.NewUsersModel(db),
		JWT:                  jwtTool,
		JwtMid:               middleware.NewJWTMid(jwtTool),
		LoggingMid:           middleware.NewLoggingMid(),
		FileRepo:             repository.NewLocalFileRepository(),
		FileSearch:           fileSearch,
		Monitor:              store,
		SystemMonitorWarning: model.NewSystemMonitorWarningModel(db),
		SystemMonitorConfig:  model.NewSystemMonitorConfigModel(db),
	}

	return res, initServer(res)
}

func initServer(svc *ServiceContext) error {
	ctx := context.Background()

	if err := initSystemUser(ctx, svc); err != nil {
		return err
	}
	if err := initSystemMonitorConfig(ctx, svc); err != nil {
		return err
	}
	return nil
}

func initSystemUser(ctx context.Context, svc *ServiceContext) error {
	systemUser, err := svc.UsersModel.SystemUser()

	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}

	if systemUser != nil {
		return nil
	}

	// 防止重复 root
	u, err := svc.UsersModel.FindByName("root")
	if err == nil && u != nil {
		return nil
	}

	pwd, err := encrypt.GenPasswordHash([]byte("000000"))
	if err != nil {
		return err
	}

	return svc.UsersModel.Insert(ctx, &model.Users{
		Name:     "root",
		Phone:    "123456789",
		Password: string(pwd),
		Status:   model.UserStatusNormal,
		IsSystem: model.IsSystemUser,
	})
}

func initSystemMonitorConfig(ctx context.Context, svc *ServiceContext) error {
	_, err := svc.SystemMonitorConfig.Get(ctx)

	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}

	if errors.Is(err, model.ErrNotFound) {
		defaultCfg := &model.SystemMonitorConfig{
			IsStart:      false,
			CpuLimit:     90,
			DiskLimit:    80,
			MenLimit:     80,
			NetSendLimit: 1024,
			NetRecvLimit: 1024,
			NotifyType:   1,
			Email:        "",
		}

		if err := svc.SystemMonitorConfig.Insert(ctx, defaultCfg); err != nil {
			return err
		}
	}

	return nil
}
