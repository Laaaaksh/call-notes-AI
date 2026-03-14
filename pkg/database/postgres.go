package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/call-notes-ai-service/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrParseConnString = errors.New("failed to parse connection string")
	ErrCreatePool      = errors.New("failed to create connection pool")
	ErrPingDatabase    = errors.New("failed to ping database")
	ErrConnExhausted   = errors.New("database connection failed after retries")
)

// DatabaseConfig holds all database connection parameters
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

// IDatabase is the interface for database operations
type IDatabase interface {
	GetPool() *pgxpool.Pool
	GetIPool() IPool
	Close()
	Ping(ctx context.Context) error
}

// Database implements IDatabase
type Database struct {
	pool    *pgxpool.Pool
	wrapper *PoolWrapper
}

var _ IDatabase = (*Database)(nil)

// Initialize creates and configures the database connection pool
func Initialize(ctx context.Context, cfg *DatabaseConfig) (*Database, error) {
	connString := buildConnectionString(cfg)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseConnString, err)
	}

	applyPoolSettings(poolConfig, cfg)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreatePool, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %v", ErrPingDatabase, err)
	}

	logPoolInitialized(cfg)

	return &Database{pool: pool, wrapper: NewPoolWrapper(pool)}, nil
}

func buildConnectionString(cfg *DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)
}

func applyPoolSettings(poolConfig *pgxpool.Config, cfg *DatabaseConfig) {
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
}

func logPoolInitialized(cfg *DatabaseConfig) {
	logger.Info("Database pool initialized",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name,
		"max_connections", cfg.MaxConnections,
	)
}

// InitializeWithRetry attempts to connect with exponential backoff
func InitializeWithRetry(ctx context.Context, cfg *DatabaseConfig, maxRetries int, initialBackoff time.Duration) (*Database, error) {
	if maxRetries <= 1 {
		return Initialize(ctx, cfg)
	}
	return connectWithRetry(ctx, cfg, maxRetries, initialBackoff)
}

func connectWithRetry(ctx context.Context, cfg *DatabaseConfig, maxRetries int, initialBackoff time.Duration) (*Database, error) {
	var lastErr error
	backoff := initialBackoff
	maxBackoff := 30 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logConnectionAttempt(attempt, maxRetries)

		db, err := Initialize(ctx, cfg)
		if err == nil {
			logConnectionSuccess(attempt)
			return db, nil
		}

		lastErr = err
		if attempt < maxRetries {
			logRetryWithBackoff(attempt, maxRetries, backoff, err)
			sleepWithContext(ctx, backoff)
			backoff = calculateNextBackoff(backoff, maxBackoff)
		}
	}

	logConnectionFailed(maxRetries, lastErr)
	return nil, fmt.Errorf("%w (%d attempts): %v", ErrConnExhausted, maxRetries, lastErr)
}

func logConnectionAttempt(attempt, maxRetries int) {
	logger.Info("Database connection attempt", "attempt", attempt, "max_retries", maxRetries)
}

func logConnectionSuccess(attempt int) {
	logger.Info("Database connected", "attempt", attempt)
}

func logRetryWithBackoff(attempt, maxRetries int, backoff time.Duration, err error) {
	logger.Warn("Database connection failed, retrying",
		"attempt", attempt,
		"max_retries", maxRetries,
		"next_backoff", backoff.String(),
		"error", err,
	)
}

func logConnectionFailed(maxRetries int, err error) {
	logger.Error("Database connection failed after all retries",
		"max_retries", maxRetries,
		"error", err,
	)
}

func sleepWithContext(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
		return
	}
}

func calculateNextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}
	return next
}

// GetPool returns the raw connection pool
func (d *Database) GetPool() *pgxpool.Pool {
	return d.pool
}

// GetIPool returns the IPool interface wrapper for dependency injection
func (d *Database) GetIPool() IPool {
	return d.wrapper
}

// Close closes the database connection pool
func (d *Database) Close() {
	if d.pool == nil {
		return
	}
	d.pool.Close()
	logger.Info("Database pool closed")
}

// Ping checks the database connection
func (d *Database) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// GetStats returns the current pool statistics
func (d *Database) GetStats() *pgxpool.Stat {
	if d.pool == nil {
		return nil
	}
	return d.pool.Stat()
}
