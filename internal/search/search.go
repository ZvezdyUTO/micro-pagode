package search

import (
	"context"
	"strings"
)

type fileSearchService struct {
	snapshot *snapshotManager
}

func NewFileSearchService(
	builder func(ctx context.Context) ([]FileMeta, error), // 传入快照管理逻辑，在后面可以进行拼装
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

	// 切片的本质是包含三个字段的结构体：指向底层数组的指针、长度和容量，因此这里返回切片其实最后也只是返回了内存的地址
	files, err := s.snapshot.getSnapshot(ctx) // 创建一个 files 切片代表文件快照
	if err != nil {
		return nil, err
	}

	// O(n) 遍历搜索
	var result []FileMeta
	for _, f := range files {
		if opt.Prefix {
			if strings.HasPrefix(f.SearchName, keyword) { // 如果前缀匹配
				result = append(result, f)
			}
		} else {
			if strings.Contains(f.SearchName, keyword) { // 否则尝试模糊搜索
				result = append(result, f)
			}
		}
	}

	return result, nil
}

func (s *fileSearchService) MarkDirty() {
	s.snapshot.MarkDirty()
}
