package main

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal"
	"github.com/ilia-tolliu-go-event-store/internal/config"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	mode := config.NewFromEnv()
	log := logger.New(mode)
	defer eserror.Ignore(log.Sync)

	err := run(mode, log)
	if err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(mode config.AppMode, log *zap.SugaredLogger) error {
	webApp, esConfig, err := internal.BootstrapWebApp(mode, log)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", esConfig.Port),
		Handler: webApp,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	serverErrorsChan := make(chan error, 1)
	go func() {
		log.Infow("startup", "status", "router starting", "host", server.Addr)
		serverErrorsChan <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrorsChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdownChan:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), internal.WebShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			eserror.Ignore(server.Close)
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}
	}

	return nil
}
