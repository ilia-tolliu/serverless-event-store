package eserror

import "fmt"

type DataConflictError struct {
	Err error
}

func NewDataConflictError(err error) *DataConflictError {
	return &DataConflictError{
		Err: err,
	}
}

func (e *DataConflictError) Error() string {
	return fmt.Errorf("data conflict: %w", e.Err).Error()
}

func (e *DataConflictError) Unwrap() error {
	return e.Err
}
