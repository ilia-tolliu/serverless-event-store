package webapp

import (
	"context"
	"errors"
	"github.com/ilia-tolliu/serverless-event-store/internal/logger"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/resp"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/weberr"
	"net/http"
	"time"
)

func MwLogRequest(handler types.EsHandler) types.EsHandler {
	h := func(ctx context.Context, r *http.Request) (resp.EsResponse, error) {
		log := logger.FromContext(ctx)
		start := time.Now()

		log.Infow("request received",
			"method", r.Method,
			"path", r.URL.Path,
			"remoteAddr", r.RemoteAddr,
		)

		response, err := handler(ctx, r)

		latency := time.Since(start)

		if err != nil {
			status := http.StatusInternalServerError

			webErr := &weberr.WebError{}
			if errors.As(err, &webErr) {
				status = webErr.Status
			}

			log.Errorw("response completed with error",
				"latency", latency,
				"status", status,
			)

			return resp.New(), err
		}

		log.Infow("response completed",
			"latency", latency,
			"status", response.Status(),
		)

		return response, nil
	}

	return h
}
