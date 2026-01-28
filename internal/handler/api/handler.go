package api

import (
	"github.com/gin-gonic/gin"

	"mp/internal/handler"
	"mp/internal/svc"
	"mp/pkg/httpx"
)

type Handler interface {
	InitRegister(*gin.Engine)
}

type handle struct {
	srv  *gin.Engine
	addr string
}

func NewHandle(svc *svc.ServiceContext) *handle {
	h := &handle{
		srv:  gin.Default(),
		addr: "0.0.0.0:8080",
	}
	if len(svc.Config.Addr) > 0 {
		h.addr = svc.Config.Addr
	}

	httpx.SetErrorHandler(handler.ErrorHandler)

	handlers := initHandler(svc)
	for _, handler := range handlers {
		handler.InitRegister(h.srv)
	}

	return h
}

func (h *handle) Run() error {
	return h.srv.Run(h.addr)
}
