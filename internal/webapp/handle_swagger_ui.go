package webapp

import (
	"net/http"
	"os"
	"path/filepath"
)

func HandleSwaggerUi(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()
	swaggerUiDir := http.Dir(filepath.Join(workDir, "swagger_ui"))
	fs := http.StripPrefix(r.Pattern, http.FileServer(swaggerUiDir))
	fs.ServeHTTP(w, r)
}
