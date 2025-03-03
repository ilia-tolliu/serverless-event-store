package internal

import (
	"context"
	"fmt"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ilia-tolliu/serverless-event-store/internal/config"
	"github.com/ilia-tolliu/serverless-event-store/internal/repo"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp"
	"go.uber.org/zap"
	"runtime"
	"time"
)

const WebShutdownTimeout = 5 * time.Second

func BootstrapWebApp(mode config.AppMode, log *zap.SugaredLogger) (*webapp.WebApp, *config.EsConfig, error) {
	startupCtx := context.Background() // todo: maybe use context with deadline?

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))
	log.Infow("startup", "mode", mode.String())

	awsConfig, err := awsconfig.LoadDefaultConfig(startupCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load AWS SDK config, %w", err)
	}

	esConfig, err := config.EsConfigFromAws(startupCtx, mode, awsConfig, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config from AWS, %w", err)
	}
	log.Infow("startup", "config", esConfig)

	dynamoDb := dynamodb.NewFromConfig(awsConfig)
	esRepo := repo.NewEsRepo(dynamoDb, esConfig.TableName)

	webApp := webapp.New(esRepo, log)

	return webApp, esConfig, nil
}
