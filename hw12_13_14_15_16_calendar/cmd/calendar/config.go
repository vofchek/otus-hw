package main

import (
	"github.com/spf13/viper"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
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

	return nil
}
