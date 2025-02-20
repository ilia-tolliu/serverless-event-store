package web

import (
	"context"
	"net/http"
)

type Handler func(context.Context, *http.Request) (Response, error)
