package web

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type EsWebApp struct {
	*chi.Mux
	mw []Middleware
}

func NewEsWebApp() *EsWebApp {
	app := &EsWebApp{
		Mux: chi.NewRouter(),
		mw:  []Middleware{},
	}

	app.mw = append(app.mw, MwLogger)
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
			log.Printf("handler error: %v", err)
			return
		}
	}

	a.Mux.MethodFunc(method, path, h)
}
