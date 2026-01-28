package task

import (
	"context"
	"mp/internal/svc"

	"github.com/jasonlvhit/gocron"
)

type Runner struct {
	svc *svc.ServiceContext
}

func NewRunner(svc *svc.ServiceContext) *Runner {
	return &Runner{svc: svc}
}

func (r *Runner) Start(ctx context.Context) {
	// 在这里统一注册所有 task
	services := []interface {
		Fix()
		Run()
	}{
		NewMonitor(r.svc),
	}

	// 启动前修复
	for _, s := range services {
		s.Fix()
	}

	// 启动 task
	for _, s := range services {
		s.Run()
	}

	// 阻塞调度器（必须）
	go func() {
		<-gocron.Start()
	}()

	// 监听退出信号（可选但推荐）
	go func() {
		<-ctx.Done()
		gocron.Clear()
	}()
}
