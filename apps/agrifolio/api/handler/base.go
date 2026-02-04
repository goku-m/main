package handler

import (
	"time"

	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/goku-m/main/internal/shared/validation"
	"github.com/labstack/echo/v4"
)

// Handler provides base functionality for all handlers
type Handler struct {
	server *server.Server
}

// NewHandler creates a new base handler
func NewHandler(s *server.Server) Handler {
	return Handler{server: s}
}

// HandlerFunc represents a typed handler function that processes a request and returns a response
type HandlerFunc[Req validation.Validatable, Res any] func(c echo.Context, req Req) (Res, error)

// HandlerFuncNoContent represents a typed handler function that processes a request without returning content
type HandlerFuncNoContent[Req validation.Validatable] func(c echo.Context, req Req) error

// ResponseHandler defines the interface for handling different response types
type ResponseHandler interface {
	Handle(c echo.Context, result interface{}) error
	GetOperation() string
}

// JSONResponseHandler handles JSON responses
type JSONResponseHandler struct {
	status int
}

func (h JSONResponseHandler) Handle(c echo.Context, result interface{}) error {
	return c.JSON(h.status, result)
}

func (h JSONResponseHandler) GetOperation() string {
	return "handler"
}

// NoContentResponseHandler handles no-content responses
type NoContentResponseHandler struct {
	status int
}

func (h NoContentResponseHandler) Handle(c echo.Context, result interface{}) error {
	return c.NoContent(h.status)
}

func (h NoContentResponseHandler) GetOperation() string {
	return "handler_no_content"
}

// handleRequest is the unified handler function that eliminates code duplication
func handleRequest[Req validation.Validatable](
	c echo.Context,
	req Req,
	handler func(c echo.Context, req Req) (interface{}, error),
	responseHandler ResponseHandler,
) error {
	start := time.Now()
	method := c.Request().Method
	path := c.Path()
	route := path

	// Get context-enhanced logger
	loggerBuilder := middleware.GetLogger(c).With().
		Str("operation", responseHandler.GetOperation()).
		Str("method", method).
		Str("path", path).
		Str("route", route)

	logger := loggerBuilder.Logger()

	logger.Info().Msg("handling request")

	// Validation with observability
	validationStart := time.Now()
	if err := validation.BindAndValidate(c, req); err != nil {
		validationDuration := time.Since(validationStart)

		logger.Error().
			Err(err).
			Dur("validation_duration", validationDuration).
			Msg("request validation failed")

		return err
	}

	validationDuration := time.Since(validationStart)

	logger.Debug().
		Dur("validation_duration", validationDuration).
		Msg("request validation successful")

	// Execute handler with observability
	handlerStart := time.Now()
	result, err := handler(c, req)
	handlerDuration := time.Since(handlerStart)

	if err != nil {
		totalDuration := time.Since(start)

		logger.Error().
			Err(err).
			Dur("handler_duration", handlerDuration).
			Dur("total_duration", totalDuration).
			Msg("handler execution failed")

		return err
	}

	totalDuration := time.Since(start)

	logger.Info().
		Dur("handler_duration", handlerDuration).
		Dur("validation_duration", validationDuration).
		Dur("total_duration", totalDuration).
		Msg("request completed successfully")

	return responseHandler.Handle(c, result)
}

// Handle wraps a handler with validation, error handling, logging, metrics, and tracing
func Handle[Req validation.Validatable, Res any](
	h Handler,
	handler HandlerFunc[Req, Res],
	status int,
	req Req,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, req, func(c echo.Context, req Req) (interface{}, error) {
			return handler(c, req)
		}, JSONResponseHandler{status: status})
	}
}

// HandleNoContent wraps a handler with validation, error handling, logging, metrics, and tracing for endpoints that don't return content
func HandleNoContent[Req validation.Validatable](
	h Handler,
	handler HandlerFuncNoContent[Req],
	status int,
	req Req,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		return handleRequest(c, req, func(c echo.Context, req Req) (interface{}, error) {
			err := handler(c, req)
			return nil, err
		}, NoContentResponseHandler{status: status})
	}
}
