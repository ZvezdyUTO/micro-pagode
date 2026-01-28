package task

import (
	"context"
	"mp/internal/logic"
	"mp/internal/svc"
	"mp/pkg/logx"

	"github.com/jasonlvhit/gocron"
)

type Monitor struct { // 导入依赖注入和实现逻辑
	svc      *svc.ServiceContext
	sys      logic.System
	flushSec uint64
}

func NewMonitor(svc *svc.ServiceContext) *Monitor {
	return &Monitor{
		svc:      svc,
		sys:      logic.NewSystem(svc),
		flushSec: uint64(svc.Config.Monitor.File.StudTime), // 从svc注入刷新秒数
	}
}

func (m *Monitor) Run() {
	// 1. 采集周期（CaptureOnce）
	sec := m.flushSec
	if sec <= 0 {
		sec = 10 // 默认 10 秒，防御
	}

	if err := gocron.Every(sec).Seconds().Do(func() {
		ctx := logx.SetTraceID(context.Background(), "")
		if err := m.sys.CaptureOnce(ctx); err != nil {
			logx.Error(ctx, "monitor_task:capture", err.Error())
		}
	}); err != nil {
		panic(err)
	}

	// 2. Flush 周期（写入文件）
	// 如果你现在想“每次采集都 flush”，可以和采集周期一样
	if err := gocron.Every(sec).Seconds().Do(func() {
		ctx := logx.SetTraceID(context.Background(), "")
		if err := m.sys.Flush(ctx); err != nil {
			logx.Error(ctx, "monitor_task:flush", err.Error())
		}
	}); err != nil {
		panic(err)
	}
}

func (m *Monitor) Fix() { // 如果程序被强制退出，则直接执行 Flush
	ctx := logx.SetTraceID(context.Background(), "")
	_ = m.sys.Flush(ctx)
}
