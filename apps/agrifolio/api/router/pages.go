package router

import (
	"github.com/goku-m/main/apps/agrifolio/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"

	"github.com/labstack/echo/v4"
)

func registerPagesRoutes(r *echo.Echo, h *handler.Handlers, auth *middleware.AuthMiddleware) {

	r.GET("/login", h.Auth.LoginPage)
	r.GET("/", h.User.GetUserPage)
	r.GET("/create", h.User.CreateUserPage)
	r.Use(auth.RequireAuthIP)
	r.GET("/update/:id", h.User.UpdateUserPage)
}
