package txmanager

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ConnectionProvider struct {
	manager *Manager
}

func NewConnectionProvider(manager *Manager) *ConnectionProvider {
	return &ConnectionProvider{manager: manager}
}

func (c *ConnectionProvider) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return c.manager.GetConn(ctx).Exec(ctx, sql, args...)
}

func (c *ConnectionProvider) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return c.manager.GetConn(ctx).Query(ctx, sql, args...)
}

func (c *ConnectionProvider) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.manager.GetConn(ctx).QueryRow(ctx, sql, args...)
}
