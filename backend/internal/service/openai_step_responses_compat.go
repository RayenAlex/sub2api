package service

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
)

const openAIStepResponsesHostname = "api.stepfun.com"

func shouldNormalizeOpenAIStepResponsesAssistantTextContent(account *Account, transport OpenAIUpstreamTransport, compactPath bool) bool {
	if account == nil || !account.IsOpenAIApiKey() || compactPath || transport != OpenAIUpstreamTransportHTTPSSE {
		return false
	}
	if shouldForwardOpenAIResponsesViaRawChatCompletions(account) {
		return false
	}

	baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
	parsed, err := url.Parse(baseURL)
	if err != nil || !strings.EqualFold(parsed.Hostname(), openAIStepResponsesHostname) {
		return false
	}
	return true
}

// normalizeOpenAIStepResponsesAssistantTextContent converts only assistant
// message content arrays made exclusively of output_text blocks into the
// string form accepted by Step's native Responses endpoint. Every other input
// shape is returned unchanged.
func normalizeOpenAIStepResponsesAssistantTextContent(body []byte) ([]byte, int, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(body, &document); err != nil {
		return body, 0, nil
	}

	rawInput, ok := document["input"]
	if !ok {
		return body, 0, nil
	}

	var input []json.RawMessage
	if err := json.Unmarshal(rawInput, &input); err != nil {
		return body, 0, nil
	}

	convertedMessages := 0
	for index, rawItem := range input {
		normalizedItem, itemChanged := normalizeOpenAIStepResponsesAssistantItem(rawItem)
		if !itemChanged {
			continue
		}
		input[index] = normalizedItem
		convertedMessages++
	}
	if convertedMessages == 0 {
		return body, 0, nil
	}

	normalizedInput, err := json.Marshal(input)
	if err != nil {
		return nil, 0, err
	}
	document["input"] = normalizedInput

	normalizedBody, err := json.Marshal(document)
	if err != nil {
		return nil, 0, err
	}
	return normalizedBody, convertedMessages, nil
}

func normalizeOpenAIStepResponsesAssistantItem(rawItem json.RawMessage) (json.RawMessage, bool) {
	var item map[string]json.RawMessage
	if err := json.Unmarshal(rawItem, &item); err != nil || item == nil {
		return rawItem, false
	}

	var role string
	if rawRole, ok := item["role"]; !ok || json.Unmarshal(rawRole, &role) != nil || role != "assistant" {
		return rawItem, false
	}
	if rawType, ok := item["type"]; ok {
		var itemType string
		if json.Unmarshal(rawType, &itemType) != nil || itemType != "message" {
			return rawItem, false
		}
	}

	rawContent, ok := item["content"]
	if !ok {
		return rawItem, false
	}
	var blocks []json.RawMessage
	if err := json.Unmarshal(rawContent, &blocks); err != nil || len(blocks) == 0 {
		return rawItem, false
	}

	var text strings.Builder
	for _, rawBlock := range blocks {
		blockText, ok := openAIStepResponsesOutputTextBlock(rawBlock)
		if !ok {
			return rawItem, false
		}
		text.WriteString(blockText)
	}

	normalizedContent, err := json.Marshal(text.String())
	if err != nil {
		return rawItem, false
	}
	item["content"] = normalizedContent
	// With content in string form, the ResponseItem "message" discriminator
	// selects Step's full output-message schema and rejects the string. Omit
	// only that discriminator so the item stays an EasyInputMessage while all
	// other replay metadata is preserved.
	delete(item, "type")

	normalizedItem, err := json.Marshal(item)
	if err != nil {
		return rawItem, false
	}
	return normalizedItem, true
}

func openAIStepResponsesOutputTextBlock(rawBlock json.RawMessage) (string, bool) {
	var block map[string]json.RawMessage
	if err := json.Unmarshal(rawBlock, &block); err != nil || block == nil {
		return "", false
	}

	if len(block) < 2 || len(block) > 3 {
		return "", false
	}
	for key := range block {
		if key != "type" && key != "text" && key != "annotations" {
			return "", false
		}
	}

	var blockType string
	rawType, ok := block["type"]
	if !ok || json.Unmarshal(rawType, &blockType) != nil || blockType != "output_text" {
		return "", false
	}

	var text string
	rawText, ok := block["text"]
	if !ok || bytes.Equal(bytes.TrimSpace(rawText), []byte("null")) || json.Unmarshal(rawText, &text) != nil {
		return "", false
	}

	if rawAnnotations, ok := block["annotations"]; ok {
		var annotations []json.RawMessage
		if bytes.Equal(bytes.TrimSpace(rawAnnotations), []byte("null")) || json.Unmarshal(rawAnnotations, &annotations) != nil || len(annotations) != 0 {
			return "", false
		}
	}
	return text, true
}
