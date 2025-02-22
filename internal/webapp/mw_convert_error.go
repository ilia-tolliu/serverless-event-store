package webapp

import (
	"context"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"net/http"
)

func MwConvertError(handler Handler) Handler {
	h := func(ctx context.Context, r *http.Request) (Response, error) {
		log := logger.FromContext(ctx)
		requestId := RequestIdFromContext(ctx)

		response, err := handler(ctx, r)
		if err != nil {
			webErr := NewWebError(requestId, err)
			log.Errorw("failed to handle request", "error", err.Error())
			return Response{}, webErr
		}
		return response, nil
	}

	return h
}
