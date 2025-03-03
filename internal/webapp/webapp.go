package webapp

import (
	"encoding/json"
	"errors"
	"github.com/ilia-tolliu/serverless-event-store/internal/logger"
	"github.com/ilia-tolliu/serverless-event-store/internal/repo"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/middleware"
	"github.com/ilia-tolliu/serverless-event-store/internal/webapp/types/weberr"
	"go.uber.org/zap"
	"net/http"
)

type WebApp struct {
	*http.ServeMux
	mw     []middleware.EsMiddleware
	log    *zap.SugaredLogger
	esRepo *repo.EsRepo
}

func New(esRepo *repo.EsRepo, log *zap.SugaredLogger) *WebApp {
	webApp := &WebApp{
		ServeMux: http.NewServeMux(),
		mw:       []middleware.EsMiddleware{},
		log:      log,
		esRepo:   esRepo,
	}

	webApp.mw = append(webApp.mw, MwLogRequest)
	webApp.mw = append(webApp.mw, MwConvertError)
	webApp.esHandle("GET /liveness-check", webApp.HandleLivenessCheck)
	webApp.esHandle("POST /streams/{streamType}", webApp.HandleCreateStream)
	webApp.esHandle("GET /streams/{streamType}", webApp.HandleGetStreams)
	webApp.esHandle("GET /streams/{streamType}/{streamId}/details", webApp.HandleGetStreamDetails)
	webApp.esHandle("PUT /streams/{streamType}/{streamId}/events/{streamRevision}", webApp.HandleAppendEvent)
	webApp.esHandle("GET /streams/{streamType}/{streamId}/events", webApp.HandleGetStreamEvents)

	webApp.HandleFunc("/openapi/openapi-spec.json", HandleOpenapiSpec)
	webApp.HandleFunc("/openapi/", HandleSwaggerUi)
	webApp.Handle("/", http.RedirectHandler("/openapi/", http.StatusMovedPermanently))

	return webApp
}

func (a *WebApp) esHandle(pattern string, handler types.EsHandler, mw ...middleware.EsMiddleware) {
	handler = middleware.Wrap(mw, handler)
	handler = middleware.Wrap(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		requestId := NewRequestId()
		log := a.log.With(zap.String("requestId", requestId.String()))

		ctx := r.Context()
		ctx = logger.WithLogger(ctx, log)
		ctx = WithRequestId(ctx, requestId)

		response, err := handler(ctx, r)
		if err != nil {
			webErr := &weberr.WebError{}
			if errors.As(err, &webErr) {
				response = weberr.IntoResponse(*webErr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		for key, value := range response.Headers() {
			w.Header().Set(key, value)
		}
		w.WriteHeader(response.Status())

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(response.Json())
		if err != nil {
			log.Errorw("failed to encode response", "error", err)
		}
	}

	a.ServeMux.HandleFunc(pattern, h)
}
