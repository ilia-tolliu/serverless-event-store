package web

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func MwLogger(log *zap.SugaredLogger) Middleware {
	mw := func(handler Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			log.Infow("request received",
				"method", r.Method,
				"path", r.URL.Path,
				"remoteAddr", r.RemoteAddr,
			)

			err := handler(ctx, w, r)

			latency := time.Since(start)

			if err != nil {
				log.Errorw("response error",
					"latency", latency,
					"error", err,
				)

				return err
			}

			log.Infow("response completed",
				"latency", latency,
			)

			return nil
		}

		return h
	}

	return mw
}
