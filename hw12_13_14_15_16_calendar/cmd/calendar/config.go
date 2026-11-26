package main

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	Server  ServerConf  `mapstructure:"server"`
}

type ServerConf struct {
	Type string `mapstructure:"type"`
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type LoggerConf struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

type StorageConf struct {
	Type string `mapstructure:"type"`
	DSN  string `mapstructure:"dsn"`
}

func NewConfig() Config {
	return Config{}
}

func (c *Config) ParseFile(configPath string) error {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	err := viper.Unmarshal(&c)
	if err != nil {
		return err
	}

	if c.Server.Type != "http" && c.Server.Type != "grpc" {
		return fmt.Errorf("unknown server type, %s", c.Server.Type)
	}

	if len(c.Server.Host) == 0 {
		return errors.New("listen host not set")
	}

	if len(c.Server.Port) == 0 {
		return errors.New("listen port not set")
	}

	return nil
}
