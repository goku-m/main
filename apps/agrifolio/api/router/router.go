package router

import (
	"net/http"

	"github.com/goku-m/main/apps/agrifolio/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/render"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func NewRouter(s *server.Server, h *handler.Handlers) *echo.Echo {
	middlewares := middleware.NewMiddlewares(s)

	router := echo.New()
	router.Renderer = render.NewRenderer("./apps/agrifolio/views", true)
	router.Static("/public", "./web/static")
	router.File("/favicon.ico", "./web/static/favicon.ico")

	router.HTTPErrorHandler = middlewares.Global.GlobalErrorHandler

	// global middlewares
	router.Use(
		echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
			Store: echoMiddleware.NewRateLimiterMemoryStore(rate.Limit(20)),
			DenyHandler: func(c echo.Context, identifier string, err error) error {
				// Record rate limit hit metrics
				if rateLimitMiddleware := middlewares.RateLimit; rateLimitMiddleware != nil {
					rateLimitMiddleware.RecordRateLimitHit(c.Path())
				}

				s.Logger.Warn().
					Str("request_id", middleware.GetRequestID(c)).
					Str("identifier", identifier).
					Str("path", c.Path()).
					Str("method", c.Request().Method).
					Str("ip", c.RealIP()).
					Msg("rate limit exceeded")

				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
			},
		}),
		middlewares.Global.CORS(),
		middlewares.Global.Secure(),
		middleware.RequestID(),
		middlewares.ContextEnhancer.EnhanceContext(),
		middlewares.Global.RequestLogger(),
		middlewares.Global.Recover(),
	)

	// register system routes
	registerSystemRoutes(router, h)
	//register pages route
	registerPagesRoutes(router, h, middlewares.Auth)
	// register api routes
	r := router.Group("/api")
	registerUserRoutes(r, h.User, middlewares.Auth)

	return router
}
