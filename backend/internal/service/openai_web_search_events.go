package service

import (
	"crypto/sha256"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/tidwall/gjson"
)

const (
	maxWebSearchEvents            = 32
	maxWebSearchEventQueryLength  = 1024
	maxWebSearchEventCallIDLength = 128
	maxWebSearchEventStatusLength = 64
	maxWebSearchEventSources      = 20
	maxWebSearchEventURLLength    = 2048
	maxWebSearchEventTitleLength  = 512
)

// WebSearchSource intentionally retains no snippets, page content or raw payloads.
type WebSearchSource struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// WebSearchEvent is display-only metadata under a usage log, never a billable row.
type WebSearchEvent struct {
	ID          int64
	UsageLogID  int64
	Sequence    int
	CallID      string
	Query       string
	Status      string
	SourceCount int
	Sources     []WebSearchSource
	CreatedAt   time.Time
}

// NormalizeWebSearchEvents bounds and detaches metadata before asynchronous
// persistence. It does not retain the upstream response backing strings.
func NormalizeWebSearchEvents(events []WebSearchEvent) []WebSearchEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]WebSearchEvent, min(len(events), maxWebSearchEvents))
	for i := range out {
		out[i] = events[i]
		out[i].Sequence = i + 1
		out[i].CallID = boundedWebSearchString(events[i].CallID, maxWebSearchEventCallIDLength)
		out[i].Query = boundedWebSearchString(events[i].Query, maxWebSearchEventQueryLength)
		out[i].Status = boundedWebSearchString(events[i].Status, maxWebSearchEventStatusLength)
		out[i].Sources = normalizeWebSearchSources(events[i].Sources)
		out[i].SourceCount = len(out[i].Sources)
	}
	return out
}

type collectedWebSearchEvent struct {
	event       WebSearchEvent
	itemID      [32]byte
	callID      [32]byte
	outputIndex int
}

// This observer consumes the same decoded Responses payloads / SSE frames used
// by usage accounting. Keep a bounded projection, not another response buffer.
type webSearchEventCollector struct {
	items []collectedWebSearchEvent
}

// Limit first-phase collection to OpenAI/Codex; Grok's standalone search and
// its billing counters keep their existing behavior.
func newOpenAIWebSearchEventCollector(account *Account) *webSearchEventCollector {
	if account == nil || !account.IsOpenAI() {
		return nil
	}
	return newWebSearchEventCollector()
}

func parseOpenAIWebSearchEvents(account *Account, body []byte, sse bool) []WebSearchEvent {
	collector := newOpenAIWebSearchEventCollector(account)
	if collector == nil {
		return nil
	}
	if sse {
		forEachOpenAISSEFrame(string(body), collector.observeSSE)
	} else {
		collector.observeJSON(body)
	}
	return collector.eventsCopy()
}
func newWebSearchEventCollector() *webSearchEventCollector {
	return &webSearchEventCollector{}
}

func (c *webSearchEventCollector) observeJSON(body []byte) {
	if c == nil || !gjson.ValidBytes(body) {
		return
	}
	c.observeOutput(webSearchOutputFromJSON(body))
}

func (c *webSearchEventCollector) observeSSE(eventType string, data []byte) {
	if c == nil {
		return
	}
	eventType = effectiveOpenAISSEEventType(data, eventType)
	switch eventType {
	case "response.output_item.done", "response.completed", "response.done":
		if !gjson.ValidBytes(data) {
			return
		}
	default:
		return
	}
	if eventType == "response.output_item.done" {
		index := -1
		if value := gjson.GetBytes(data, "output_index"); value.Type == gjson.Number && value.Int() >= 0 {
			index = int(value.Int())
		}
		c.observeItem(gjson.GetBytes(data, "item"), index, nil)
		return
	}
	c.observeOutput(webSearchOutputFromJSON(data))
}

