package webapp

import (
	"encoding/json"
	"fmt"
	"github.com/ilia-tolliu-go-event-store/internal/eserror"
	"net/http"
)

func ExtractRequestBody[T any](r *http.Request, target T) error {
	err := json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		err = fmt.Errorf("failed to parse request body: %w", err)
		validationErrors := eserror.NewSimpleValidationError("requestBody", "failed to parse request body")
		return eserror.NewValidationError(err, validationErrors)
	}

	return err
}
