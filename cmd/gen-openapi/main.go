package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func main() {
	_ = os.MkdirAll("docs/swagger", 0o755)
	writeFile("docs/swagger/index.html", []byte(swaggerHTML))

	doc := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Video Server API",
			"version":     "1.0.0",
			"description": "Generated OpenAPI for video-server.",
		},
		"servers": []map[string]any{{"url": "/api/v1"}},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
		"paths": pathsSpec(),
	}
	raw, _ := json.MarshalIndent(doc, "", "  ")
	writeFile("docs/swagger/openapi.json", raw)
}

func writeFile(path string, data []byte) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
}

func pathsSpec() map[string]any {
	return map[string]any{
		"/healthz": map[string]any{
			"get": map[string]any{
				"summary": "Health check",
				"responses": map[string]any{
					"200": map[string]any{"description": "OK"},
				},
			},
		},
		"/auth/register": map[string]any{
			"post": map[string]any{
				"summary":     "Register",
				"description": "Create user account",
				"responses":   map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/auth/login": map[string]any{
			"post": map[string]any{
				"summary":   "Login",
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/upload/check": map[string]any{
			"post": map[string]any{
				"summary":   "Upload hash check",
				"security":  []map[string]any{{"BearerAuth": []string{}}},
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/upload": map[string]any{
			"post": map[string]any{
				"summary":   "Upload video",
				"security":  []map[string]any{{"BearerAuth": []string{}}},
				"responses": map[string]any{"202": map[string]any{"description": "Accepted"}},
			},
		},
		"/videos/{id}/source": map[string]any{
			"get": map[string]any{
				"summary":  "Get video source stream",
				"security": []map[string]any{{"BearerAuth": []string{}}},
				"parameters": []map[string]any{
					{
						"name":     "id",
						"in":       "path",
						"required": true,
						"schema":   map[string]any{"type": "string"},
					},
				},
				"responses": map[string]any{
					"200": map[string]any{"description": "Video stream"},
					"401": map[string]any{"description": "Unauthorized"},
					"404": map[string]any{"description": "Not found"},
					"409": map[string]any{"description": "Video not ready"},
				},
			},
		},
		"/admin/scrape/preview": map[string]any{
			"post": map[string]any{
				"summary":   "Admin scrape preview",
				"security":  []map[string]any{{"BearerAuth": []string{}}},
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/admin/scrape/confirm": map[string]any{
			"put": map[string]any{
				"summary":   "Admin scrape confirm",
				"security":  []map[string]any{{"BearerAuth": []string{}}},
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
	}
}

const swaggerHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Video Server API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/swagger/openapi.json",
      dom_id: "#swagger-ui"
    });
  </script>
</body>
</html>
`
