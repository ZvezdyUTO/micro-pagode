package task

import (
	"context"
	"mp/internal/svc"
	"time"

	"github.com/jasonlvhit/gocron"
)

type Runner struct {
	svc *svc.ServiceContext
}

func NewRunner(svc *svc.ServiceContext) *Runner {
	return &Runner{svc: svc}
}

type Service interface {
	Fix(ctx context.Context)      // 启动前修复
	Register(ctx context.Context) // 注册定时任务（不要叫 Run，避免和 Runner 冲突）
	Stop(ctx context.Context)     // 停止时收尾（最终 flush）
}

// Start 阻塞直到 ctx 取消（你现在 main 用 errgroup 管它）
func (r *Runner) Start(ctx context.Context) error {
	services := []Service{
		NewMonitor(r.svc), // NewMonitor 没问题，问题是 Monitor 要实现 Service
	}

	for _, s := range services {
		s.Fix(ctx)
	}
	for _, s := range services {
		s.Register(ctx)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		<-gocron.Start()
	}()

	select {
	case <-ctx.Done():
		// 1) 停止未来调度
		gocron.Clear()
		// 2) 等调度器 goroutine 收敛
		<-done
		// 3) 收尾（最终 flush），给一个短超时避免卡死
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, s := range services {
			s.Stop(stopCtx)
		}
		return ctx.Err()
	case <-done:
		return nil
	}
}
