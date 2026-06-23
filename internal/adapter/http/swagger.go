package http

import (
	"fmt"
	stdhttp "net/http"
	"os"
	"path/filepath"
)

const openAPIPath = "openapi.yaml"

func swaggerUI(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerHTML))
}

func openAPIYAML(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	openAPISpec, err := readOpenAPISpec()
	if err != nil {
		stdhttp.Error(w, "openapi specification is not available", stdhttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.oai.openapi; charset=utf-8")
	w.WriteHeader(stdhttp.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func readOpenAPISpec() ([]byte, error) {
	candidates := []string{
		openAPIPath,
		filepath.Join("..", "..", "..", openAPIPath),
	}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), openAPIPath))
	}
	for _, candidate := range candidates {
		content, err := os.ReadFile(candidate)
		if err == nil {
			return content, nil
		}
	}
	return nil, fmt.Errorf("openapi specification was not found")
}

const swaggerHTML = `<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <title>Task Service API</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #f6f8fa; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/swagger/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true,
      displayRequestDuration: true
    });
  </script>
</body>
</html>`
