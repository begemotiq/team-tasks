package request

import (
	"bytes"
	"encoding/json"
)

// OptionalJSON distinguishes an omitted JSON field from explicit null and a concrete value.
type OptionalJSON[T any] struct {
	Set   bool
	Valid bool
	Value T
}

func (o *OptionalJSON[T]) UnmarshalJSON(data []byte) error {
	o.Set = true

	var zero T
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Valid = false
		o.Value = zero
		return nil
	}

	if err := json.Unmarshal(data, &o.Value); err != nil {
		o.Valid = false
		o.Value = zero
		return err
	}
	o.Valid = true
	return nil
}
