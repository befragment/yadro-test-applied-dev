package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/befragment/yadro-test-applied-dev/internal/config"
	"github.com/befragment/yadro-test-applied-dev/internal/handler/routing"
	l "github.com/befragment/yadro-test-applied-dev/pkg/logger/zap"
	"github.com/befragment/yadro-test-applied-dev/pkg/shutdown"
)

func main() {
	ctx := shutdown.WaitForShutdown()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := l.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	logger.Info(cfg.Port)

	router := routing.Router(logger)

	logger.Info("Starting service...")
	go startServer(ctx, cfg.Port, router, logger)
	<-ctx.Done()
	logger.Info("Service stopped gracefully")
}

func startServer(ctx context.Context, port string, handler http.Handler, logger *l.Logger) {
	srv := &http.Server{
		Addr:    port,
		Handler: handler,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			logger.Errorf("Server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("error shutting down server: %v", err)
	}
}
