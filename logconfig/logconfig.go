package logconfig

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

var Log *logrus.Logger
var config *viper.Viper

func init() {

	config = viper.New()
	config.SetConfigFile("config.yml")
	if err := config.ReadInConfig(); err != nil {
		fmt.Errorf("failed to read pvs application config file: %w", err)
		fmt.Errorf("using stdout for logging")
	}

	enabled := config.GetBool("logging.enabled")
	output := config.GetString("logging.output")
	logLevel := config.GetString("logging.level")

	Log = logrus.New()
	if enabled {
		switch logLevel {
		case "Debug":
			Log.SetLevel(logrus.DebugLevel)
		case "Info":
			Log.SetLevel(logrus.InfoLevel)
		case "Warn":
			Log.SetLevel(logrus.WarnLevel)
		case "Error":
			Log.SetLevel(logrus.ErrorLevel)
		case "Fatal":
			Log.SetLevel(logrus.FatalLevel)
		case "Panic":
			Log.SetLevel(logrus.PanicLevel)
		default:
			Log.SetLevel(logrus.InfoLevel)
		}
	} else {
		Log.SetLevel(logrus.FatalLevel)
	}

	if output == "file" {
		Log.SetFormatter(&logrus.TextFormatter{}) // Or TextFormatter if you prefer
		file, err := os.OpenFile("/var/log/pvs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Log.SetOutput(file)
		} else {
			Log.Info("Failed to log to file, using default stderr")
		}
	} else {
		Log.SetFormatter(&logrus.TextFormatter{})
		Log.SetOutput(os.Stdout)
	}
}
