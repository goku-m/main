package router

import (
	"github.com/goku-m/main/apps/task/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

func registerTaskRoutes(r *echo.Group, h *handler.TaskHandler, auth *middleware.AuthMiddleware) {
	// User operations
	tasks := r.Group("/tasks")
	// tasks.Use(auth.RequireAuthIP)

	// Collection operations for pages
	tasks.POST("/create", h.CreateTask)
	tasks.POST("/delete", h.DeleteTask)
	tasks.POST("/update/:id", h.UpdateTask)

}
