package types

import (
	"context"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"net/http"
)

type EsHandler func(context.Context, *http.Request) (resp.EsResponse, error)
