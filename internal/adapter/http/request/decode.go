package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"task-service/internal/domain"
)

const maxJSONBodyBytes int64 = 1 << 20

func DecodeJSON(r *http.Request, dst any) error {
	defer func() { _ = r.Body.Close() }()
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxJSONBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		if isBodyTooLarge(err) {
			return fmt.Errorf("%w: request body must not exceed %d bytes", domain.ErrPayloadTooLarge, maxJSONBodyBytes)
		}
		return fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if isBodyTooLarge(err) {
			return fmt.Errorf("%w: request body must not exceed %d bytes", domain.ErrPayloadTooLarge, maxJSONBodyBytes)
		}
		return fmt.Errorf("%w: request body must contain a single JSON object", domain.ErrInvalidInput)
	}
	return nil
}

func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
