package logic

import (
	"context"
	"fmt"
	"math"
	"mp/internal/infra/monitorstore"
	"strconv"
	"strings"
	"time"

	"mp/internal/domain"
	"mp/internal/infra/collector"
	"mp/internal/model"
	"mp/pkg/logx"
	"mp/pkg/mail"
)

type System interface {
	// Monitor 一次性获取100s内所有类型历史指标数据
	Monitor(ctx context.Context) (*domain.MonitorInfoResp, error)
	// MonitorState 根据请求参数查询一段记录
	MonitorState(ctx context.Context, req *domain.MonitorStateReq) (*domain.MonitorStateResp, error)
	// GetMonitorConfig 从数据库中读取告警配置
	GetMonitorConfig(ctx context.Context) (*domain.MonitorConfigResp, error)
	// UpdateMonitorConfig 更新或者初始化告警配置
	UpdateMonitorConfig(ctx context.Context, req *domain.MonitorConfigReq) error

	// GetConfig UpdateConfig 管理存储器的最大记录和路径
	GetConfig(ctx context.Context) (maxRecord int, path string)
	UpdateConfig(ctx context.Context, maxRecord int) error

	// CaptureOnce 调用 gosutil 库抓去当前系统状态并且将数据插入存储器
	CaptureOnce(ctx context.Context) error
	// Flush 将内存中咱村的监控指标强制写入磁盘，实现持久化
	Flush(ctx context.Context) error
}

type system struct {
	monitor   *monitorstore.FileStore
	cfgModel  model.SystemMonitorConfigModel
	warnModel model.SystemMonitorWarningModel
	collector collector.Collector
}

func NewSystem(
	monitor *monitorstore.FileStore,
	cfg model.SystemMonitorConfigModel,
	warn model.SystemMonitorWarningModel,
	c collector.Collector,
) System {
	return &system{
		monitor:   monitor,
		cfgModel:  cfg,
		warnModel: warn,
		collector: c,
	}
}

