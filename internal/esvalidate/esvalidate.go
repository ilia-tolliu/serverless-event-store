package esvalidate

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/ilia-tolliu/serverless-event-store/internal/eserror"
	"reflect"
	"strings"
)

var validate = newValidator()

func Validate(v interface{}) error {
	err := validate.Struct(v)

	invalid := &validator.InvalidValidationError{}
	validationErrs := validator.ValidationErrors{}

	if errors.As(err, &validationErrs) {
		messages := make(map[string][]string)
		for _, fieldErr := range validationErrs {
			fieldName := fieldErr.Namespace()
			fieldName = cutPrefix(fieldName)
			messages[fieldName] = append(messages[fieldName], fieldErr.Tag())
		}
		validationErrors := eserror.NewValidationErrors(messages)

		return eserror.NewValidationError(err, validationErrors)
	} else if errors.As(err, &invalid) {
		return eserror.NewValidationError(err, eserror.NewEmptyValidationErrors())
	}

	return err
}

func newValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}

func cutPrefix(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		return strings.Join(parts[1:], ".")
	}

	return path
}
