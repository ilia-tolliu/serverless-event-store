package logger

import (
	"context"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func New(mode config.AppMode) *zap.SugaredLogger {
	var zapConfig zap.Config

	switch mode {
	case config.Development:
		zapConfig = zap.NewDevelopmentConfig()
	case config.Production:
		zapConfig = zap.NewProductionConfig()
	case config.Staging:
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.DisableStacktrace = true
	zapConfig.InitialFields = map[string]any{
		"service": "event-store",
	}

	zapConfig.OutputPaths = []string{"stdout"}

	log, err := zapConfig.Build(zap.WithCaller(true))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to initialize logger: %w", err))
		os.Exit(1)
	}

	return log.Sugar()
}

type loggerCtxKey int

const loggerKey loggerCtxKey = 1

func WithLogger(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	log := ctx.Value(loggerKey).(*zap.SugaredLogger)

	return log
}
