package main

import (
	"context"
	"fmt"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ilia-tolliu-go-event-store/internal/config"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"github.com/ilia-tolliu-go-event-store/internal/repo"
	"github.com/ilia-tolliu-go-event-store/internal/webapp"
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
	mode := config.NewFromEnv(AppModeKey)
	log := logger.New(mode)
	defer eserror.Ignore(log.Sync)

	err := run(mode, log)
	if err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(mode config.AppMode, log *zap.SugaredLogger) error {
	startupCtx := context.Background()

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))
	log.Infow("startup", "mode", mode.String())

	awsConfig, err := awsconfig.LoadDefaultConfig(startupCtx)
	if err != nil {
		return fmt.Errorf("failed to load AWS SDK config, %w", err)
	}

	esConfig, err := config.FromAws(startupCtx, mode, awsConfig, log)
	if err != nil {
		return fmt.Errorf("failed to load config from AWS, %w", err)
	}
	log.Infow("startup", "config", esConfig)

	dynamoDb := dynamodb.NewFromConfig(awsConfig)
	esRepo := repo.NewEsRepo(dynamoDb, esConfig.TableName)

	webApp := webapp.NewEsWebApp(esRepo, log)
	server := http.Server{
		Addr:    ":8080",
		Handler: webApp,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

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
			eserror.Ignore(server.Close)
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}
	}

	return nil
}
