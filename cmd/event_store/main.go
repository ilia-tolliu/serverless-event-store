package main

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal"
	"github.com/ilia-tolliu-go-event-store/internal/appmode"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"github.com/ilia-tolliu-go-event-store/internal/web"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const AppModeKey = "EVENT_STORE_MODE"
const WebShutdownTimeout = 5 * time.Second

func main() {
	mode := appmode.NewFromEnv(AppModeKey)

	log := logger.New(mode)
	defer internal.IgnoreError(log.Sync)

	err := run(log)
	if err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	webApp := web.NewEsWebApp(log)
	server := http.Server{
		Addr:    ":8080",
		Handler: webApp,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Infow("startup", "status", "router starting", "host", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), WebShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			internal.IgnoreError(server.Close)
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}
	}

	return nil
}
