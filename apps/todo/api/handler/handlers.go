package handler

import (
	"github.com/goku-m/main/apps/todo/api/service"
	"github.com/goku-m/main/internal/shared/server"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Todo    *TodoHandler
	Auth    *AuthHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		Todo:    NewTodoHandler(s, services.Todo),
		Auth:    NewAuthHandler(s),
	}
}
