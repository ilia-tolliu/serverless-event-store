package webapp

import (
	"context"
	"github.com/ilia-tolliu/serverless-event-store/internal/logger"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/weberr"
	"net/http"
)

func MwConvertError(handler types.EsHandler) types.EsHandler {
	h := func(ctx context.Context, r *http.Request) (resp.EsResponse, error) {
		log := logger.FromContext(ctx)
		requestId := RequestIdFromContext(ctx)

		response, err := handler(ctx, r)
		if err != nil {
			webErr := weberr.New(requestId, err)
			log.Errorw("failed to handle request", "error", err.Error())
			return resp.EsResponse{}, webErr
		}
		return response, nil
	}

	return h
}
