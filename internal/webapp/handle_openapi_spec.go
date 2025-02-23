package webapp

import (
	"net/http"
)

func HandleOpenapiSpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi_spec.json")
}
