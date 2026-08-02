package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DBURL       string `mapstructure:"DB_URL"`
	DBToken     string `mapstructure:"DB_TOKEN"`
	DBMode      string `mapstructure:"DB_MODE"`        // "local" | "remote" | "sync"
	DBLocalPath string `mapstructure:"DB_LOCAL_PATH"` // SQLite file path for local/sync modes
	NomeEmpresa string `mapstructure:"NOME_EMPRESA"`
	Addr        string `mapstructure:"ADDR"`
}

func GetConfig() (Config, error) {
	var config Config

	viper.SetConfigName("norte.conf")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("parsing config data: %w", err)
	}

	if config.Addr == "" {
		config.Addr = ":4000"
	}
	if config.NomeEmpresa == "" {
		config.NomeEmpresa = "Norte"
	}
	if config.DBMode == "" {
		config.DBMode = "remote"
	}
	if config.DBLocalPath == "" {
		config.DBLocalPath = "local.db"
	}

	return config, nil
}
