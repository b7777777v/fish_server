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

// DBManager 管理讀寫資料庫連接
type DBManager struct {
	writeDB *Client
	readDB  *Client
	logger  logger.Logger
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

// NewDBManager 創建資料庫管理器，支持讀寫分離
// 當前實現：連線字串共用，主從使用相同的資料庫配置
// 已廢棄：建議使用 NewDBManagerWithConfig
func NewDBManager(dbConfig *conf.Database, logger logger.Logger) (*DBManager, error) {
	if dbConfig == nil {
		logger.Error("database config is nil")
		return nil, fmt.Errorf("database config is nil")
	}

	// 創建寫庫連接
	writeDB, err := NewClientFromDatabase(dbConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create write db client: %w", err)
	}

	// 創建讀庫連接（現階段使用相同的配置）
	// 未來可以擴展為獨立的讀庫配置
	readDB, err := NewClientFromDatabase(dbConfig, logger)
	if err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("failed to create read db client: %w", err)
	}

	return &DBManager{
		writeDB: writeDB,
		readDB:  readDB,
		logger:  logger.With("module", "data/postgres/dbmanager"),
	}, nil
}

// NewDBManagerWithConfig 創建資料庫管理器，支持獨立的讀寫庫配置
// writeDBConfig: 寫庫配置（主庫）
// readDBConfig: 讀庫配置（從庫），如果與 writeDBConfig 相同則表示主從共用
func NewDBManagerWithConfig(writeDBConfig, readDBConfig *conf.Database, logger logger.Logger) (*DBManager, error) {
	if writeDBConfig == nil {
		logger.Error("write database config is nil")
		return nil, fmt.Errorf("write database config is nil")
	}

	// 創建寫庫連接
	writeDB, err := NewClientFromDatabase(writeDBConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create write db client: %w", err)
	}

	// 創建讀庫連接
	var readDB *Client
	if readDBConfig != nil && readDBConfig != writeDBConfig {
		// 使用獨立的讀庫配置
		logger.Infof("Creating separate read database connection: %s:%d/%s",
			readDBConfig.Host, readDBConfig.Port, readDBConfig.DBName)
		readDB, err = NewClientFromDatabase(readDBConfig, logger)
		if err != nil {
			writeDB.Close()
			return nil, fmt.Errorf("failed to create read db client: %w", err)
		}
	} else {
		// 使用寫庫配置創建讀庫連接（主從共用）
		logger.Infof("Using write database for read operations (no separate read database configured)")
		readDB, err = NewClientFromDatabase(writeDBConfig, logger)
		if err != nil {
			writeDB.Close()
			return nil, fmt.Errorf("failed to create read db client: %w", err)
		}
	}

	return &DBManager{
		writeDB: writeDB,
		readDB:  readDB,
		logger:  logger.With("module", "data/postgres/dbmanager"),
	}, nil
}

// Write 返回寫庫客戶端（用於 INSERT、UPDATE、DELETE 操作）
func (m *DBManager) Write() *Client {
	return m.writeDB
}

// Read 返回讀庫客戶端（用於 SELECT 查詢）
func (m *DBManager) Read() *Client {
	return m.readDB
}

// Close 關閉所有資料庫連接
func (m *DBManager) Close() error {
	var errs []error

	if err := m.writeDB.Close(); err != nil {
		m.logger.Errorf("failed to close write db: %v", err)
		errs = append(errs, err)
	}

	if err := m.readDB.Close(); err != nil {
		m.logger.Errorf("failed to close read db: %v", err)
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close db connections: %v", errs)
	}

	return nil
}
