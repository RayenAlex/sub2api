//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUsageWebSearchEventsSQLTransactionAndConcurrentRetries(t *testing.T) {
	for _, bestEffort := range []bool{false, true} {
		t.Run(fmt.Sprintf("best_effort=%t", bestEffort), func(t *testing.T) {
			ctx := context.Background()
			client := testEntClient(t)
			repo := newUsageLogRepositoryWithSQL(client, integrationDB)
			suffix := uuid.NewString()
			user := mustCreateUser(t, client, &service.User{Email: "search-sql-" + suffix + "@example.com"})
			key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-search-sql-" + suffix, Name: "search-sql"})
			account := mustCreateAccount(t, client, &service.Account{Name: "search-sql-" + suffix})
			newLog := func(requestID string) *service.UsageLog {
				return &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: requestID, Model: "gpt-test", InputTokens: 10, OutputTokens: 5, TotalCost: 1, ActualCost: 0.75}
			}
			ordinary := newLog("ordinary-" + suffix)
			inserted, err := repo.Create(ctx, ordinary)
			require.NoError(t, err)
			require.True(t, inserted)
			search := newLog("search-" + suffix)
			search.WebSearchEvents = []service.WebSearchEvent{{CallID: "ws_1", Query: "original", Status: "completed", Sources: []service.WebSearchSource{{URL: "https://example.com", Title: "Example"}}}, {CallID: "ws_2", Query: "follow-up", Status: "completed"}}
			if bestEffort {
				require.NoError(t, repo.CreateBestEffort(ctx, search))
			} else {
				inserted, err = repo.Create(ctx, search)
				require.NoError(t, err)
				require.True(t, inserted)
			}
			require.NotZero(t, search.ID)

			// Exercise the production database/sql transaction path, including mixed
			// mandatory/best-effort retries. A retry cannot replace original details.
			const retries = 8
			errors := make(chan error, retries)
			var wg sync.WaitGroup
			for i := 0; i < retries; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					duplicate := newLog(search.RequestID)
					duplicate.WebSearchEvents = []service.WebSearchEvent{{Query: "must not overwrite"}}
					var err error
					if i%2 == 0 {
						err = repo.CreateBestEffort(ctx, duplicate)
					} else {
						var inserted bool
						inserted, err = repo.Create(ctx, duplicate)
						if err == nil && inserted {
							err = fmt.Errorf("duplicate retry inserted a parent")
						}
					}
					if err == nil && duplicate.ID != search.ID {
						err = fmt.Errorf("retry parent ID = %d, want %d", duplicate.ID, search.ID)
					}
					errors <- err
				}(i)
			}
			wg.Wait()
			close(errors)
			for err := range errors {
				require.NoError(t, err)
			}

			filters := UsageLogFilters{UserID: user.ID, ExactTotal: true}
			logs, page, err := repo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, filters)
			require.NoError(t, err)
			require.Equal(t, int64(2), page.Total)
			require.Len(t, logs, 2)
			require.Equal(t, search.ID, logs[0].ID)
			require.Len(t, logs[0].WebSearchEvents, 2)
			require.Equal(t, "original", logs[0].WebSearchEvents[0].Query)
			require.Equal(t, "Example", logs[0].WebSearchEvents[0].Sources[0].Title)
			require.Empty(t, logs[1].WebSearchEvents)
			stats, err := repo.GetStatsWithFilters(ctx, filters)
			require.NoError(t, err)
			require.Equal(t, int64(2), stats.TotalRequests)
			require.Equal(t, int64(20), stats.TotalInputTokens)
			require.Equal(t, int64(10), stats.TotalOutputTokens)
			require.Equal(t, int64(30), stats.TotalTokens)
			require.Equal(t, 2.0, stats.TotalCost)
			require.Equal(t, 1.5, stats.TotalActualCost)
			require.NoError(t, repo.Delete(ctx, search.ID))
			var count int
			require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM usage_web_search_events WHERE usage_log_id = $1", search.ID).Scan(&count))
			require.Zero(t, count)
		})
	}
}