func (c *webSearchEventCollector) observeOutput(output gjson.Result) {
	if !output.IsArray() {
		return
	}
	// Terminal output is authoritative for order and can enrich item.done with
	// sources. Match each ID-less completion at most once within this snapshot.
	matched := make(map[int]bool)
	output.ForEach(func(index, item gjson.Result) bool {
		c.observeItem(item, int(index.Int()), matched)
		return true
	})
}

func webSearchIdentity(value gjson.Result) [32]byte {
	if value.Type != gjson.String || strings.TrimSpace(value.Str) == "" {
		return [32]byte{}
	}
	return sha256.Sum256([]byte(strings.TrimSpace(value.Str)))
}

func (c *webSearchEventCollector) observeItem(item gjson.Result, outputIndex int, terminalMatched map[int]bool) {
	if !item.IsObject() || webSearchString(item.Get("type")) != "web_search_call" {
		return
	}
	itemID, callID := webSearchIdentity(item.Get("id")), webSearchIdentity(item.Get("call_id"))
	var emptyID [32]byte
	match := -1
	for i, previous := range c.items {
		if (itemID != emptyID && previous.itemID == itemID) ||
			(callID != emptyID && previous.callID == callID) {
			match = i
			break
		}
	}
	// Prefer stable identities over positions when a compatible terminal snapshot
	// shifts output indexes. Never merge two explicitly different IDs by index.
	if match < 0 && outputIndex >= 0 {
		for i, previous := range c.items {
			itemConflict := itemID != emptyID && previous.itemID != emptyID && itemID != previous.itemID
			callConflict := callID != emptyID && previous.callID != emptyID && callID != previous.callID
			if previous.outputIndex == outputIndex && !itemConflict && !callConflict {
				match = i
				break
			}
		}
	}
	if match < 0 && terminalMatched != nil && itemID == emptyID && callID == emptyID {
		for i, previous := range c.items {
			if previous.itemID == emptyID && previous.callID == emptyID && previous.outputIndex < 0 && !terminalMatched[i] {
				match = i
				break
			}
		}
	}
	if match < 0 && len(c.items) >= maxWebSearchEvents {
		return
	}

	event := WebSearchEvent{
		CallID:    boundedWebSearchString(firstNonEmpty(webSearchString(item.Get("call_id")), webSearchString(item.Get("id"))), maxWebSearchEventCallIDLength),
		Query:     webSearchQuery(item),
		Status:    boundedWebSearchString(webSearchString(item.Get("status")), maxWebSearchEventStatusLength),
		Sources:   extractWebSearchSources(item),
		CreatedAt: time.Now(),
	}
	event.SourceCount = len(event.Sources)
	if match >= 0 {
		previous := &c.items[match]
		// Missing/malformed fields in a later terminal snapshot must not erase
		// usable metadata from item.done. Sources can arrive only at completion.
		if event.CallID != "" {
			previous.event.CallID = event.CallID
		}
		if event.Query != "" {
			previous.event.Query = event.Query
		}
		if event.Status != "" {
			previous.event.Status = event.Status
		}
		if len(event.Sources) > 0 {
			previous.event.Sources = event.Sources
			previous.event.SourceCount = len(event.Sources)
		}
		if outputIndex >= 0 {
			previous.outputIndex = outputIndex
		}
		if itemID != emptyID {
			previous.itemID = itemID
		}
		if callID != emptyID {
			previous.callID = callID
		}
	} else {
		match = len(c.items)
		c.items = append(c.items, collectedWebSearchEvent{event: event, itemID: itemID, callID: callID, outputIndex: outputIndex})
	}
	if terminalMatched != nil {
		terminalMatched[match] = true
	}
}

func (c *webSearchEventCollector) eventsCopy() []WebSearchEvent {
	if c == nil || len(c.items) == 0 {
		return nil
	}
	items := append([]collectedWebSearchEvent(nil), c.items...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].outputIndex < 0 || items[j].outputIndex < 0 {
			return items[i].outputIndex >= 0 && items[j].outputIndex < 0
		}
		return items[i].outputIndex < items[j].outputIndex
	})
	out := make([]WebSearchEvent, len(items))
	for i := range items {
		out[i] = items[i].event
	}
	return NormalizeWebSearchEvents(out)
}

