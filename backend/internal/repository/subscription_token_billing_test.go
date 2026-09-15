//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// Seam 3: UsageBillingCommand 携带 token 增量并在同一 UPDATE 中累加三档 token usage

func TestIncrementUsageBillingSubscription_UpdatesTokenUsageInSameStatement(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	// 期望 SQL 同时累加 USD 和 Token 三档用量；token 为 int64
	mock.ExpectExec(`UPDATE user_subscriptions`).
		WithArgs(1.5, int64(101), int64(22)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = incrementUsageBillingSubscription(ctx, tx, 101, 1.5, 22)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncrementUsageBillingSubscription_ZeroTokensKeepsTokenColumnsUnchanged(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	mock.ExpectExec(`UPDATE user_subscriptions`).
		WithArgs(0.5, int64(102), int64(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = incrementUsageBillingSubscription(ctx, tx, 102, 0.5, 0)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
