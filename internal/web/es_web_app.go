package web

import (
	"github.com/go-chi/chi/v5"
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

	app.mw = append(app.mw, MwLogger(app.log))
	app.Handle(http.MethodGet, "/liveness-check", app.HandleLivenessCheck)

	return app
}

func (a *EsWebApp) Handle(method string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := handler(ctx, w, r)
		if err != nil {
			// todo convert into HTTP error response
			return
		}
	}

	a.Mux.MethodFunc(method, path, h)
}
