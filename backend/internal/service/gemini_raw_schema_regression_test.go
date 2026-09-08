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
	if gotType, _ := properties["fields"].(map[string]any)["type"].(string); gotType != "array" {
		t.Fatalf("fields.type = %#v, want array", gotType)
	}
	if got, _ := properties["fields"].(map[string]any)["items"].(map[string]any)["type"].(string); got != "string" {
		t.Fatalf("fields.items.type = %#v, want string", got)
	}
	if got, _ := properties["pages"].(map[string]any)["items"].(map[string]any)["items"].(map[string]any)["type"].(string); got != "string" {
		t.Fatalf("pages.items.items.type = %#v, want string", got)
	}
	if got, _ := properties["files"].(map[string]any)["items"].(map[string]any)["type"].(string); got != "object" {
		t.Fatalf("files.items.type = %#v, want object", got)
	}
}

func TestNormalizeGeminiRawChatToolSchemasRemovesObjectOnlyFieldsFromNonObjectSchemas(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"size":{"type":"string","description":"width","properties":{"width":{"type":"integer"}},"required":["width"]},"ok":{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("decode normalized body: %v", err)
	}
	properties := payload["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["parameters"].(map[string]any)["properties"].(map[string]any)

	size := properties["size"].(map[string]any)
	if _, exists := size["properties"]; exists {
		t.Fatalf("non-object size schema should not retain properties: %#v", size)
	}
	if _, exists := size["required"]; exists {
		t.Fatalf("non-object size schema should not retain required: %#v", size)
	}
	if size["type"] != "string" || size["description"] != "width" {
		t.Fatalf("non-object schema should retain its type and description: %#v", size)
	}

	ok := properties["ok"].(map[string]any)
	if _, exists := ok["properties"]; !exists {
		t.Fatalf("object schema should retain properties: %#v", ok)
	}
	if _, exists := ok["required"]; !exists {
		t.Fatalf("object schema should retain required: %#v", ok)
	}
}

func TestNormalizeGeminiRawChatToolSchemasPromotesUntypedPropertiesToObject(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"size":{"properties":{"width":{"type":"integer"}},"required":["width"]}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("decode normalized body: %v", err)
	}
	properties := payload["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["parameters"].(map[string]any)["properties"].(map[string]any)
	size := properties["size"].(map[string]any)
	if size["type"] != "object" {
		t.Fatalf("untyped properties schema should be promoted to object: %#v", size)
	}
	if _, exists := size["properties"]; !exists {
		t.Fatalf("untyped properties schema should retain properties: %#v", size)
	}
	if _, exists := size["required"]; !exists {
		t.Fatalf("untyped properties schema should retain required: %#v", size)
	}
}

func TestNormalizeGeminiRawChatToolSchemasHandlesNullableTypes(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"maybe":{"type":["object","null"],"properties":{"id":{"type":"string"}},"required":["id"]},"text":{"type":["string","null"],"properties":{"bad":{"type":"string"}},"required":["bad"]},"nullObject":{"type":null,"properties":{"id":{"type":"string"}}}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("decode normalized body: %v", err)
	}
	properties := payload["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["parameters"].(map[string]any)["properties"].(map[string]any)

	maybe := properties["maybe"].(map[string]any)
	if _, exists := maybe["properties"]; !exists {
		t.Fatalf("nullable object schema should retain properties: %#v", maybe)
	}
	if _, exists := maybe["required"]; !exists {
		t.Fatalf("nullable object schema should retain required: %#v", maybe)
	}

	text := properties["text"].(map[string]any)
	if _, exists := text["properties"]; exists {
		t.Fatalf("nullable scalar schema should not retain properties: %#v", text)
	}
	if _, exists := text["required"]; exists {
		t.Fatalf("nullable scalar schema should not retain required: %#v", text)
	}

	nullObject := properties["nullObject"].(map[string]any)
	if nullObject["type"] != "object" {
		t.Fatalf("null type with properties should be normalized to object: %#v", nullObject)
	}
	if _, exists := nullObject["properties"]; !exists {
		t.Fatalf("null type with properties should retain properties: %#v", nullObject)
	}
}

func TestNormalizeGeminiRawChatToolSchemasDoesNotNormalizeSchemaAnnotationValues(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"value":{"type":"string","default":{"properties":"example","required":"required"},"examples":[{"type":"string","properties":"example"}],"const":{"type":"object"}}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("decode normalized body: %v", err)
	}
	properties := payload["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["parameters"].(map[string]any)["properties"].(map[string]any)
	value := properties["value"].(map[string]any)
	assertDeepEqualJSON(t, "default", value["default"], map[string]any{"properties": "example", "required": "required"})
	assertDeepEqualJSON(t, "examples[0]", value["examples"].([]any)[0], map[string]any{"type": "string", "properties": "example"})
	assertDeepEqualJSON(t, "const", value["const"], map[string]any{"type": "object"})
}

func TestNormalizeGeminiRawChatToolSchemasToleratesMalformedProperties(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":[{"type":"string"}]}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("malformed properties should be left unchanged: %s", got)
	}
}

func assertDeepEqualJSON(t *testing.T, name string, got, want any) {
	t.Helper()
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got %s: %v", name, err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want %s: %v", name, err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("%s = %s, want %s", name, gotJSON, wantJSON)
	}
}
