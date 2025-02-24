package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ilia-tolliu-go-event-store/internal"
	"github.com/ilia-tolliu-go-event-store/internal/config"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"github.com/its-felix/aws-lambda-go-http-adapter/adapter"
	"github.com/its-felix/aws-lambda-go-http-adapter/handler"
	"go.uber.org/zap"
	"os"
)

var mode config.AppMode
var log *zap.SugaredLogger

func init() {
	mode = config.NewFromEnv()
	log = logger.New(mode)

	log.Info("Cold start")
}

func main() {
	defer eserror.Ignore(log.Sync)

	webApp, _, err := internal.BootstrapWebApp(mode, log)
	if err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}

	a := adapter.NewVanillaAdapter(webApp)
	h := handler.NewFunctionURLHandler(a)

	lambda.Start(h)
}