func parseWebSearchEventsFromJSONBytes(body []byte) []WebSearchEvent {
	collector := newWebSearchEventCollector()
	collector.observeJSON(body)
	return collector.eventsCopy()
}

func parseWebSearchEventsFromSSEBody(body string) []WebSearchEvent {
	collector := newWebSearchEventCollector()
	forEachOpenAISSEFrame(body, collector.observeSSE)
	return collector.eventsCopy()
}

func webSearchOutputFromJSON(body []byte) gjson.Result {
	if nested := gjson.GetBytes(body, "response.output"); nested.IsArray() {
		return nested
	}
	return gjson.GetBytes(body, "output")
}

func webSearchString(value gjson.Result) string {
	if value.Type != gjson.String {
		return ""
	}
	return value.Str
}

func webSearchQuery(item gjson.Result) string {
	query := firstNonEmpty(webSearchString(item.Get("action.query")), webSearchString(item.Get("query")))
	if query == "" {
		// Newer Responses variants can report several queries for one invocation.
		queries := item.Get("action.queries")
		if queries.IsArray() {
			queries.ForEach(func(_, value gjson.Result) bool {
				part := boundedWebSearchString(webSearchString(value), maxWebSearchEventQueryLength)
				if part != "" {
					if query != "" {
						query += "\n"
					}
					query += part
				}
				return len(query) < maxWebSearchEventQueryLength
			})
		}
	}
	return boundedWebSearchString(query, maxWebSearchEventQueryLength)
}

func extractWebSearchSources(item gjson.Result) []WebSearchSource {
	sources := item.Get("action.sources")
	if !sources.IsArray() {
		sources = item.Get("sources")
	}
	if !sources.IsArray() {
		return nil
	}
	out := make([]WebSearchSource, 0, maxWebSearchEventSources)
	seen := make(map[string]bool)
	sources.ForEach(func(_, source gjson.Result) bool {
		if source.IsObject() {
			if normalized, ok := normalizeWebSearchSource(WebSearchSource{URL: webSearchString(source.Get("url")), Title: webSearchString(source.Get("title"))}); ok && !seen[normalized.URL] {
				out = append(out, normalized)
				seen[normalized.URL] = true
			}
		}
		return len(out) < maxWebSearchEventSources
	})
	return out
}

func normalizeWebSearchSources(sources []WebSearchSource) []WebSearchSource {
	out := make([]WebSearchSource, 0, min(len(sources), maxWebSearchEventSources))
	seen := make(map[string]bool)
	for _, source := range sources {
		if normalized, ok := normalizeWebSearchSource(source); ok && !seen[normalized.URL] {
			out = append(out, normalized)
			seen[normalized.URL] = true
		}
		if len(out) >= maxWebSearchEventSources {
			break
		}
	}
	return out
}

func normalizeWebSearchSource(source WebSearchSource) (WebSearchSource, bool) {
	rawURL := strings.TrimSpace(source.URL)
	// Reject rather than truncate a URL into a different/broken destination.
	if rawURL == "" || len(rawURL) > maxWebSearchEventURLLength {
		return WebSearchSource{}, false
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return WebSearchSource{}, false
	}
	return WebSearchSource{URL: strings.Clone(rawURL), Title: boundedWebSearchString(source.Title, maxWebSearchEventTitleLength)}, true
}

func boundedWebSearchString(value string, maxRunes int) string {
	value = strings.TrimSpace(value)
	count := 0
	for offset := range value {
		if count == maxRunes {
			value = value[:offset]
			break
		}
		count++
	}
	// PostgreSQL text rejects NUL. Other control characters aren't useful in
	// these plain-text audit fields, except tab/newline in multi-query values.
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, value)
	return strings.Clone(value)
}
