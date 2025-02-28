package webapp

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"net/http"
)

type WebError struct {
	Status           int       `json:"-"`
	MessageForClient string    `json:"message"`
	RequestId        uuid.UUID `json:"requestId"`
	MessageForLog    string    `json:"-"`
	Err              error     `json:"-"`
	Details          any       `json:"details"`
}

func NewWebError(requestId uuid.UUID, err error) *WebError {
	webErr := &WebError{
		Status:           http.StatusInternalServerError,
		MessageForClient: "Something went wrong",
		RequestId:        requestId,
		MessageForLog:    err.Error(),
		Err:              err,
	}

	dataConflictErr := &eserror.DataConflictError{}
	notFoundErr := &eserror.NotFoundError{}
	invalid := &eserror.ValidationError{}

	if errors.As(err, &dataConflictErr) {
		webErr.Status = http.StatusConflict
		webErr.MessageForClient = "Trying to write based on invalid state. Refetch and try again."
	} else if errors.As(err, &notFoundErr) {
		webErr.Status = http.StatusNotFound
		webErr.MessageForClient = "Requested resource not found"
	} else if errors.As(err, &invalid) {
		webErr.Status = http.StatusBadRequest
		webErr.MessageForLog = "Bad request"
		webErr.Details = invalid.ValidationErrors
	}

	return webErr
}

func (e *WebError) Error() string {
	return fmt.Errorf("%s: %w", e.MessageForLog, e.Err).Error()
}

func (e *WebError) Unwrap() error {
	return e.Err
}

func IntoResponse(err WebError) Response {
	return NewResponse(Status(err.Status), Json(err))
}
