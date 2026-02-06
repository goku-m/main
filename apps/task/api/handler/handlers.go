package handler

import (
	"github.com/goku-m/main/apps/task/api/service"
	"github.com/goku-m/main/internal/shared/server"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	Task    *TaskHandler
	Auth    *AuthHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		Task:    NewTaskHandler(s, services.Task),
		Auth:    NewAuthHandler(s),
	}
}
