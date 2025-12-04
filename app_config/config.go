package app_config

import (
	"pearson-vpn-service/logconfig"

	"github.com/spf13/viper"
)

var Config *viper.Viper

func init() {
	Config = viper.New()
	Config.SetConfigFile("config.yml")
	if err := Config.ReadInConfig(); err != nil {
		logconfig.Log.Fatalf("Failed to read pvs application config file: %v", err)
	}
}
