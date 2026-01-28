package main

import (
	"context"
	"flag"
	"mp/internal/handler/task"

	"mp/internal/app"
	"mp/internal/config"
	"mp/internal/handler/api"
	"mp/internal/svc"
	"mp/pkg/conf"
)

type Serve interface {
	Run() error
}

const (
	Api = "api"
)

var (
	configFile = flag.String("f", "./etc/local/api.yaml", "the config file")
	modeType   = flag.String("m", "api", "server run mod")
)

func startTasks(svcCtx *svc.ServiceContext) {
	ctx := context.Background()

	runner := task.NewRunner(svcCtx)

	go runner.Start(ctx)
}

func main() {
	flag.Parse()

	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)

	app.InitErrorHandler()

	svcCtx, err := svc.NewServiceContext(cfg)
	if err != nil {
		panic(err)
	}

	var srv Serve

	switch *modeType {
	case Api:
		// 1. 启动 task（后台）
		startTasks(svcCtx)

		// 2. 启动 API（前台，阻塞）
		srv = api.NewHandle(svcCtx)

	default:
		panic("请指定正确的服务")
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
