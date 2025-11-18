// internal/conf/conf.go
package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config 是所有配置的集合
type Config struct {
    Environment string    `mapstructure:"environment"`
    Server      *Server   `mapstructure:"server"`
    Data        *Data     `mapstructure:"data"`
    JWT         *JWT      `mapstructure:"jwt"`
    Log         *Log      `mapstructure:"log"`
    Debug       *Debug    `mapstructure:"debug"`
    CORS        *CORS     `mapstructure:"cors"`
    RateLimit   *RateLimit `mapstructure:"rate_limit"`
    Security    *Security `mapstructure:"security"`
    Game        *Game     `mapstructure:"game"`
}

type Server struct {
	Game  *Service `mapstructure:"game"`
	Admin *Service `mapstructure:"admin"`
}

type Service struct {
	Port int `mapstructure:"port"`
}

type Data struct {
	MasterDatabase *Database `mapstructure:"master_database"` // 主庫配置（寫操作）
	SlaveDatabase  *Database `mapstructure:"slave_database"`  // 從庫配置（讀操作，可選）
	Redis          *Redis    `mapstructure:"redis"`
}

type Database struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

// GetDSN 根據 Database 結構構建 DSN
func (d *Database) GetDSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

// GetSlaveDatabase 獲取從庫配置，如果未配置則返回主庫配置
func (d *Data) GetSlaveDatabase() *Database {
	if d.SlaveDatabase != nil {
		return d.SlaveDatabase
	}
	return d.MasterDatabase
}

// GetMasterDatabase 獲取主庫配置
func (d *Data) GetMasterDatabase() *Database {
	return d.MasterDatabase
}

// Deprecated: GetReadDatabase 已廢棄，請使用 GetSlaveDatabase
func (d *Data) GetReadDatabase() *Database {
	return d.GetSlaveDatabase()
}

// Deprecated: GetWriteDatabase 已廢棄，請使用 GetMasterDatabase
func (d *Data) GetWriteDatabase() *Database {
	return d.GetMasterDatabase()
}


type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWT struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
	Expire int64  `mapstructure:"expire"`
}
type Log struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	FilePath string `mapstructure:"file_path"` // 日誌檔案路徑，為空則只輸出到控制台
}

// Debug 調試相關配置
type Debug struct {
	EnablePprof    bool   `mapstructure:"enable_pprof"`
	PprofAuth      bool   `mapstructure:"pprof_auth"`
	PprofAuthKey   string `mapstructure:"pprof_auth_key"`
	EnableGinDebug bool   `mapstructure:"enable_gin_debug"`
	EnableSQLDebug bool   `mapstructure:"enable_sql_debug"`
}

// CORS 跨域配置
type CORS struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// RateLimit 限流配置
type RateLimit struct {
	Enable             bool `mapstructure:"enable"`
	RequestsPerMinute  int  `mapstructure:"requests_per_minute"`
}

// Security 安全配置
type Security struct {
    EnableCSRF         bool   `mapstructure:"enable_csrf"`
    EnableSecureHeaders bool   `mapstructure:"enable_secure_headers"`
    MaxRequestSize     string `mapstructure:"max_request_size"`
}

// Game 遊戲相關配置
type Game struct {
    PrebuiltRooms []PrebuiltRoom `mapstructure:"prebuilt_rooms"`
}

// PrebuiltRoom 預建房間配置
type PrebuiltRoom struct {
    Type       string `mapstructure:"type"`
    MaxPlayers int32  `mapstructure:"max_players"`
    Count      int    `mapstructure:"count"`
}

// NewConfig 創建並加載配置
func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	
	// 如果沒有指定配置文件，根據環境自動選擇
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}
	
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	
	// 啟用環境變量替換
	v.AutomaticEnv()
	v.SetEnvPrefix("FISH")
	
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	// 設置默認值
	setDefaultValues(&c)
	
	// 驗證配置
	if err := validateConfig(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

// getDefaultConfigPath 根據環境變量獲取默認配置路徑
func getDefaultConfigPath() string {
	env := GetEnvironment()
	switch env {
	case "dev", "development":
		return "./configs/config.dev.yaml"
	case "staging", "stag":
		return "./configs/config.staging.yaml"
	case "prod", "production":
		return "./configs/config.prod.yaml"
	default:
		return "./configs/config.yaml"
	}
}

// GetEnvironment 獲取當前環境
func GetEnvironment() string {
	env := viper.GetString("ENVIRONMENT")
	if env == "" {
		env = viper.GetString("ENV")
	}
	if env == "" {
		env = "dev" // 默認開發環境
	}
	return env
}

// setDefaultValues 設置默認值
func setDefaultValues(c *Config) {
	if c.Debug == nil {
		c.Debug = &Debug{}
	}
	if c.CORS == nil {
		c.CORS = &CORS{}
	}
	if c.RateLimit == nil {
		c.RateLimit = &RateLimit{}
	}
    if c.Security == nil {
        c.Security = &Security{}
    }
    if c.Game == nil {
        c.Game = &Game{}
    }
	
	// 根據環境設置默認值
	switch c.Environment {
	case "dev", "development":
		setDevDefaults(c)
	case "staging", "stag":
		setStagingDefaults(c)
	case "prod", "production":
		setProdDefaults(c)
	}
}

// setDevDefaults 設置開發環境默認值
func setDevDefaults(c *Config) {
	if c.Debug.EnablePprof == false && c.Environment == "dev" {
		c.Debug.EnablePprof = true
	}
	if c.Debug.PprofAuth == false {
		c.Debug.PprofAuth = false
	}
	if len(c.CORS.AllowOrigins) == 0 {
		c.CORS.AllowOrigins = []string{"*"}
	}
	if c.RateLimit.Enable == false {
		c.RateLimit.Enable = false
	}
}

// setStagingDefaults 設置預發布環境默認值
func setStagingDefaults(c *Config) {
	c.Debug.EnablePprof = false // Staging 強制關閉 pprof
	c.Debug.PprofAuth = true
	c.RateLimit.Enable = true
	if c.RateLimit.RequestsPerMinute == 0 {
		c.RateLimit.RequestsPerMinute = 100
	}
}

// setProdDefaults 設置生產環境默認值
func setProdDefaults(c *Config) {
	c.Debug.EnablePprof = false // 生產環境強制關閉 pprof
	c.Debug.PprofAuth = true
	c.Debug.EnableGinDebug = false
	c.Debug.EnableSQLDebug = false
	c.RateLimit.Enable = true
	if c.RateLimit.RequestsPerMinute == 0 {
		c.RateLimit.RequestsPerMinute = 60
	}
	if c.Security == nil {
		c.Security = &Security{}
	}
	c.Security.EnableCSRF = true
	c.Security.EnableSecureHeaders = true
}

// validateConfig 驗證配置
func validateConfig(c *Config) error {
	if c.Server == nil || c.Server.Admin == nil {
		return fmt.Errorf("server.admin configuration is required")
	}
	if c.Data == nil || c.Data.MasterDatabase == nil {
		return fmt.Errorf("data.master_database configuration is required")
	}
	if c.JWT == nil || c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}

	// 生產環境額外檢查
	if c.Environment == "prod" || c.Environment == "production" {
		if c.JWT.Secret == "your-super-secret-key" {
			return fmt.Errorf("production environment must use a secure JWT secret")
		}
		if c.Debug.EnablePprof {
			return fmt.Errorf("pprof must be disabled in production environment")
		}
	}

	return nil
}
