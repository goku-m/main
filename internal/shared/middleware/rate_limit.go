package middleware

import (
	"github.com/goku-m/main/internal/shared/server"
)

type RateLimitMiddleware struct {
	server *server.Server
}

func NewRateLimitMiddleware(s *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		server: s,
	}
}

func (r *RateLimitMiddleware) RecordRateLimitHit(endpoint string) {
	if r.server != nil && r.server.Logger != nil {
		r.server.Logger.Warn().
			Str("endpoint", endpoint).
			Msg("rate limit hit")
	}
}
