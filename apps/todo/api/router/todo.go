package router

import (
	"github.com/goku-m/main/apps/todo/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

func registerTodoRoutes(r *echo.Group, h *handler.TodoHandler, auth *middleware.AuthMiddleware) {
	// User operations
	todos := r.Group("/todos")
	// todos.Use(auth.RequireAuthIP)

	// Collection operations for pages
	todos.POST("/create", h.CreateTodo)
	todos.POST("/delete", h.DeleteTodo)
	todos.POST("/update/:id", h.UpdateTodo)

}
