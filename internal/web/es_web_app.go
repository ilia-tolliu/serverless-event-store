package web

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

type EsWebApp struct {
	*chi.Mux
	mw  []Middleware
	log *zap.SugaredLogger
}

func NewEsWebApp(log *zap.SugaredLogger) *EsWebApp {
	app := &EsWebApp{
		Mux: chi.NewRouter(),
		mw:  []Middleware{},
		log: log,
	}

	app.mw = append(app.mw, MwLogRequest)
	app.Handle(http.MethodGet, "/liveness-check", app.HandleLivenessCheck)

	return app
}

func (a *EsWebApp) Handle(method string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := a.log.With(zap.String("requestId", uuid.New().String()))
		ctx = logger.WithLogger(ctx, log)

		response, err := handler(ctx, r)
		if err != nil {
			// todo convert into HTTP error response
			return
		}

		w.WriteHeader(response.status)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(response.json)
		if err != nil {
			log.Errorw("failed to encode response", "error", err)
		}
	}

	a.Mux.MethodFunc(method, path, h)
}
