package logger

import (
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/appmode"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func New(mode appmode.AppMode) *zap.SugaredLogger {
	var config zap.Config

	switch mode {
	case appmode.Development:
		config = zap.NewDevelopmentConfig()
	case appmode.Production:
		config = zap.NewProductionConfig()
	case appmode.Staging:
		config = zap.NewProductionConfig()
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]any{
		"service": "event-store",
	}

	config.OutputPaths = []string{"stdout"}

	log, err := config.Build(zap.WithCaller(true))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to initialize logger: %w", err))
		os.Exit(1)
	}

	return log.Sugar()
}
