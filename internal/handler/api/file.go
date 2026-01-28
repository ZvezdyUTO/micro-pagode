package api

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"mp/internal/domain"
	"mp/internal/logic"
	"mp/internal/svc"
	"mp/pkg/httpx"
)

type File struct {
	svcCtx *svc.ServiceContext
	file   logic.File
}

func NewFile(svcCtx *svc.ServiceContext, file logic.File) *File {
	return &File{
		svcCtx: svcCtx,
		file:   file,
	}
}

func (h *File) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/file", h.svcCtx.JwtMid.Handler)
	g.GET("", h.List)
	g.POST("/dir", h.CreateDir)
	g.POST("/file", h.CreateFile)
	g.DELETE("", h.Delete)
	g.POST("/upload", h.Upload)
	g.GET("/download", h.Download)
	g.GET("/search", h.Search)
}

// 文件列表
func (h *File) List(ctx *gin.Context) {
	var req domain.FilePathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.file.List(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// 创建目录
func (h *File) CreateDir(ctx *gin.Context) {
	var req domain.FilePathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.file.CreateDir(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// 创建文件
func (h *File) CreateFile(ctx *gin.Context) {
	var req domain.FilePathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.file.CreateFile(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// 删除文件或目录
func (h *File) Delete(ctx *gin.Context) {
	var req domain.FileDeleteReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.file.Delete(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}

// 文件上传
func (h *File) Upload(ctx *gin.Context) {
	var req domain.UploadFileReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.file.Upload(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

// 文件下载
func (h *File) Download(ctx *gin.Context) {
	var req domain.FilePathReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	reader, meta, err := h.file.Download(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()

	// 设置响应头
	ctx.Header("Content-Disposition", "attachment; filename=\""+meta.Filename+"\"")
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", strconv.FormatInt(meta.Size, 10))

	// 流式写入响应
	if _, err := io.Copy(ctx.Writer, reader); err != nil {
		// 这里不能再 httpx.FailWithErr，只能记录日志
		return
	}
}

// 文件搜索
func (h *File) Search(ctx *gin.Context) {
	var req domain.FileSearchReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	res, err := h.file.Search(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}
