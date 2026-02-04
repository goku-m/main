package repository

import "github.com/goku-m/main/internal/shared/server"

type Repositories struct {
	User *UserRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		User: NewUserRepository(s),
	}
}
