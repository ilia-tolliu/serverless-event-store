package eshttp

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
)

type Error struct {
	StatusCode    int    `json:"statusCode"`
	CorrelationId string `json:"correlationId"`
	Url           string `json:"url"`
	Message       string `json:"message"`
	Details       string `json:"details"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("failed to request Event Store: %s", err.Message)
}

func ErrorFromHttpResponse(resp *http.Response, message string) *Error {
	var details string
	detailsBytes, err := io.ReadAll(resp.Body)
	if err == nil {
		details = string(detailsBytes)
	}

	return &Error{
		StatusCode:    resp.StatusCode,
		CorrelationId: uuid.Nil.String(), // todo
		Url:           resp.Request.URL.String(),
		Message:       message,
		Details:       details,
	}
}
