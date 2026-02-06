package api

import (
	"github.com/goku-m/main/apps/todo/api/handler"
	"github.com/goku-m/main/apps/todo/api/service"
	"github.com/goku-m/main/internal/shared/server"
)

func NewHandlers(s *server.Server, services *service.Services) *handler.Handlers {
	return handler.NewHandlers(s, services)
}
