package repository

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/stretchr/testify/require"
)

func TestUserSubscriptionIncrementUsageBindsTokenArgument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })

	mock.ExpectExec(`UPDATE user_subscriptions`).
		WithArgs(1.25, int64(7), int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewUserSubscriptionRepository(client)
	err = repo.IncrementUsage(context.Background(), 7, 1.25, 42)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
