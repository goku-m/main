package api

import (
	"github.com/goku-m/main/apps/agrifolio/api/handler"
	"github.com/goku-m/main/apps/agrifolio/api/router"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

func NewRouter(s *server.Server, h *handler.Handlers) *echo.Echo {
	return router.NewRouter(s, h)
}
