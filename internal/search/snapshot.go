package search

import (
	"context"
	"sync"
)

type snapshotManager struct {
	mu    sync.RWMutex
	dirty bool
	files []FileMeta
	build func(ctx context.Context) ([]FileMeta, error)
}

func newSnapshotManager(
	builder func(ctx context.Context) ([]FileMeta, error),
) *snapshotManager {
	return &snapshotManager{
		dirty: true, // 初始必须构建
		build: builder,
	}
}

// getSnapshot 返回一个“有效的快照”
// 如果快照过期，则在这里进行重建
func (s *snapshotManager) getSnapshot(ctx context.Context) ([]FileMeta, error) {
	// 快路径：快照有效
	s.mu.RLock()
	if !s.dirty { // 没有被更新（脏数据）则直接返回
		files := s.files
		s.mu.RUnlock()
		return files, nil
	}
	s.mu.RUnlock()

	// 慢路径：需要重建
	s.mu.Lock()
	defer s.mu.Unlock()

	// double check，避免并发重复构建
	if !s.dirty {
		return s.files, nil
	}

	files, err := s.build(ctx) // 这里的build函数实际上是被创建出实体后才被定义的，在svcCtx那里
	if err != nil {
		return nil, err
	}

	s.files = files
	s.dirty = false
	return files, nil
}

// MarkDirty 标记快照失效
func (s *snapshotManager) MarkDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = true
}
