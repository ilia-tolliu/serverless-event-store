package webapp

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/logger"
	"github.com/ilia-tolliu-go-event-store/internal/repo"
	"go.uber.org/zap"
	"net/http"
)

type WebApp struct {
	*chi.Mux
	mw     []Middleware
	log    *zap.SugaredLogger
	esRepo *repo.EsRepo
}

func NewEsWebApp(esRepo *repo.EsRepo, log *zap.SugaredLogger) *WebApp {
	webApp := &WebApp{
		Mux:    chi.NewRouter(),
		mw:     []Middleware{},
		log:    log,
		esRepo: esRepo,
	}

	webApp.mw = append(webApp.mw, MwLogRequest)
	webApp.Handle(http.MethodGet, "/liveness-check", webApp.HandleLivenessCheck)
	webApp.Handle(http.MethodPost, "/streams/{streamType}", webApp.HandleCreateStream)
	webApp.Handle(http.MethodGet, "/streams/{streamType}/{streamId}/details", webApp.HandleGetStreamDetails)

	return webApp
}

func (a *WebApp) Handle(method string, path string, handler Handler, mw ...Middleware) {
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
