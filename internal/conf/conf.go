// internal/conf/conf.go
package conf

import (
	"github.com/spf13/viper"
)

// Config 是所有配置的集合
type Config struct {
	Server *Server `mapstructure:"server"`
	Data   *Data   `mapstructure:"data"`
	JWT    *JWT    `mapstructure:"jwt"`
	Log    *Log    `mapstructure:"log"`
}

type Server struct {
	Game  *Service `mapstructure:"game"`
	Admin *Service `mapstructure:"admin"`
}

type Service struct {
	Port int `mapstructure:"port"`
}

type Data struct {
	Database *Database `mapstructure:"database"`
	Redis    *Redis    `mapstructure:"redis"`
}

type Database struct {
	Driver string `mapstructure:"driver"`
	Source string `mapstructure:"source"`
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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// NewConfig 創建並加載配置
func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
