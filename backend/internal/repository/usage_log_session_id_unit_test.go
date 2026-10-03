//go:build unit

package repository

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newSessionIDUsageLog(sessionID *string) *service.UsageLog {
	return &service.UsageLog{
		UserID:       1,
		APIKeyID:     2,
		AccountID:    3,
		RequestID:    "req-session-id",
		Model:        "claude-3",
		InputTokens:  10,
		OutputTokens: 5,
		TotalCost:    1.0,
		ActualCost:   1.0,
		SessionID:    sessionID,
		CreatedAt:    time.Now().UTC(),
	}
}

// TestPrepareUsageLogInsert_SessionIDArgWiring pins the session_id column to the
// arg slice / arg-type table so the five INSERT column lists stay in sync. session_id
// is followed by cache_diagnostic, then native_compaction_v2; created_at is always last.
func TestPrepareUsageLogInsert_SessionIDArgWiring(t *testing.T) {
	sessionID := "sess-persisted-123"
	prepared := prepareUsageLogInsert(newSessionIDUsageLog(&sessionID))

	require.Len(t, prepared.args, len(usageLogInsertArgTypes),
		"prepared args must match the arg-type table length")

	// created_at is last; native_compaction_v2 is penultimate; cache_diagnostic precedes it; session_id precedes that.
	sessionArg := prepared.args[len(prepared.args)-4]
	ns, ok := sessionArg.(sql.NullString)
	require.True(t, ok, "session_id arg should be a sql.NullString, got %T", sessionArg)
	require.True(t, ns.Valid)
	require.Equal(t, sessionID, ns.String)

	require.Equal(t, "text", usageLogInsertArgTypes[len(usageLogInsertArgTypes)-4],
		"session_id arg type must be text")
	require.Equal(t, "jsonb", usageLogInsertArgTypes[len(usageLogInsertArgTypes)-3],
		"cache_diagnostic arg type must be jsonb")
	require.Equal(t, "boolean", usageLogInsertArgTypes[len(usageLogInsertArgTypes)-2],
		"native_compaction_v2 arg type must be boolean")
}

// TestPrepareUsageLogInsert_SessionIDNullWhenAbsent proves an absent session id is
// persisted as SQL NULL rather than an empty string.
func TestPrepareUsageLogInsert_SessionIDNullWhenAbsent(t *testing.T) {
	prepared := prepareUsageLogInsert(newSessionIDUsageLog(nil))
	sessionArg := prepared.args[len(prepared.args)-4]
	ns, ok := sessionArg.(sql.NullString)
	require.True(t, ok, "session_id arg should be a sql.NullString, got %T", sessionArg)
	require.False(t, ns.Valid, "absent session id must be NULL, not empty string")

	empty := ""
	preparedEmpty := prepareUsageLogInsert(newSessionIDUsageLog(&empty))
	nsEmpty := preparedEmpty.args[len(preparedEmpty.args)-4].(sql.NullString)
	require.False(t, nsEmpty.Valid, "empty session id must also be NULL")
}

func TestPrepareUsageLogInsert_RequestedReasoningEffortArgWiring(t *testing.T) {
	requested := "max"
	forwarded := "xhigh"
	prepared := prepareUsageLogInsert(&service.UsageLog{
		UserID:                   1,
		APIKeyID:                 2,
		AccountID:                3,
		RequestID:                "req-requested-effort",
		Model:                    "gpt-5.4",
		ReasoningEffort:          &forwarded,
		RequestedReasoningEffort: &requested,
		CreatedAt:                time.Now().UTC(),
	})

	require.Len(t, prepared.args, len(usageLogInsertArgTypes))
	// Resolve positions from the emitted INSERT instead of pinning numeric offsets:
	// adding an unrelated column must not shift this contract to service_tier.
	query, _ := buildUsageLogBestEffortInsertQuery([]usageLogInsertPrepared{prepared})
	_, insert, found := strings.Cut(query, "INSERT INTO usage_logs (")
	require.True(t, found)
	columns, _, found := strings.Cut(insert, ")")
	require.True(t, found)
	columnNames := strings.Split(columns, ",")
	require.Len(t, columnNames, len(prepared.args))
	columnIndex := func(name string) int {
		for i, column := range columnNames {
			if strings.TrimSpace(column) == name {
				return i
			}
		}
		t.Fatalf("missing usage log INSERT column %q", name)
		return -1
	}
	forwardedIndex := columnIndex("reasoning_effort")
	requestedIndex := columnIndex("requested_reasoning_effort")
	require.Equal(t, forwardedIndex+1, requestedIndex, "requested_reasoning_effort must follow reasoning_effort")
	require.Equal(t, "text", usageLogInsertArgTypes[forwardedIndex])
	require.Equal(t, "text", usageLogInsertArgTypes[requestedIndex])

	forwardedArg, ok := prepared.args[forwardedIndex].(sql.NullString)
	require.True(t, ok)
	require.True(t, forwardedArg.Valid)
	require.Equal(t, forwarded, forwardedArg.String)

	requestedArg, ok := prepared.args[requestedIndex].(sql.NullString)
	require.True(t, ok)
	require.True(t, requestedArg.Valid)
	require.Equal(t, requested, requestedArg.String)
}

// TestUsageLogInsertQueries_IncludeSessionID guards that every generated INSERT path
// and the SELECT column list reference session_id.
func TestUsageLogInsertQueries_IncludeSessionID(t *testing.T) {
	require.Contains(t, usageLogSelectColumns, "requested_reasoning_effort",
		"SELECT column list must include requested_reasoning_effort")
	require.Contains(t, usageLogSelectColumns, "session_id",
		"SELECT column list must include session_id")

	sessionID := "sess-in-query"
	log := newSessionIDUsageLog(&sessionID)
	prepared := prepareUsageLogInsert(log)
	key := usageLogBatchKey(log.RequestID, log.APIKeyID)

	batchQuery, batchArgs := buildUsageLogBatchInsertQuery([]string{key},
		map[string]usageLogInsertPrepared{key: prepared})
	require.Contains(t, batchQuery, "session_id")
	require.Contains(t, batchQuery, "requested_reasoning_effort")
	// Two column references (INSERT column list + SELECT ... FROM input) plus the CTE def.
	require.GreaterOrEqual(t, strings.Count(batchQuery, "session_id"), 3)
	require.Len(t, batchArgs, len(prepared.args)+1,
		"batch args include the synthetic input_index before usage-log values")

	bestEffortQuery, bestEffortArgs := buildUsageLogBestEffortInsertQuery([]usageLogInsertPrepared{prepared})
	require.Contains(t, bestEffortQuery, "session_id")
	require.Len(t, bestEffortArgs, len(prepared.args))
}
