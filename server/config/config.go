package config

import "github.com/spf13/viper"

type Config struct {
	Port        string `mapstructure:"PORT"`
	JWTSECRET   string `mapstructure:"JWT_SECRET"`
	StoragePath string `mapstructure:"STORAGE_PATH"`
	DB_DSN      string `mapstructure:"DB_DSN"`
}

func LoadConfig() (*Config, error) {

	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
