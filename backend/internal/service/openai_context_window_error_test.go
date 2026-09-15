package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestIsOpenAIContextWindowErrorRecognizesDevinPromptTooLong(t *testing.T) {
	message := "devin upstream error (invalid_argument): The prompt is too long for this model (trace ID: trace-redacted)"

	require.True(t, isOpenAIContextWindowError(message, nil))
	require.True(t, isOpenAIContextWindowError("", []byte(`{"error":{"message":"The prompt is too long for this model","type":"invalid_request_error"}}`)))
}

func TestSanitizeOpenAIResponseFailedEventForClientNormalizesContextErrorEvent(t *testing.T) {
	payload := []byte(`{"type":"error","status":400,"error":{"message":"devin upstream error (invalid_argument): The prompt is too long for this model (trace ID: trace-redacted)","type":"invalid_request_error"}}`)

	out, changed := sanitizeOpenAIResponseFailedEventForClient(payload, "error", true)

	require.True(t, changed)
	require.Equal(t, "context_length_exceeded", gjson.GetBytes(out, "error.code").String())
	require.Equal(t, "invalid_request_error", gjson.GetBytes(out, "error.type").String())
	require.Contains(t, gjson.GetBytes(out, "error.message").String(), "prompt is too long")
}
