package agrifolio

import (
	"github.com/goku-m/main/apps/agrifolio/api"
	"github.com/goku-m/main/apps/agrifolio/api/handler"
	"github.com/goku-m/main/apps/agrifolio/api/repository"
	"github.com/goku-m/main/apps/agrifolio/api/service"
	"github.com/goku-m/main/internal/gateway"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

// NewRouter builds the Agrifolio app router for the gateway to mount.
func NewRouter(s *server.Server, h *handler.Handlers) *echo.Echo {
	return api.NewRouter(s, h)
}

func Module(s *server.Server) (gateway.Module, error) {
	repos := repository.NewRepositories(s)
	services, err := service.NewServices(s, repos)
	if err != nil {
		return gateway.Module{}, err
	}

	handlers := api.NewHandlers(s, services)
	router := api.NewRouter(s, handlers)

	return gateway.Module{
		Name:   "agrifolio",
		Prefix: "/agrifolio",
		Router: router,
	}, nil
}
