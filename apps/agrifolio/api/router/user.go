package router

import (
	"github.com/goku-m/main/apps/agrifolio/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

func registerUserRoutes(r *echo.Group, h *handler.UserHandler, auth *middleware.AuthMiddleware) {
	// User operations
	users := r.Group("/users")
	// users.Use(auth.RequireAuthIP)

	// Collection operations for pages
	users.POST("/create", h.CreateUser)
	users.POST("/delete", h.DeleteUser)
	users.POST("/update/:id", h.UpdateUser)

}
