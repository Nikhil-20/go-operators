package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Watch struct {
		Resources  []string `mapstructure:"resources"`
		Namespaces []string `mapstructure:"namespaces"`
	} `mapstructure:"watch"`

	Notifier struct {
		Type string `mapstructure:"type"`
	} `mapstructure:"notifier"`
}

func LoadConfig(path string) (*Config, error) {

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error in reading the config file")
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println("Error in unmarshaling the config file")
		return nil, err
	}

	return &config, nil

}
