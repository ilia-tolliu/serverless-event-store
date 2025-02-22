package webapp

import (
	"context"
	"errors"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"net/http"
	"time"
)

func MwLogRequest(handler Handler) Handler {
	h := func(ctx context.Context, r *http.Request) (Response, error) {
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

			webErr := &WebError{}
			if errors.As(err, &webErr) {
				status = webErr.Status
			}

			log.Errorw("response completed with error",
				"latency", latency,
				"status", status,
			)

			return NewResponse(), err
		}

		log.Infow("response completed",
			"latency", latency,
			"status", response.status,
		)

		return response, nil
	}

	return h
}
