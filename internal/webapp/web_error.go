package webapp

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"net/http"
)

type WebError struct {
	Status           int       `json:"-"`
	MessageForClient string    `json:"message"`
	RequestId        uuid.UUID `json:"requestId"`
	MessageForLog    string    `json:"-"`
	Err              error     `json:"-"`
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

	if errors.As(err, &dataConflictErr) {
		webErr.Status = http.StatusConflict
		webErr.MessageForClient = "Trying to write based on invalid state. Refetch and try again."
	} else if errors.As(err, &notFoundErr) {
		webErr.Status = http.StatusNotFound
		webErr.MessageForClient = "Requested resource not found"
	}

	return webErr
}

func (e *WebError) Error() string {
	return fmt.Errorf("%s: %w", e.MessageForLog, e.Err).Error()
}

func (e *WebError) Unwrap() error {
	return e.Err
}
