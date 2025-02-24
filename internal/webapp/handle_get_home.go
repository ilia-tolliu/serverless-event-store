package webapp

import (
	"context"
	"net/http"
)

func (a *WebApp) HandleGetHome(_ context.Context, _ *http.Request) (Response, error) {
	response := NewResponse(Status(http.StatusPermanentRedirect), Header("Location", "/openapi"))

	return response, nil
}
