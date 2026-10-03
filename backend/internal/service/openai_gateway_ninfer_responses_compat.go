package service

import "github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"

// ninfer_responses_compat is an explicit opt-in for OpenAI API-key accounts
// whose Responses upstream is NInfer. Keep it account-scoped: official OpenAI
// Responses and other compatible providers can return encrypted reasoning.
const ninferResponsesCompatExtraKey = "ninfer_responses_compat"

// The Chat Completions bridge synthesizes these two Responses fields; neither
// was requested by its Chat Completions client. NInfer cannot represent
// encrypted reasoning or a reasoning summary and rejects both at validation.
// Only remove the bridge-added values, not any explicit Responses client input.
func normalizeNInferChatResponsesRequest(account *Account, req *apicompat.ResponsesRequest) {
	if account == nil || !account.IsOpenAIApiKey() || req == nil {
		return
	}
	compat, ok := account.Extra[ninferResponsesCompatExtraKey].(bool)
	if !ok || !compat {
		return
	}
	if len(req.Include) == 1 && req.Include[0] == "reasoning.encrypted_content" {
		req.Include = nil
	}
	if req.Reasoning != nil && req.Reasoning.Summary == "auto" {
		req.Reasoning.Summary = ""
	}
}
