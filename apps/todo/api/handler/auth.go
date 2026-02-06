package handler

import (
	"net/http"

	"github.com/goku-m/main/internal/shared/render"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Handler
}

func NewAuthHandler(s *server.Server) *AuthHandler {
	return &AuthHandler{
		Handler: NewHandler(s),
	}
}

func (h *Handler) LoginPage(c echo.Context) error {
	td := &render.TemplateData{
		Data: map[string]interface{}{
			"key": "", // frontend key
		},
	}

	err := c.Render(http.StatusOK, "login", td)
	if err != nil {
		c.Logger().Error("LoginPage render error: ", err)
	}
	return err
}
