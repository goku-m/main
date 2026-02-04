package service

import (
	"github.com/goku-m/main/apps/agrifolio/api/repository"
	"github.com/goku-m/main/internal/shared/lib/job"
	"github.com/goku-m/main/internal/shared/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
	User *UserService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	// s.Job.SetAuthService(authService)

	// awsClient, err := aws.NewAWS(s)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create AWS client: %w", err)
	// }

	return &Services{
		Job:  s.Job,
		Auth: authService,
		User: NewUserService(s, repos.User),
	}, nil
}
