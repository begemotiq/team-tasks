package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"task-service/internal/domain"
)

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, domain.ErrPayloadTooLarge):
		status = http.StatusRequestEntityTooLarge
		message = err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = "not found"
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		message = "conflict"
	case errors.Is(err, domain.ErrExternal):
		status = http.StatusServiceUnavailable
		message = err.Error()
	}

	JSON(w, status, NewError(message))
}
