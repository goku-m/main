package api

import (
	"github.com/goku-m/main/apps/task/api/handler"
	"github.com/goku-m/main/apps/task/api/service"
	"github.com/goku-m/main/internal/shared/server"
)

func NewHandlers(s *server.Server, services *service.Services) *handler.Handlers {
	return handler.NewHandlers(s, services)
}
