package middleware

import (
	"mp/pkg/logx"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggingMid struct{}

func NewLoggingMid() *LoggingMid {
	return &LoggingMid{}
}

func (m *LoggingMid) Handler(ctx *gin.Context) {
	// 1. 注入 TraceID
	reqCtx := ctx.Request.Context()
	reqCtx = logx.SetTraceID(reqCtx, "")
	ctx.Request = ctx.Request.WithContext(reqCtx)

	start := time.Now()

	// 2. 放行请求
	ctx.Next()

	// 3. 访问日志
	logx.Infos(
		ctx.Request.Context(),
		"",
		"http_request",
		logx.Fields{
			"log_type": "access",
			"method":   ctx.Request.Method,
			"path":     ctx.FullPath(),
			"status":   ctx.Writer.Status(),
			"latency":  time.Since(start).Milliseconds(),
			"ip":       ctx.ClientIP(),
		},
	)
}
