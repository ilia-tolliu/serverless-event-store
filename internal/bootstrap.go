package internal

import (
	"context"
	"fmt"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ilia-tolliu-go-event-store/internal/config"
	"github.com/ilia-tolliu-go-event-store/internal/repo"
	"github.com/ilia-tolliu-go-event-store/internal/webapp"
	"go.uber.org/zap"
	"net/http"
	"runtime"
)

func BootstrapEsServer(mode config.AppMode, log *zap.SugaredLogger) (*http.Server, error) {
	startupCtx := context.Background() // todo: maybe use context with deadline?

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))
	log.Infow("startup", "mode", mode.String())

	awsConfig, err := awsconfig.LoadDefaultConfig(startupCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config, %w", err)
	}

	esConfig, err := config.FromAws(startupCtx, mode, awsConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from AWS, %w", err)
	}
	log.Infow("startup", "config", esConfig)

	dynamoDb := dynamodb.NewFromConfig(awsConfig)
	esRepo := repo.NewEsRepo(dynamoDb, esConfig.TableName)

	webApp := webapp.NewEsWebApp(esRepo, log)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", esConfig.Port),
		Handler: webApp,
	}

	return server, nil
}
