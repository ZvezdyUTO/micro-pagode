package errno

import "errors"

// 通用
var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrNotExist     = errors.New("target does not exist")
	ErrAlreadyExist = errors.New("target already exists")
	ErrPermission   = errors.New("permission denied")
)

// 文件 / 目录相关
var (
	ErrDirNotEmpty    = errors.New("directory not empty")
	ErrParentNotExist = errors.New("parent directory does not exist")
	ErrHashMismatch   = errors.New("file hash mismatch")
)
