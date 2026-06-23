package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSwaggerRoutes(t *testing.T) {
	router := NewRouter(Dependencies{})

	uiResponse := httptest.NewRecorder()
	uiRequest := httptest.NewRequest(stdhttp.MethodGet, "/swagger/", nil)
	router.ServeHTTP(uiResponse, uiRequest)

	if uiResponse.Code != stdhttp.StatusOK {
		t.Fatalf("expected swagger UI status 200, got %d", uiResponse.Code)
	}
	if !strings.Contains(uiResponse.Body.String(), "SwaggerUIBundle") {
		t.Fatal("swagger UI response does not contain SwaggerUIBundle")
	}

	specResponse := httptest.NewRecorder()
	specRequest := httptest.NewRequest(stdhttp.MethodGet, "/swagger/openapi.yaml", nil)
	router.ServeHTTP(specResponse, specRequest)

	if specResponse.Code != stdhttp.StatusOK {
		t.Fatalf("expected openapi status 200, got %d", specResponse.Code)
	}
	if !strings.Contains(specResponse.Body.String(), "openapi: 3.0.3") {
		t.Fatal("openapi response does not contain OpenAPI version")
	}

	var spec map[string]any
	if err := yaml.Unmarshal(specResponse.Body.Bytes(), &spec); err != nil {
		t.Fatalf("openapi yaml is invalid: %v", err)
	}
	if _, ok := spec["paths"].(map[string]any); !ok {
		t.Fatal("openapi yaml does not contain paths object")
	}
}
