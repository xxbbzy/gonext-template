package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

const swaggerUI = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>GoNext Template API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: './swagger.json',
          dom_id: '#swagger-ui'
        });
      };
    </script>
  </body>
</html>
`

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		fail("resolve working directory", err)
	}

	sourcePath := filepath.Join(rootDir, "..", "api", "openapi.yaml")
	yamlBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		fail("read OpenAPI source", err)
	}

	var document any
	if err := yaml.Unmarshal(yamlBytes, &document); err != nil {
		fail("parse OpenAPI yaml", err)
	}

	jsonBytes, err := json.MarshalIndent(normalize(document), "", "  ")
	if err != nil {
		fail("render OpenAPI json", err)
	}
	jsonBytes = append(jsonBytes, '\n')

	docsDir := filepath.Join(rootDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		fail("create docs directory", err)
	}

	writeFile(filepath.Join(docsDir, "swagger.yaml"), yamlBytes)
	writeFile(filepath.Join(docsDir, "swagger.json"), jsonBytes)
	writeFile(filepath.Join(docsDir, "index.html"), []byte(swaggerUI))
}

func normalize(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = normalize(item)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[fmt.Sprint(key)] = normalize(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = normalize(item)
		}
		return out
	default:
		return typed
	}
}

func writeFile(path string, content []byte) {
	if err := os.WriteFile(path, content, 0644); err != nil {
		fail("write "+filepath.Base(path), err)
	}
}

func fail(action string, err error) {
	fmt.Fprintf(os.Stderr, "failed to %s: %v\n", action, err)
	os.Exit(1)
}
