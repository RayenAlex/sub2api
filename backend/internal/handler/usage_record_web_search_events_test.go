package handler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSubmitOpenAIUsageRecordTaskWebSearchEventsMandatory(t *testing.T) {
	pool := service.NewUsageRecordWorkerPoolWithOptions(service.UsageRecordWorkerPoolOptions{
		WorkerCount: 1, QueueSize: 1, TaskTimeout: time.Second, OverflowPolicy: "drop", AutoScaleEnabled: false,
	})
	t.Cleanup(pool.Stop)
	release, started := make(chan struct{}), make(chan struct{})
	defer close(release)
	pool.Submit(func(context.Context) { close(started); <-release })
	<-started
	pool.Submit(func(context.Context) {})
	var called atomic.Bool
	h := &OpenAIGatewayHandler{usageRecordWorkerPool: pool}
	h.submitOpenAIUsageRecordTask(context.Background(), &service.OpenAIForwardResult{
		WebSearchEvents: []service.WebSearchEvent{{Sequence: 1, Query: "query"}},
	}, func(context.Context) { called.Store(true) })
	require.True(t, called.Load(), "parent and child metadata must survive the drop overflow policy")
}
