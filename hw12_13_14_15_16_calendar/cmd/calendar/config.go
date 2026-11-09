package main

import (
	"errors"

	"github.com/spf13/viper"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
	HTTP    HTTPConf    `mapstructure:"http"`
}

type LoggerConf struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

type StorageConf struct {
	Type string `mapstructure:"type"`
	DSN  string `mapstructure:"dsn"`
}

type HTTPConf struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
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

	if len(c.HTTP.Host) == 0 {
		return errors.New("http host not set")
	}

	if len(c.HTTP.Port) == 0 {
		return errors.New("http port not set")
	}

	return nil
}
