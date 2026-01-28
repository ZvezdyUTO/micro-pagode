package search

// FileMeta 是搜索系统内部使用的文件描述
// 注意：这是“快照数据”，不是 domain.File
type FileMeta struct {
	Path       string
	Name       string
	SearchName string // 已经 ToLower
	IsDir      bool
}