func (l *system) Monitor(ctx context.Context) (*domain.MonitorInfoResp, error) {
	// 定义滑动窗口，猛猛查询所有数据，需要调用底层库实现的具体查询逻辑
	const defaultWindow = 100
	endTime := time.Now().Unix()
	startTime := endTime - defaultWindow

	cpu, err := l.monitor.State(ctx, model.MonitorTypeCpu, startTime, endTime)
	if err != nil {
		return nil, err
	}
	mem, err := l.monitor.State(ctx, model.MonitorTypeMemory, startTime, endTime)
	if err != nil {
		return nil, err
	}
	dist, err := l.monitor.State(ctx, model.MonitorTypeDisk, startTime, endTime)
	if err != nil {
		return nil, err
	}
	netSend, err := l.monitor.State(ctx, model.MonitorTypeNetSend, startTime, endTime)
	if err != nil {
		return nil, err
	}
	netRecv, err := l.monitor.State(ctx, model.MonitorTypeNetRecv, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 然后直接返回吧
	return &domain.MonitorInfoResp{
		CPUState:     model.MonitorKVToDomainStateKV(cpu),
		MemoryState:  model.MonitorKVToDomainStateKV(mem),
		DiskState:    model.MonitorKVToDomainStateKV(dist),
		NetSendState: model.MonitorKVToDomainStateKV(netSend),
		NetRecvState: model.MonitorKVToDomainStateKV(netRecv),
	}, nil
}

func parseMonitorType(t string) (model.MonitorType, error) {
	switch strings.ToLower(t) {
	case "cpu":
		return model.MonitorTypeCpu, nil
	case "memory", "mem":
		return model.MonitorTypeMemory, nil
	case "disk":
		return model.MonitorTypeDisk, nil
	case "netsend", "net_send":
		return model.MonitorTypeNetSend, nil
	case "netrecv", "net_recv":
		return model.MonitorTypeNetRecv, nil
	default:
		return model.MonitorType((strconv.Itoa(0))), fmt.Errorf("monitor type(%s) not support", t)
	}
}

func (l *system) MonitorState(ctx context.Context, req *domain.MonitorStateReq) (*domain.MonitorStateResp, error) {
	mt, err := parseMonitorType(req.Type)
	if err != nil {
		return nil, err
	}

	data, err := l.monitor.State(ctx, mt, 0, math.MaxInt64)
	if err != nil {
		return nil, err
	}

	return &domain.MonitorStateResp{
		Data: model.MonitorKVToDomainStateKV(data),
	}, nil
}

func (l *system) GetMonitorConfig(ctx context.Context) (*domain.MonitorConfigResp, error) {
	cfg, err := l.cfgModel.Get(ctx)

	if err != nil {
		return nil, err
	}

	return &domain.MonitorConfigResp{
		Id:           cfg.Id,
		IsStart:      cfg.IsStart,
		CpuLimit:     cfg.CpuLimit,
		DiskLimit:    cfg.DiskLimit,
		MenLimit:     cfg.MenLimit,
		NetSendLimit: cfg.NetSendLimit,
		NetRecvLimit: cfg.NetRecvLimit,
		NotifyType:   int(cfg.NotifyType),
		Email:        cfg.Email,
	}, nil
}

func (l *system) UpdateMonitorConfig(ctx context.Context, req *domain.MonitorConfigReq) error {
	exist, err := l.cfgModel.Get(ctx)
	if err != nil && err != model.ErrNotFound {
		return err
	}

	notify := model.NotifyType(req.NotifyType)

	if err == model.ErrNotFound {
		return l.cfgModel.Insert(ctx, &model.SystemMonitorConfig{
			IsStart:      req.IsStart,
			CpuLimit:     req.CpuLimit,
			DiskLimit:    req.DiskLimit,
			MenLimit:     req.MenLimit,
			NetSendLimit: req.NetSendLimit,
			NetRecvLimit: req.NetRecvLimit,
			NotifyType:   notify,
			Email:        req.Email,
		})
	}

	// 已存在，更新
	exist.IsStart = req.IsStart
	exist.CpuLimit = req.CpuLimit
	exist.DiskLimit = req.DiskLimit
	exist.MenLimit = req.MenLimit
	exist.NetSendLimit = req.NetSendLimit
	exist.NetRecvLimit = req.NetRecvLimit
	exist.NotifyType = notify
	exist.Email = req.Email

	return l.cfgModel.Update(ctx, exist)
}

func (l *system) GetConfig(ctx context.Context) (maxRecord int, path string) {
	return l.monitor.GetConfig(ctx)
}

func (l *system) UpdateConfig(ctx context.Context, maxRecord int) error {
	return l.monitor.UpdateConfig(ctx, maxRecord)
}

func (l *system) Flush(ctx context.Context) error {
	return l.monitor.Flush(ctx)
}

func (l *system) CaptureOnce(ctx context.Context) error {
	// 先调用采集器获取当前时刻所有数据
	m, err := l.collector.Collect(ctx)
	if err != nil {
		return err
	}

	// 将原始数据转换为项目统一的MonitorKV格式，并且存入内存
	nowKey := m.UnixSec
	if err := l.monitor.InsertOne(ctx, map[model.MonitorType]*model.MonitorKV{
		model.MonitorTypeCpu:     {Key: nowKey, Value: m.CPUPercent},
		model.MonitorTypeMemory:  {Key: nowKey, Value: m.MemPercent},
		model.MonitorTypeDisk:    {Key: nowKey, Value: m.DiskPercent},
		model.MonitorTypeNetSend: {Key: nowKey, Value: m.NetBytesSent},
		model.MonitorTypeNetRecv: {Key: nowKey, Value: m.NetBytesRecv},
	}); err != nil {
		return err
	}

	// 储存完成后，立刻丢去给检查器检查是否有问题
	l.checkAndAlert(ctx, time.Unix(nowKey, 0),
		m.CPUPercent, m.MemPercent, m.DiskPercent, m.NetBytesSent, m.NetBytesRecv)

	return nil
}

func (l *system) checkAndAlert(ctx context.Context, occurrence time.Time, cpuV, memV, diskV, netSendV, netRecvV float64) {
	// 先查看数据库中有没有配置告警规则
	cfg, err := l.cfgModel.Get(ctx)
	if err != nil {
		if err == model.ErrNotFound {
			return
		}
		logx.Error(ctx, "system_monitor:GetConfig", err.Error())
		return
	}
	if cfg.IsStart == false { // 如果开关没打开就塑封么也不做
		return
	}

	// 定义一个闭包函数收集所有异常指标
	day := occurrence.Day()

	var warns []*model.SystemMonitorWarning
	appendWarn := func(t model.MonitorType, limit, val float64) {
		warns = append(warns, &model.SystemMonitorWarning{
			Occurrence: occurrence,
			StateType:  t,
			LimitValue: limit,
			StateValue: val,
			Day:        day,
		})
	}

	// 程序依次比较5项指标，如果这里需要告警并且也超过了阈值，就记录警告
	if cfg.CpuLimit > 0 && cpuV > cfg.CpuLimit {
		appendWarn(model.MonitorTypeCpu, cfg.CpuLimit, cpuV)
	}
	if cfg.MenLimit > 0 && memV > cfg.MenLimit {
		appendWarn(model.MonitorTypeMemory, cfg.MenLimit, memV)
	}
	if cfg.DiskLimit > 0 && diskV > cfg.DiskLimit {
		appendWarn(model.MonitorTypeDisk, cfg.DiskLimit, diskV)
	}
	if cfg.NetSendLimit > 0 && netSendV > cfg.NetSendLimit {
		appendWarn(model.MonitorTypeNetSend, cfg.NetSendLimit, netSendV)
	}
	if cfg.NetRecvLimit > 0 && netRecvV > cfg.NetRecvLimit {
		appendWarn(model.MonitorTypeNetRecv, cfg.NetRecvLimit, netRecvV)
	}

	// 如果没有警告就不管了，否则执行
	if len(warns) == 0 {
		return
	}

	if err := l.warnModel.Inserts(ctx, warns); err != nil {
		logx.Error(ctx, "system_monitor:InsertWarnings", err.Error())
	}

	// 执行！读取smtp配置并且发送邮件
	if cfg.NotifyType == model.NotifyTypeEmail && strings.TrimSpace(cfg.Email) != "" {
		smtpCfg, err := mail.LoadFromEnv()
		if err != nil {
			logx.Error(ctx, "system_monitor:mail_config", err.Error())
			return
		}

		subject := "System Monitor Warning"
		var b strings.Builder
		b.WriteString("System monitor warnings:\n")
		b.WriteString("Occurrence: " + occurrence.Format("2006-01-02 15:04:05") + "\n\n")
		for _, w := range warns {
			b.WriteString(fmt.Sprintf("- type=%s val=%.4f limit=%.4f\n", w.StateType, w.StateValue, w.LimitValue))
		}

		if err := mail.Send(ctx, smtpCfg, []string{strings.TrimSpace(cfg.Email)}, subject, b.String()); err != nil {
			logx.Error(ctx, "system_monitor:send_mail", err.Error())
		}
	}
}
