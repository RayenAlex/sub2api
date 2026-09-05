//go:build geminirawtest

package service

import (
	"encoding/json"
	"testing"
)

func TestNormalizeGeminiRawChatToolSchemasAddsItemsToBareNestedArrays(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"fields":{"type":"array"},"pages":{"type":"array","items":{"type":"array"}},"files":{"type":"array","items":{"type":"object"}}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("decode normalized body: %v", err)
	}
	properties := payload["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["parameters"].(map[string]any)["properties"].(map[string]any)
	if got := properties["fields"].(map[string]any)["items"].(map[string]any)["type"]; got != "string" {
		t.Fatalf("fields.items.type = %#v, want string", got)
	}
	if got := properties["pages"].(map[string]any)["items"].(map[string]any)["items"].(map[string]any)["type"]; got != "string" {
		t.Fatalf("pages.items.items.type = %#v, want string", got)
	}
	if got := properties["files"].(map[string]any)["items"].(map[string]any)["type"]; got != "object" {
		t.Fatalf("files.items.type = %#v, want object", got)
	}
}
