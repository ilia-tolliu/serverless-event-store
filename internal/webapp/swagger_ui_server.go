package webapp

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SwaggerUiServer(r chi.Router, path string) {
	if strings.ContainsAny(path, "{}*") {
		panic("SwaggerUI does not support URL parameters")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		routeCtx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(routeCtx.RoutePattern(), "/*")

		workDir, _ := os.Getwd()
		swaggerUiDir := http.Dir(filepath.Join(workDir, "swagger_ui"))
		fs := http.StripPrefix(pathPrefix, http.FileServer(swaggerUiDir))

		fs.ServeHTTP(w, r)
	})
}
