package collector

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
)

type Metrics struct { // 定义数据模型，记录每个硬件占用百分比
	At      time.Time
	UnixSec int64

	CPUPercent  float64
	MemPercent  float64
	DiskPercent float64

	NetBytesSent float64
	NetBytesRecv float64
}

type Collector interface {
	Collect(ctx context.Context) (*Metrics, error)
}

type GopsutilCollector struct{}

func NewGopsutilCollector() *GopsutilCollector { return &GopsutilCollector{} }

func (c *GopsutilCollector) Collect(ctx context.Context) (*Metrics, error) {
	// 单次采集逻辑，把采集数据写入结构体中
	now := time.Now()
	m := &Metrics{
		At:      now,
		UnixSec: now.Unix(),
	}

	if v, err := cpu.Percent(0, false); err == nil && len(v) > 0 {
		m.CPUPercent = v[0]
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		m.MemPercent = vm.UsedPercent
	}

	if du, err := disk.Usage("/"); err == nil {
		m.DiskPercent = du.UsedPercent
	}

	if ios, err := gnet.IOCounters(false); err == nil && len(ios) > 0 {
		m.NetBytesRecv = float64(ios[0].BytesRecv)
		m.NetBytesSent = float64(ios[0].BytesSent)
	}

	return m, nil
}
