package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestWebSearchEventsJSON(t *testing.T) {
	body := []byte(`{"output":[
		{"type":"message","content":[{"text":"not logged"}]},
		{"type":"web_search_call","id":"ws_1","status":"completed","action":{"type":"search","query":"first query","sources":[{"url":"https://example.com/one","title":"One","snippet":"not logged"}]}},
		{"type":"function_call","name":"web_search","arguments":"not a hosted call"},
		{"type":"web_search_call","id":"ws_2","call_id":"call_2","status":"failed","action":{"queries":["second query","third query"]}}
	]}`)
	events := parseWebSearchEventsFromJSONBytes(body)
	require.Len(t, events, 2)
	require.Equal(t, 1, events[0].Sequence)
	require.Equal(t, "ws_1", events[0].CallID)
	require.Equal(t, "first query", events[0].Query)
	require.Equal(t, "completed", events[0].Status)
	require.Equal(t, []WebSearchSource{{URL: "https://example.com/one", Title: "One"}}, events[0].Sources)
	require.Equal(t, 1, events[0].SourceCount)
	require.WithinDuration(t, time.Now(), events[0].CreatedAt, time.Second)
	require.Equal(t, 2, events[1].Sequence)
	require.Equal(t, "call_2", events[1].CallID)
	require.Equal(t, "second query\nthird query", events[1].Query)
	require.Equal(t, "failed", events[1].Status)
	require.Empty(t, events[1].Sources)
	serialized, err := json.Marshal(events)
	require.NoError(t, err)
	require.NotContains(t, string(serialized), "not logged")
	require.NotContains(t, string(serialized), "snippet")
}

func TestWebSearchEventsTolerateMissingAndMalformedMetadata(t *testing.T) {
	for _, body := range []string{"", "{", "null", `{"output":{}}`, `{"output":[{"type":"x_search_call"}]}`} {
		require.Empty(t, parseWebSearchEventsFromJSONBytes([]byte(body)))
	}
	body := []byte(`{"output":[
		{"type":"web_search_call"},
		{"type":"web_search_call","call_id":{"secret":"ignore"},"status":true,"action":{"query":{"password":"ignore"},"sources":[null,"invalid",{"url":123},{"url":"javascript:alert(1)"},{"url":"https://user:password@example.com"},{"url":"https://example.com/ok","title":{"html":"ignore"}},{"url":"https://example.com/ok"},{"url":"/relative"},{"url":"https://example.com/last","title":"Last"}]}}
	]}`)
	events := parseWebSearchEventsFromJSONBytes(body)
	require.Len(t, events, 2)
	require.Empty(t, events[0].Query)
	require.Empty(t, events[0].Sources)
	require.Empty(t, events[1].CallID)
	require.Empty(t, events[1].Query)
	require.Empty(t, events[1].Status)
	require.Equal(t, []WebSearchSource{{URL: "https://example.com/ok"}, {URL: "https://example.com/last", Title: "Last"}}, events[1].Sources)
	require.Equal(t, 2, events[1].SourceCount)
}

func TestWebSearchEventsSSEDeduplicateEnrichAndOrder(t *testing.T) {
	body := `event: response.output_item.done
` + `data: {"output_index":3,"item":{"type":"web_search_call","id":"ws_2","status":"completed","action":{"query":"second"}}}` + "\n\n" +
		`data: {"type":"response.output_item.done","output_index":1,"item":{"type":"web_search_call","id":"ws_1","status":"completed","action":{"query":"first"}}}` + "\n\n" +
		`data: {"type":"response.output_item.done","output_index":1,"item":{"type":"web_search_call","id":"ws_1","action":{"query":"first"}}}` + "\n\n" +
		`data: {"type":"response.completed","response":{"output":[{"type":"message"},{"type":"web_search_call","id":"ws_1","action":{"sources":[null,{"url":"https://example.com"}]}},{"type":"message"},{"type":"web_search_call","id":"ws_2"}]}}` + "\n\n"
	events := parseWebSearchEventsFromSSEBody(body)
	require.Len(t, events, 2)
	require.Equal(t, "ws_1", events[0].CallID)
	require.Equal(t, "first", events[0].Query)
	require.Equal(t, "completed", events[0].Status)
	require.Equal(t, 1, events[0].SourceCount)
	require.Equal(t, "second", events[1].Query)
	require.Equal(t, 1, events[0].Sequence)
	require.Equal(t, 2, events[1].Sequence)
}

func TestWebSearchEventsSSEAnonymousAndTerminalOnly(t *testing.T) {
	for _, done := range []string{"", `data: {"type":"response.output_item.done","item":{"type":"web_search_call","action":{"query":"same"}}}` + "\n\n" +
		`data: {"type":"response.output_item.done","item":{"type":"web_search_call","action":{"query":"same"}}}` + "\n\n"} {
		terminal := `data: {"type":"response.completed","response":{"output":[{"type":"web_search_call","action":{"query":"same"}},{"type":"web_search_call","action":{"query":"same"}}]}}` + "\n\n"
		events := parseWebSearchEventsFromSSEBody(done + terminal + terminal)
		require.Len(t, events, 2, "identical queries are distinct calls; repeated terminal snapshots aren't")
		require.Equal(t, 2, events[1].Sequence)
	}
	// Identity can be enriched with call_id in the final snapshot.
	events := parseWebSearchEventsFromSSEBody(`data: {"type":"response.output_item.done","item":{"type":"web_search_call","id":"ws_1"}}` + "\n\n" +
		`data: {"type":"response.done","response":{"output":[{"type":"web_search_call","id":"ws_1","call_id":"call_1"}]}}` + "\n\n")
	require.Len(t, events, 1)
	require.Equal(t, "call_1", events[0].CallID)
}

