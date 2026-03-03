package task

import (
	"context"
	"mp/internal/infra/collector"
	"mp/internal/logic"
	"mp/internal/svc"
	"mp/pkg/logx"
	"sync"

	"github.com/jasonlvhit/gocron"
)

type Monitor struct { // 导入依赖注入和实现逻辑
	svc      *svc.ServiceContext
	sys      logic.System
	flushSec uint64
	mu       sync.Mutex // 关键：防止 CaptureOnce / Flush 并发重叠
}

func NewMonitor(svc *svc.ServiceContext) *Monitor {
	return &Monitor{
		svc: svc,
		sys: logic.NewSystem(
			svc.Monitor,
			svc.SystemMonitorConfig,
			svc.SystemMonitorWarning,
			collector.NewGopsutilCollector(),
		),
		flushSec: uint64(svc.Config.Monitor.File.StudTime), // 从svc注入刷新秒数
	}
}

func (m *Monitor) Fix(ctx context.Context) {
	// 启动前尽量把上次积压的数据刷掉
	ctx = logx.SetTraceID(ctx, "")
	m.mu.Lock()
	defer m.mu.Unlock()
	_ = m.sys.Flush(ctx)
}

// Register 负责定期执行采集任务和落盘任务
func (m *Monitor) Register(ctx context.Context) {
	sec := m.flushSec
	if sec == 0 {
		sec = 10
	}

	// 采集任务
	if err := gocron.Every(sec).Seconds().Do(func() {
		taskCtx := logx.SetTraceID(ctx, "")
		m.mu.Lock()
		defer m.mu.Unlock()

		if err := m.sys.CaptureOnce(taskCtx); err != nil {
			logx.Error(taskCtx, "monitor_task:capture", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// flush 任务
	if err := gocron.Every(sec).Seconds().Do(func() {
		taskCtx := logx.SetTraceID(ctx, "")
		m.mu.Lock()
		defer m.mu.Unlock()

		if err := m.sys.Flush(taskCtx); err != nil {
			logx.Error(taskCtx, "monitor_task:flush", err.Error())
		}
	}); err != nil {
		panic(err)
	}
}

// Stop 退出的时候最后落盘一次
func (m *Monitor) Stop(ctx context.Context) {
	// 退出时最终 flush，尽量别丢数据
	ctx = logx.SetTraceID(ctx, "")
	m.mu.Lock()
	defer m.mu.Unlock()
	_ = m.sys.Flush(ctx)
}
