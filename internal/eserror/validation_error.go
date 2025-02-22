package eserror

import "fmt"

type ValidationError struct {
	ValidationErrors ValidationErrors
	Err              error
}

func NewValidationError(err error, validationErrors ValidationErrors) *ValidationError {
	return &ValidationError{
		Err:              err,
		ValidationErrors: validationErrors,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Errorf("validation error: %w", e.Err).Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type ValidationErrors struct {
	Messages map[string][]string `json:"messages"`
}

func NewValidationErrors(messages map[string][]string) ValidationErrors {
	return ValidationErrors{Messages: messages}
}

func NewEmptyValidationErrors() ValidationErrors {
	messages := make(map[string][]string)

	return ValidationErrors{Messages: messages}
}
