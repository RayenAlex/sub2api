//go:build unit

package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageWebSearchEventsAtomicCreate(t *testing.T) {
	for _, bestEffort := range []bool{false, true} {
		for _, outcome := range []string{"insert", "duplicate", "child_failure", "commit_failure"} {
			t.Run(outcome+map[bool]string{false: "/Create", true: "/BestEffort"}[bestEffort], func(t *testing.T) {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()
				repo := &usageLogRepository{db: db, sql: db}
				now := time.Now().UTC()
				log := &service.UsageLog{RequestID: "req-web-search", APIKeyID: 2, CreatedAt: now, WebSearchEvents: []service.WebSearchEvent{{Sequence: 99, CallID: "call_1", Query: "search", Status: "completed", SourceCount: 99}, {CallID: "call_2", Query: "second", Status: "completed", Sources: []service.WebSearchSource{{URL: "https://example.com", Title: "Example"}}}}}
				mock.ExpectBegin()
				parent := mock.ExpectQuery("INSERT INTO usage_logs")
				if outcome == "duplicate" {
					parent.WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
					mock.ExpectQuery("SELECT id, created_at FROM usage_logs").WithArgs("req-web-search", int64(2)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(42, now))
				} else {
					parent.WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(42, now))
					child := mock.ExpectExec("INSERT INTO usage_web_search_events").WithArgs(int64(42), 1, "call_1", "search", "completed", 0, "[]", now, int64(42), 2, "call_2", "second", "completed", 1, `[{"url":"https://example.com","title":"Example"}]`, now)
					if outcome == "child_failure" {
						child.WillReturnError(errors.New("child failed"))
					} else {
						child.WillReturnResult(sqlmock.NewResult(0, 2))
					}
				}
				if outcome == "child_failure" {
					mock.ExpectRollback()
				} else if outcome == "commit_failure" {
					mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
				} else {
					mock.ExpectCommit()
				}
				var inserted bool
				if bestEffort {
					err = repo.CreateBestEffort(context.Background(), log)
				} else {
					inserted, err = repo.Create(context.Background(), log)
				}
				if outcome == "child_failure" || outcome == "commit_failure" {
					require.Error(t, err)
					require.Zero(t, log.ID, "failed transaction must not publish speculative parent ID")
					require.Equal(t, 99, log.WebSearchEvents[0].Sequence, "caller metadata must remain retryable")
				} else {
					require.NoError(t, err)
					require.Equal(t, int64(42), log.ID)
					if !bestEffort {
						require.Equal(t, outcome == "insert", inserted)
					}
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}

func TestUsageWebSearchEventsBulkHydration(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := &usageLogRepository{sql: db}
	now := time.Now().UTC()
	query := `SELECT id, usage_log_id, sequence, call_id, query, status, source_count, sources, created_at FROM usage_web_search_events WHERE usage_log_id = ANY($1) ORDER BY usage_log_id, sequence`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("{20,10,30}").WillReturnRows(sqlmock.NewRows([]string{"id", "usage_log_id", "sequence", "call_id", "query", "status", "source_count", "sources", "created_at"}).AddRow(1, 10, 1, "a", "one", "completed", 0, "[]", now).AddRow(2, 10, 2, "b", "two", "completed", 0, "[]", now).AddRow(3, 20, 1, "c", "three", "completed", 1, `[{"url":"https://example.com","title":"Example"}]`, now))
	logs := []service.UsageLog{{ID: 20}, {ID: 10}, {ID: 30}}
	require.NoError(t, repo.hydrateUsageWebSearchEvents(context.Background(), logs))
	require.Len(t, logs[0].WebSearchEvents, 1)
	require.Len(t, logs[1].WebSearchEvents, 2)
	require.Equal(t, []int{1, 2}, []int{logs[1].WebSearchEvents[0].Sequence, logs[1].WebSearchEvents[1].Sequence})
	require.NotNil(t, logs[1].WebSearchEvents[0].Sources)
	require.Nil(t, logs[2].WebSearchEvents)
	require.NoError(t, repo.hydrateUsageWebSearchEvents(context.Background(), nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageWebSearchEventsHydrationErrors(t *testing.T) {
	for _, data := range []string{"invalid", `{}`} {
		t.Run(data, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			mock.ExpectQuery("SELECT id, usage_log_id").WillReturnRows(sqlmock.NewRows([]string{"id", "usage_log_id", "sequence", "call_id", "query", "status", "source_count", "sources", "created_at"}).AddRow(1, 10, 1, "a", "q", "completed", 0, data, time.Now()))
			repo := &usageLogRepository{sql: db}
			require.Error(t, repo.hydrateUsageWebSearchEvents(context.Background(), []service.UsageLog{{ID: 10}}))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
