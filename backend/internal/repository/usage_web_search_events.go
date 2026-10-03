package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

// Metadata-bearing writes bypass the batchers so parent and children are atomic.
func (r *usageLogRepository) createWithWebSearchEvents(ctx context.Context, log *service.UsageLog) (bool, error) {
	snapshot := *log
	snapshot.WebSearchEvents = service.NormalizeWebSearchEvents(log.WebSearchEvents)
	write := func(exec sqlExecutor) (bool, error) {
		inserted, err := r.createSingle(ctx, exec, &snapshot)
		if err != nil || !inserted {
			return inserted, err
		}
		if err := insertUsageWebSearchEvents(ctx, exec, &snapshot); err != nil {
			return false, err
		}
		return true, nil
	}
	if tx := dbent.TxFromContext(ctx); tx != nil {
		inserted, err := write(tx.Client())
		if err == nil {
			*log = snapshot
		}
		return inserted, err
	}
	if r.db == nil {
		return false, service.MarkUsageLogCreateNotPersisted(fmt.Errorf("web search usage requires a transaction"))
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, service.MarkUsageLogCreateNotPersisted(err)
	}
	defer tx.Rollback()
	inserted, err := write(tx)
	if err != nil {
		return false, service.MarkUsageLogCreateNotPersisted(err)
	}
	// A commit error can have an ambiguous outcome; preserve normal retry semantics.
	if err := tx.Commit(); err != nil {
		return false, err
	}
	*log = snapshot
	return inserted, nil
}

func insertUsageWebSearchEvents(ctx context.Context, exec sqlExecutor, log *service.UsageLog) error {
	if len(log.WebSearchEvents) == 0 {
		return nil
	}
	values := make([]string, 0, len(log.WebSearchEvents))
	args := make([]any, 0, len(log.WebSearchEvents)*8)
	for i := range log.WebSearchEvents {
		event := &log.WebSearchEvents[i]
		event.UsageLogID = log.ID
		if event.CreatedAt.IsZero() {
			event.CreatedAt = log.CreatedAt
		}
		if event.Sources == nil {
			event.Sources = []service.WebSearchSource{}
		}
		event.SourceCount = len(event.Sources)
		sources, err := json.Marshal(event.Sources)
		if err != nil {
			return err
		}
		n := len(args)
		values = append(values, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d::jsonb,$%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8))
		args = append(args, log.ID, event.Sequence, event.CallID, event.Query, event.Status, event.SourceCount, string(sources), event.CreatedAt)
	}
	_, err := exec.ExecContext(ctx, `INSERT INTO usage_web_search_events (usage_log_id, sequence, call_id, query, status, source_count, sources, created_at) VALUES `+strings.Join(values, ","), args...)
	return err
}

// Hydrate only the selected page, without changing parent counts or aggregates.
func (r *usageLogRepository) hydrateUsageWebSearchEvents(ctx context.Context, logs []service.UsageLog) (err error) {
	if len(logs) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(logs))
	byID := make(map[int64]*service.UsageLog, len(logs))
	for i := range logs {
		ids = append(ids, logs[i].ID)
		byID[logs[i].ID] = &logs[i]
		logs[i].WebSearchEvents = nil
	}
	rows, err := r.sql.QueryContext(ctx, `SELECT id, usage_log_id, sequence, call_id, query, status, source_count, sources, created_at FROM usage_web_search_events WHERE usage_log_id = ANY($1) ORDER BY usage_log_id, sequence`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}()
	for rows.Next() {
		var event service.WebSearchEvent
		var sources []byte
		if err := rows.Scan(&event.ID, &event.UsageLogID, &event.Sequence, &event.CallID, &event.Query, &event.Status, &event.SourceCount, &sources, &event.CreatedAt); err != nil {
			return err
		}
		if err := json.Unmarshal(sources, &event.Sources); err != nil {
			return fmt.Errorf("decode web search sources: %w", err)
		}
		if event.Sources == nil {
			event.Sources = []service.WebSearchSource{}
		}
		event.SourceCount = len(event.Sources)
		if log := byID[event.UsageLogID]; log != nil {
			log.WebSearchEvents = append(log.WebSearchEvents, event)
		}
	}
	return rows.Err()
}
