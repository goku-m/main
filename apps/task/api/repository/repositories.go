package repository

import "github.com/goku-m/main/internal/shared/server"

type Repositories struct {
	Task *TaskRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		Task: NewTaskRepository(s),
	}
}
