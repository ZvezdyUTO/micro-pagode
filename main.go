package main

import (
	"context"
	"flag"
	"mp/internal/handler/task"
	"os"
	"os/signal"
	"syscall"

	"mp/internal/app"
	"mp/internal/config"
	"mp/internal/handler/api"
	"mp/internal/svc"
	"mp/pkg/conf"

	"golang.org/x/sync/errgroup"
)

type Serve interface {
	Run(ctx context.Context) error
}

const (
	Api = "api"
)

var (
	configFile = flag.String("f", "./etc/local/api.yaml", "the config file")
	modeType   = flag.String("m", "api", "server run mod")
)

func main() {
	flag.Parse()

	// 根的 ctx 必须从 main 里面统一产生，所有 goroutine 共享一个退出信号
	// 任何后台任务不允许私自 contextg.Background() 否则不可控
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
		runner := task.NewRunner(svcCtx)
		srv = api.NewHandle(svcCtx)

		g, ctx := errgroup.WithContext(rootCtx)

		// 后台任务：阻塞运行，ctx 取消后退出
		g.Go(func() error {
			return runner.Start(ctx) // 你现在的 Start 已经是阻塞式了
		})

		// API 服务：需要你把实现改成 Run(ctx)
		g.Go(func() error {
			return srv.Run(ctx)
		})

		if err := g.Wait(); err != nil && err != context.Canceled {
			panic(err)
		}
		return

	default:
		panic("请指定正确的服务")
	}
}
