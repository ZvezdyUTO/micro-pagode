package search

import (
	"context"
	"strings"
)

type fileSearchService struct {
	snapshot *snapshotManager
}

func NewFileSearchService(
	builder func(ctx context.Context) ([]FileMeta, error),
) FileSearchService {
	return &fileSearchService{
		// 持有snapshotManager，代表每个搜索服务实例都有自己的一份快照管理逻辑
		snapshot: newSnapshotManager(builder),
	}
}

func (s *fileSearchService) Search(
	ctx context.Context,
	keyword string,
	opt SearchOption,
) ([]FileMeta, error) {

	// 不区分大小写，线性暴力遍历扫描并且返回
	keyword = strings.ToLower(keyword)

	files, err := s.snapshot.getSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	var result []FileMeta
	for _, f := range files {
		if opt.Prefix {
			if strings.HasPrefix(f.SearchName, keyword) {
				result = append(result, f)
			}
		} else {
			if strings.Contains(f.SearchName, keyword) {
				result = append(result, f)
			}
		}
	}

	return result, nil
}

func (s *fileSearchService) MarkDirty() {
	s.snapshot.MarkDirty()
}
