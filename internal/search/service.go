package search

import "context"

type SearchOption struct {
	Prefix bool // 是否前缀匹配
}

// 定义接口，包括搜索和标记快照过期
type FileSearchService interface {
	// Search 在文件快照上搜索
	Search(ctx context.Context, keyword string, opt SearchOption) ([]FileMeta, error)

	// MarkDirty 标记快照过期（文件变化后调用）
	MarkDirty()
}
