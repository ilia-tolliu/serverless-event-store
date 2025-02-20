package web

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	handler := NewEsHandler()

	router := chi.NewRouter()
	router.Get("/liveness-check", handler.HandleLivenessCheck)

	return router
}
