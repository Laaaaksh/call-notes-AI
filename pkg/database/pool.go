// Package database provides PostgreSQL connection management.
package database

//go:generate mockgen -source=pool.go -destination=mock/mock_pool.go -package=mock
//go:generate mockgen -destination=mock/mock_row.go -package=mock github.com/jackc/pgx/v5 Row
//go:generate mockgen -destination=mock/mock_tx.go -package=mock github.com/jackc/pgx/v5 Tx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IPool defines the interface for database pool operations used by repositories.
// Enables dependency injection and easier testing via mockgen.
type IPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// PoolWrapper wraps *pgxpool.Pool to implement IPool interface.
type PoolWrapper struct {
	pool *pgxpool.Pool
}

var _ IPool = (*PoolWrapper)(nil)

// NewPoolWrapper creates a new PoolWrapper from a pgxpool.Pool
func NewPoolWrapper(pool *pgxpool.Pool) *PoolWrapper {
	return &PoolWrapper{pool: pool}
}

func (p *PoolWrapper) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, arguments...)
}

func (p *PoolWrapper) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *PoolWrapper) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

func (p *PoolWrapper) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

func (p *PoolWrapper) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.pool.BeginTx(ctx, txOptions)
}

// GetUnderlyingPool returns the underlying pgxpool.Pool for cases where direct access is needed
func (p *PoolWrapper) GetUnderlyingPool() *pgxpool.Pool {
	return p.pool
}
