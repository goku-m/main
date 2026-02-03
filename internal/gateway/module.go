package gateway

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type Module struct {
	Name   string
	Prefix string
	Router http.Handler
}

func New(modules ...Module) *echo.Echo {
	gw := echo.New()

	for _, m := range modules {
		mount(gw, m)
	}

	return gw
}

func mount(gw *echo.Echo, module Module) {
	if module.Router == nil {
		return
	}

	prefix := normalizePrefix(module.Prefix, module.Name)
	handler := stripPrefix(prefix, module.Router)

	// Match both the prefix root and any subpaths.
	gw.Any(prefix, echo.WrapHandler(handler))
	gw.Any(prefix+"/*", echo.WrapHandler(handler))
}

func normalizePrefix(prefix, name string) string {
	if prefix == "" {
		prefix = "/" + name
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if prefix != "/" && strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}
	return prefix
}

func stripPrefix(prefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
		}
		if path == "" {
			path = "/"
		}
		r.URL.Path = path
		next.ServeHTTP(w, r)
	})
}
