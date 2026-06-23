package request_test

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"task-service/internal/adapter/http/request"
	"task-service/internal/domain"
)

func TestDecodeJSONAcceptsSingleObject(t *testing.T) {
	var payload struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"Backend"}`))

	if err := request.DecodeJSON(req, &payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload.Name != "Backend" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestDecodeJSONRejectsInvalidInput(t *testing.T) {
	cases := []string{
		``,
		`{`,
		`{"unknown":"value"}`,
		`{"name":"Backend"} {"name":"API"}`,
	}

	for _, body := range cases {
		var payload struct {
			Name string `json:"name"`
		}
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		if err := request.DecodeJSON(req, &payload); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expected invalid input for %q, got %v", body, err)
		}
	}
}

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	var payload struct {
		Name string `json:"name"`
	}
	body := `{"name":"` + strings.Repeat("a", 1<<20) + `"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))

	if err := request.DecodeJSON(req, &payload); !errors.Is(err, domain.ErrPayloadTooLarge) {
		t.Fatalf("expected payload too large, got %v", err)
	}
}
