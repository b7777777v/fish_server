// internal/data/postgres/postgres.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config 是 PostgreSQL 的配置
type Config struct {
	Database *conf.Database
}

// Client 是 PostgreSQL 客戶端
type Client struct {
	Pool   *pgxpool.Pool
	Logger logger.Logger
}

// NewClient 創建一個新的 PostgreSQL 客戶端
func NewClient(cfg *Config, logger logger.Logger) (*Client, error) {
	// 驗證配置
	if cfg.Database == nil {
		logger.Error("database config is nil")
		return nil, fmt.Errorf("database config is nil")
	}
	
	return NewClientFromDatabase(cfg.Database, logger)
}

// NewClientFromDatabase 直接從 Database 配置創建客戶端
func NewClientFromDatabase(dbConfig *conf.Database, logger logger.Logger) (*Client, error) {
	// 驗證配置
	if dbConfig == nil {
		logger.Error("database config is nil")
		return nil, fmt.Errorf("database config is nil")
	}

	dsn := dbConfig.GetDSN()
	if dsn == "" {
		logger.Error("database DSN is empty")
		return nil, fmt.Errorf("database DSN is empty")
	}

	// 創建連接池配置
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Errorf("failed to parse postgres config: %v", err)
		return nil, err
	}

	// 配置連接池
	poolConfig.MaxConns = 100
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// 創建連接池
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Errorf("failed to connect to postgres: %v", err)
		return nil, err
	}

	// 測試連接
	if err := pool.Ping(ctx); err != nil {
		logger.Errorf("failed to ping postgres: %v", err)
		return nil, err
	}

	return &Client{
		Pool:   pool,
		Logger: logger.With("module", "data/postgres"),
	}, nil
}

// Close 關閉數據庫連接
func (c *Client) Close() error {
	c.Pool.Close()
	return nil
}

// Exec 執行SQL語句
func (c *Client) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.Pool.Exec(ctx, sql, args...)
}

// Query 查詢多行數據
func (c *Client) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return c.Pool.Query(ctx, sql, args...)
}

// QueryRow 查詢單行數據
func (c *Client) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return c.Pool.QueryRow(ctx, sql, args...)
}

// Begin 開始事務
func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.Pool.Begin(ctx)
}
