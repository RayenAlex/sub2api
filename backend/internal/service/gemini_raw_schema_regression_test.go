package service

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

func TestNormalizeGeminiRawChatToolSchemasAddsItemsToBareNestedArrays(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"fields":{"type":"array"},"pages":{"type":"array","items":{"type":"array"}},"files":{"type":"array","items":{"type":"object"}}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	properties := gjson.GetBytes(got, "tools.0.function.parameters.properties")
	if gotType := properties.Get("fields.type").String(); gotType != "array" {
		t.Fatalf("fields.type = %#v, want array", gotType)
	}
	if got := properties.Get("fields.items.type").String(); got != "string" {
		t.Fatalf("fields.items.type = %#v, want string", got)
	}
	if got := properties.Get("pages.items.items.type").String(); got != "string" {
		t.Fatalf("pages.items.items.type = %#v, want string", got)
	}
	if got := properties.Get("files.items.type").String(); got != "object" {
		t.Fatalf("files.items.type = %#v, want object", got)
	}
}

func TestNormalizeGeminiRawChatToolSchemasRemovesObjectOnlyFieldsFromNonObjectSchemas(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"size":{"type":"string","description":"width","properties":{"width":{"type":"integer"}},"required":["width"]},"ok":{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	properties := gjson.GetBytes(got, "tools.0.function.parameters.properties")
	size := properties.Get("size")
	if size.Get("properties").Exists() || size.Get("required").Exists() {
		t.Fatalf("non-object size schema should not retain object-only fields: %s", size.Raw)
	}
	if size.Get("type").String() != "string" || size.Get("description").String() != "width" {
		t.Fatalf("non-object schema should retain its type and description: %s", size.Raw)
	}
	object := properties.Get("ok")
	if !object.Get("properties").Exists() || !object.Get("required").Exists() {
		t.Fatalf("object schema should retain its object fields: %s", object.Raw)
	}
}

func TestNormalizeGeminiRawChatToolSchemasPromotesUntypedPropertiesToObject(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"size":{"properties":{"width":{"type":"integer"}},"required":["width"]}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	size := gjson.GetBytes(got, "tools.0.function.parameters.properties.size")
	if size.Get("type").String() != "object" {
		t.Fatalf("untyped properties schema should be promoted to object: %#v", size)
	}
	if !size.Get("properties").Exists() {
		t.Fatalf("untyped properties schema should retain properties: %#v", size)
	}
	if !size.Get("required").Exists() {
		t.Fatalf("untyped properties schema should retain required: %#v", size)
	}
}

func TestNormalizeGeminiRawChatToolSchemasHandlesNullableTypes(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"maybe":{"type":["object","null"],"properties":{"id":{"type":"string"}},"required":["id"]},"text":{"type":["string","null"],"properties":{"bad":{"type":"string"}},"required":["bad"]},"nullObject":{"type":null,"properties":{"id":{"type":"string"}}}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	properties := gjson.GetBytes(got, "tools.0.function.parameters.properties")

	maybe := properties.Get("maybe")
	if !maybe.Get("properties").Exists() {
		t.Fatalf("nullable object schema should retain properties: %#v", maybe)
	}
	if !maybe.Get("required").Exists() {
		t.Fatalf("nullable object schema should retain required: %#v", maybe)
	}

	text := properties.Get("text")
	if text.Get("properties").Exists() {
		t.Fatalf("nullable scalar schema should not retain properties: %#v", text)
	}
	if text.Get("required").Exists() {
		t.Fatalf("nullable scalar schema should not retain required: %#v", text)
	}

	nullObject := properties.Get("nullObject")
	if nullObject.Get("type").String() != "object" {
		t.Fatalf("null type with properties should be normalized to object: %#v", nullObject)
	}
	if !nullObject.Get("properties").Exists() {
		t.Fatalf("null type with properties should retain properties: %#v", nullObject)
	}
}

func TestNormalizeGeminiRawChatToolSchemasDoesNotNormalizeSchemaAnnotationValues(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","function":{"name":"search","parameters":{"type":"object","properties":{"value":{"type":"string","default":{"properties":"example","required":"required"},"examples":[{"type":"string","properties":"example"}],"const":{"type":"object"}}}}}}]}`)

	got, err := normalizeGeminiRawChatToolSchemas(body)
	if err != nil {
		t.Fatalf("normalizeGeminiRawChatToolSchemas: %v", err)
	}

	properties := gjson.GetBytes(got, "tools.0.function.parameters.properties")
	value := properties.Get("value")
	assertDeepEqualJSON(t, "default", value.Get("default").Value(), map[string]any{"properties": "example", "required": "required"})
	assertDeepEqualJSON(t, "examples[0]", value.Get("examples.0").Value(), map[string]any{"type": "string", "properties": "example"})
	assertDeepEqualJSON(t, "const", value.Get("const").Value(), map[string]any{"type": "object"})
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
