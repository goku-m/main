package handler

import (
	"github.com/goku-m/main/apps/agrifolio/api/service"
	"github.com/goku-m/main/internal/shared/server"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	User    *UserHandler
	Auth    *AuthHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		User:    NewUserHandler(s, services.User),
		Auth:    NewAuthHandler(s),
	}
}
