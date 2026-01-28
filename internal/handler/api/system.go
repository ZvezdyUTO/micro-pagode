package api

import (
	"github.com/gin-gonic/gin"

	"mp/internal/domain"
	"mp/internal/logic"
	"mp/internal/svc"
	"mp/pkg/httpx"
)

type System struct {
	svcCtx *svc.ServiceContext
	system logic.System
}

func NewSystem(svcCtx *svc.ServiceContext, system logic.System) *System {
	return &System{
		svcCtx: svcCtx,
		system: system,
	}
}

func (h *System) InitRegister(engine *gin.Engine) {
	g := engine.Group("v1/system", h.svcCtx.JwtMid.Handler)
	g.POST("/monitors", h.Monitor)
	g.POST("/monitors/:type", h.MonitorState)
	g.GET("/monitor/config", h.GetMonitorConfig)
	g.POST("/monitor/config", h.UpdateMonitorConfig)
}

func (h *System) Monitor(ctx *gin.Context) {
	res, err := h.system.Monitor(ctx.Request.Context())
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

func (h *System) MonitorState(ctx *gin.Context) {
	var req domain.MonitorStateReq

	// 绑定 body / query（startTime / endTime 等）
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	// 手动绑定路径参数
	req.Type = ctx.Param("type")

	// 调用 logic
	res, err := h.system.MonitorState(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

func (h *System) GetMonitorConfig(ctx *gin.Context) {
	res, err := h.system.GetMonitorConfig(ctx.Request.Context())
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.OkWithData(ctx, res)
	}
}

func (h *System) UpdateMonitorConfig(ctx *gin.Context) {
	var req domain.MonitorConfigReq
	if err := httpx.BindAndValidate(ctx, &req); err != nil {
		httpx.FailWithErr(ctx, err)
		return
	}

	err := h.system.UpdateMonitorConfig(ctx.Request.Context(), &req)
	if err != nil {
		httpx.FailWithErr(ctx, err)
	} else {
		httpx.Ok(ctx)
	}
}
