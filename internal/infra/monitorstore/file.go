package monitorstore

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"mp/internal/infra/flusher"
	"mp/internal/model"
)

type FileStore struct {
	rm sync.RWMutex

	data      map[model.MonitorType][]*model.MonitorKV // 按类型将监控数据地址存放在切片中
	path      string
	maxRecord int

	filename string
}

func NewFileStore(path string, maxRecord int) *FileStore {
	if maxRecord <= 0 {
		maxRecord = 1000
	}
	fs := &FileStore{
		data:      make(map[model.MonitorType][]*model.MonitorKV),
		path:      path,
		maxRecord: maxRecord,
		filename:  filepath.Join(path, "monitor.json"),
	}
	_ = fs.loadFromFile()
	return fs
}

func (s *FileStore) InsertOne(ctx context.Context, data map[model.MonitorType]*model.MonitorKV) error {
	s.rm.Lock()
	defer s.rm.Unlock()
	// 滑动窗口式写入，为了并发安全需要加入写锁

	for k, v := range data {
		if _, ok := s.data[k]; !ok {
			s.data[k] = make([]*model.MonitorKV, 0, 64)
		}
		s.data[k] = append(s.data[k], v)

		// 如果太长了就删除前面的东西
		for len(s.data[k]) > s.maxRecord {
			s.data[k] = s.data[k][1:]
		}
	}
	return nil
}

func (s *FileStore) State(ctx context.Context, t model.MonitorType, startTime, endTime int64) ([]*model.MonitorKV, error) {
	s.rm.RLock()
	defer s.rm.RUnlock()
	// 查找数据，使用读锁，使用二分查找优化

	list := s.data[t]
	if len(list) == 0 {
		return nil, nil
	}

	startIdx := s.binarySearch(list, startTime)
	endIdx := s.binarySearch(list, endTime)

	if startIdx > len(list)-1 {
		return nil, nil
	}
	if endIdx < startIdx {
		return nil, nil
	}

	out := make([]*model.MonitorKV, endIdx-startIdx)
	copy(out, list[startIdx:endIdx])
	return out, nil
}

func (s *FileStore) binarySearch(states []*model.MonitorKV, target int64) int {
	low, high := 0, len(states)
	for low < high {
		mid := low + (high-low)/2
		if states[mid].Key < target {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func (s *FileStore) GetConfig(ctx context.Context) (maxRecord int, path string) {
	s.rm.RLock()
	defer s.rm.RUnlock()
	return s.maxRecord, s.path
}

func (s *FileStore) UpdateConfig(ctx context.Context, maxRecord int) error {
	if maxRecord <= 0 {
		return nil
	}
	s.rm.Lock()
	defer s.rm.Unlock()
	s.maxRecord = maxRecord

	// trim existing
	for k := range s.data {
		for len(s.data[k]) > s.maxRecord {
			s.data[k] = s.data[k][1:]
		}
	}
	return nil
}

func (s *FileStore) Flush(ctx context.Context) error {
	s.rm.RLock() // 使用读锁
	snap := make(map[model.MonitorType][]*model.MonitorKV, len(s.data))
	for k, v := range s.data { // 创建快照并且执行深拷贝，否则释放锁的时候依旧在写入，此时写入数据为混乱数据
		cp := make([]*model.MonitorKV, len(v))
		copy(cp, v)
		snap[k] = cp
	}
	path := s.path
	s.rm.RUnlock() // 快照完成后立刻释放锁

	return flusher.FlushJSONAtomic(path, snap) // 写入数据
}

func (s *FileStore) loadFromFile() error {
	b, err := os.ReadFile(s.filename)
	if err != nil {
		return nil
	}

	var decoded map[model.MonitorType][]*model.MonitorKV
	if err := json.Unmarshal(b, &decoded); err != nil {
		return nil
	}

	// 从磁盘中读取原来的存储文件，并且反序列化回map
	for k, v := range decoded {
		sort.Slice(v, func(i, j int) bool { return v[i].Key < v[j].Key })
		if len(v) > s.maxRecord {
			v = v[len(v)-s.maxRecord:]
		}
		decoded[k] = v
	}

	s.rm.Lock()
	s.data = decoded
	s.rm.Unlock()
	return nil
}

// 静态检查FileStore是否实现了MonitoModel和Configurable的所有方法
var _ model.MonitorModel = (*FileStore)(nil)
var _ Configurable = (*FileStore)(nil)
