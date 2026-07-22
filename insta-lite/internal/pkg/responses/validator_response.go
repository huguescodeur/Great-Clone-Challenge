package responses

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func SendValidationError(w http.ResponseWriter, err error) {
	errorsMap := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errorsMap[e.Field()] = "Field is " + e.Tag()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "error",
		"errors": errorsMap,
	})
}
