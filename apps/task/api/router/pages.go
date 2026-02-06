package router

import (
	"github.com/goku-m/main/apps/task/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"

	"github.com/labstack/echo/v4"
)

func registerPagesRoutes(r *echo.Echo, h *handler.Handlers, auth *middleware.AuthMiddleware) {

	r.GET("/", h.Task.GetTaskPage)
	r.GET("/create", h.Task.CreateTaskPage)
	r.Use(auth.RequireAuthIP)
	r.GET("/update/:id", h.Task.UpdateTaskPage)
}
