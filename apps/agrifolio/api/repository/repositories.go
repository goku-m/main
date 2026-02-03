package repository

import "github.com/goku-m/main/internal/shared/server"

type Repositories struct {
	Site *SiteRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		Site: NewSiteRepository(s),
	}
}
