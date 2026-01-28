package model

import "mp/internal/domain"

type MonitorType string

const (
	MonitorTypeCpu     MonitorType = "cpu"
	MonitorTypeMemory  MonitorType = "memory"
	MonitorTypeDisk    MonitorType = "disk"
	MonitorTypeNetSend MonitorType = "netSend"
	MonitorTypeNetRecv MonitorType = "netRecv"
)

type MonitorKV struct {
	Key   int64   `json:"key"`
	Value float64 `json:"value"`
}

func MonitorKVToDomainStateKV(v []*MonitorKV) []*domain.StateKV {
	if len(v) == 0 {
		return nil
	}
	res := make([]*domain.StateKV, len(v))
	for i, kv := range v {
		res[i] = &domain.StateKV{
			Key: kv.Key,
			Val: kv.Value,
		}
	}
	return res
}
