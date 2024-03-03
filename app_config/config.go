package app_config

import (
	"github.com/spf13/viper"
	"pearson-vpn-service/logconfig"
)

var Config *viper.Viper

func init() {
	Config = viper.New()
	Config.SetConfigFile("config.yml")
	if err := Config.ReadInConfig(); err != nil {
		logconfig.Log.Fatalf("Failed to read pvs application config file: %v", err)
	} else {
		logconfig.Log.Println("Config object created")
		logconfig.Log.Println(Config.AllKeys())
		logconfig.Log.Println(Config.AllSettings())
	}

}