func TestWebSearchEventsIgnoreProgressAndPreferNestedOutput(t *testing.T) {
	c := newWebSearchEventCollector()
	for _, eventType := range []string{"response.output_item.added", "response.web_search_call.completed", "response.output_text.delta", "response.in_progress"} {
		c.observeSSE(eventType, []byte(`{"item":{"type":"web_search_call","id":"progress"}}`))
	}
	c.observeSSE("response.completed", []byte(`{"response":{"output":[{"type":"web_search_call","id":"actual"}]},"output":[{"type":"web_search_call","id":"shadow"}]}`))
	require.Len(t, c.eventsCopy(), 1)
	require.Equal(t, "actual", c.eventsCopy()[0].CallID)
}

func TestWebSearchEventsBoundsAndSnapshot(t *testing.T) {
	sources := []WebSearchSource{{URL: "https://example.com/" + strings.Repeat("x", maxWebSearchEventURLLength)}}
	for i := 0; i < maxWebSearchEventSources+5; i++ {
		sources = append(sources, WebSearchSource{URL: fmt.Sprintf("https://example.com/%d", i), Title: strings.Repeat("界", maxWebSearchEventTitleLength+10)})
	}
	event := WebSearchEvent{Sequence: 900, CallID: strings.Repeat("x", 200), Query: strings.Repeat("界", 2000), Status: "completed\x00\x01", SourceCount: 1000, Sources: sources}
	events := make([]WebSearchEvent, maxWebSearchEvents+10)
	for i := range events {
		events[i] = event
	}
	normalized := NormalizeWebSearchEvents(events)
	require.Len(t, normalized, maxWebSearchEvents)
	require.Equal(t, maxWebSearchEvents, normalized[len(normalized)-1].Sequence)
	require.Len(t, normalized[0].CallID, maxWebSearchEventCallIDLength)
	require.Equal(t, maxWebSearchEventQueryLength, utf8.RuneCountInString(normalized[0].Query))
	require.Equal(t, "completed", normalized[0].Status)
	require.Len(t, normalized[0].Sources, maxWebSearchEventSources)
	require.Equal(t, maxWebSearchEventSources, normalized[0].SourceCount)
	require.Equal(t, maxWebSearchEventTitleLength, utf8.RuneCountInString(normalized[0].Sources[0].Title))
	normalized[0].Sources[0].Title = "changed"
	require.NotEqual(t, "changed", events[0].Sources[1].Title)

	c := newWebSearchEventCollector()
	for i := 0; i < 100; i++ {
		c.observeSSE("response.output_item.done", []byte(fmt.Sprintf(`{"output_index":%d,"item":{"type":"web_search_call","id":"ws_%d","action":{"query":"query"}}}`, i, i)))
	}
	require.Len(t, c.items, maxWebSearchEvents)
	require.Len(t, c.eventsCopy(), maxWebSearchEvents)
}

func TestWebSearchEventsOnlyOpenAI(t *testing.T) {
	body := []byte(`{"output":[{"type":"web_search_call","id":"ws_1"}]}`)
	require.Len(t, parseOpenAIWebSearchEvents(&Account{Platform: PlatformOpenAI}, body, false), 1)
	require.Empty(t, parseOpenAIWebSearchEvents(&Account{Platform: PlatformGrok}, body, false))
	require.Empty(t, parseOpenAIWebSearchEvents(nil, body, false))
}

func TestWebSearchEventsStableIdentityWinsOverShiftedIndexes(t *testing.T) {
	c := newWebSearchEventCollector()
	c.observeSSE("response.output_item.done", []byte(`{"output_index":0,"item":{"type":"web_search_call","id":"ws_one","action":{"query":"one"}}}`))
	c.observeSSE("response.output_item.done", []byte(`{"output_index":1,"item":{"type":"web_search_call","id":"ws_two","action":{"query":"two"}}}`))
	c.observeSSE("response.completed", []byte(`{"response":{"output":[{"type":"message"},{"type":"web_search_call","id":"ws_one","action":{"sources":[{"url":"https://example.com/one"}]}},{"type":"web_search_call","id":"ws_two"}]}}`))
	events := c.eventsCopy()
	require.Len(t, events, 2)
	require.Equal(t, "one", events[0].Query)
	require.Equal(t, "two", events[1].Query)
	require.Equal(t, "https://example.com/one", events[0].Sources[0].URL)
	events[0].Sources[0].URL = "https://modified.example"
	require.Equal(t, "https://example.com/one", c.eventsCopy()[0].Sources[0].URL)
}
