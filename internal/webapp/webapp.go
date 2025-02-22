package webapp

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
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
	webApp.mw = append(webApp.mw, MwConvertError)
	webApp.Handle(http.MethodGet, "/liveness-check", webApp.HandleLivenessCheck)
	webApp.Handle(http.MethodPost, "/streams/{streamType}", webApp.HandleCreateStream)
	webApp.Handle(http.MethodGet, "/streams/{streamType}/{streamId}/details", webApp.HandleGetStreamDetails)
	webApp.Handle(http.MethodPut, "/streams/{streamType}/{streamId}/events/{streamRevision}", webApp.HandleAppendEvent)

	return webApp
}

func (a *WebApp) Handle(method string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		requestId := NewRequestId()
		log := a.log.With(zap.String("requestId", requestId.String()))

		ctx := r.Context()
		ctx = logger.WithLogger(ctx, log)
		ctx = WithRequestId(ctx, requestId)

		response, err := handler(ctx, r)
		if err != nil {
			webErr := &WebError{}
			if errors.As(err, &webErr) {
				response = IntoResponse(*webErr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
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
