package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/call-notes-ai-service/internal/constants"
	"github.com/call-notes-ai-service/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrParseConnString    = errors.New("failed to parse connection string")
	ErrCreatePool         = errors.New("failed to create connection pool")
	ErrPingDatabase       = errors.New("failed to ping database")
	ErrConnExhausted      = errors.New("database connection failed after retries")
)

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type Database struct {
	pool *pgxpool.Pool
}

func Initialize(ctx context.Context, cfg *DatabaseConfig) (*Database, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseConnString, err)
	}

	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreatePool, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %v", ErrPingDatabase, err)
	}

	logger.Info(constants.LogMsgDBPoolInitialized,
		constants.LogFieldHost, cfg.Host,
		constants.LogFieldPort, cfg.Port,
		constants.LogFieldDatabase, cfg.Name,
		constants.LogFieldMaxConns, cfg.MaxConnections,
	)

	return &Database{pool: pool}, nil
}

func InitializeWithRetry(ctx context.Context, cfg *DatabaseConfig, maxRetries int, initialBackoff time.Duration) (*Database, error) {
	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info(constants.LogMsgDBConnAttempt, constants.LogFieldAttempt, attempt, constants.LogFieldMaxRetries, maxRetries)

		db, err := Initialize(ctx, cfg)
		if err == nil {
			logger.Info(constants.LogMsgDBConnected, constants.LogFieldAttempt, attempt)
			return db, nil
		}

		lastErr = err
		if attempt < maxRetries {
			logger.Warn(constants.LogMsgDBConnFailed,
				constants.LogFieldAttempt, attempt,
				constants.LogFieldBackoff, backoff.String(),
				constants.LogKeyError, err,
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}

	return nil, fmt.Errorf("%w (%d attempts): %v", ErrConnExhausted, maxRetries, lastErr)
}

func (d *Database) GetPool() *pgxpool.Pool {
	return d.pool
}

func (d *Database) Close() {
	if d.pool != nil {
		d.pool.Close()
		logger.Info(constants.LogMsgDBPoolClosed)
	}
}

func (d *Database) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}
