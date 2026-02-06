package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/goku-m/main/apps/task"
	"github.com/goku-m/main/apps/todo"
	"github.com/goku-m/main/internal/gateway"
	"github.com/goku-m/main/internal/shared/config"
	"github.com/goku-m/main/internal/shared/database"
	"github.com/goku-m/main/internal/shared/logger"
	"github.com/goku-m/main/internal/shared/server"
)

const DefaultContextTimeout = 30

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.NewLogger(cfg.Observability)

	if cfg.Primary.Env != "local" {
		if err := database.Migrate(context.Background(), &log, cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to migrate database")
		}
	}

	// Initialize server
	srv, err := server.New(cfg, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize server")
	}

	todoModule, moduleErr := todo.Module(srv)
	if moduleErr != nil {
		log.Fatal().Err(moduleErr).Msg("could not initialize agrifolio module")
	}

	taskModule, moduleErr := task.Module(srv)
	if moduleErr != nil {
		log.Fatal().Err(moduleErr).Msg("could not initialize agrifolio module")
	}

	// Initialize gateway router
	r := gateway.New(todoModule, taskModule)

	// Setup HTTP server
	srv.SetupHTTPServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	// Start server
	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}
	stop()
	cancel()

	log.Info().Msg("server exited properly")
}
