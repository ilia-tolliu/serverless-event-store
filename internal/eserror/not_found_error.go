package eserror

import "fmt"

type NotFoundError struct {
	Err error
}

func NewNotFoundError(err error) *NotFoundError {
	return &NotFoundError{Err: err}
}

func (e *NotFoundError) Error() string {
	return fmt.Errorf("not found: %w", e.Err).Error()
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}
