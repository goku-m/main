package router

import (
	"github.com/goku-m/main/apps/todo/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"

	"github.com/labstack/echo/v4"
)

func registerPagesRoutes(r *echo.Echo, h *handler.Handlers, auth *middleware.AuthMiddleware) {

	r.GET("/", h.Todo.GetTodoPage)
	r.GET("/create", h.Todo.CreateTodoPage)
	r.Use(auth.RequireAuthIP)
	r.GET("/update/:id", h.Todo.UpdateTodoPage)
}
