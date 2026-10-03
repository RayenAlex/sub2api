//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUsageWebSearchEventsIntegration(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newUsageLogRepositoryWithSQL(client, tx)
	suffix := uuid.NewString()
	user := mustCreateUser(t, client, &service.User{Email: "web-search-" + suffix + "@example.com"})
	key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-web-search-" + suffix, Name: "k"})
	account := mustCreateAccount(t, client, &service.Account{Name: "web-search-" + suffix})
	create := func(requestID string, events []service.WebSearchEvent) *service.UsageLog {
		log := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: requestID, Model: "gpt-test", InputTokens: 10, OutputTokens: 5, TotalCost: 1, ActualCost: 1, WebSearchEvents: events}
		inserted, err := repo.Create(ctx, log)
		require.NoError(t, err)
		require.True(t, inserted)
		return log
	}
	old := create(uuid.NewString(), []service.WebSearchEvent{{CallID: "old", Query: "off page"}})
	middle := create(uuid.NewString(), nil)
	latest := create(uuid.NewString(), []service.WebSearchEvent{{Sequence: 4, CallID: "first", Query: "original", Status: "completed"}, {Sequence: 8, CallID: "second", Sources: []service.WebSearchSource{{URL: "https://example.com", Title: "safe"}, {URL: "javascript:alert(1)", Title: "unsafe"}}}})
	duplicate := *latest
	duplicate.WebSearchEvents = []service.WebSearchEvent{{CallID: "replacement", Query: "must not persist"}}
	inserted, err := repo.Create(ctx, &duplicate)
	require.NoError(t, err)
	require.False(t, inserted)
	require.Equal(t, latest.ID, duplicate.ID)
	// Best-effort duplicate must preserve the original children too.
	require.NoError(t, repo.CreateBestEffort(ctx, &duplicate))
	logs, page, err := repo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 2}, UsageLogFilters{UserID: user.ID, ExactTotal: true})
	require.NoError(t, err)
	require.NotNil(t, page)
	require.Equal(t, int64(3), page.Total, "pagination counts only parent requests")
	require.Len(t, logs, 2)
	require.Equal(t, latest.ID, logs[0].ID)
	require.Equal(t, middle.ID, logs[1].ID)
	require.Len(t, logs[0].WebSearchEvents, 2)
	require.Nil(t, logs[1].WebSearchEvents)
	require.Equal(t, "original", logs[0].WebSearchEvents[0].Query)
	require.Equal(t, 1, logs[0].WebSearchEvents[0].Sequence)
	require.Equal(t, 2, logs[0].WebSearchEvents[1].Sequence)
	require.Equal(t, 1, logs[0].WebSearchEvents[1].SourceCount)
	require.WithinDuration(t, latest.CreatedAt, logs[0].WebSearchEvents[0].CreatedAt, time.Microsecond)
	require.NotZero(t, logs[0].WebSearchEvents[0].ID)
	var count, tokens int64
	var cost float64
	require.NoError(t, scanSingleRow(ctx, tx, `SELECT COUNT(*), SUM(input_tokens), SUM(total_cost) FROM usage_logs WHERE user_id=$1`, []any{user.ID}, &count, &tokens, &cost))
	require.Equal(t, int64(3), count)
	require.Equal(t, int64(30), tokens)
	require.Equal(t, float64(3), cost)
	require.NoError(t, repo.Delete(ctx, latest.ID))
	require.NoError(t, scanSingleRow(ctx, tx, `SELECT COUNT(*) FROM usage_web_search_events WHERE usage_log_id=$1`, []any{latest.ID}, &count))
	require.Zero(t, count, "parent deletion must cascade")
	require.NoError(t, scanSingleRow(ctx, tx, `SELECT COUNT(*) FROM usage_web_search_events WHERE usage_log_id=$1`, []any{old.ID}, &count))
	require.Equal(t, int64(1), count, "cascade must leave other parents alone")
	_, err = tx.ExecContext(ctx, `INSERT INTO usage_web_search_events (usage_log_id,sequence) VALUES ($1,1)`, latest.ID)
	require.Error(t, err, "foreign key must reject an orphan")
}

func TestUsageWebSearchEventsEntTransactionRollback(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	repo := newUsageLogRepositoryWithSQL(client, tx)
	suffix := uuid.NewString()
	user := mustCreateUser(t, client, &service.User{Email: "web-rollback-" + suffix + "@example.com"})
	key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-web-rollback-" + suffix, Name: "k"})
	account := mustCreateAccount(t, client, &service.Account{Name: "web-rollback-" + suffix})
	// A transaction-local constraint forces an actual PostgreSQL child failure.
	_, err := tx.ExecContext(ctx, `ALTER TABLE usage_web_search_events ADD CONSTRAINT test_web_search_child_failure CHECK (query <> 'force-child-failure')`)
	require.NoError(t, err)
	log := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: suffix, Model: "gpt-test", WebSearchEvents: []service.WebSearchEvent{{Query: "force-child-failure"}}}
	inserted, err := repo.Create(ctx, log)
	require.Error(t, err)
	require.False(t, inserted)
	require.Zero(t, log.ID)
	require.NoError(t, tx.Rollback())
	var count int64
	require.NoError(t, integrationDB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM usage_logs WHERE request_id=$1`, suffix).Scan(&count))
	require.Zero(t, count, "child failure must not leave parent usage behind")
}
