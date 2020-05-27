package main

import (
	"github.com/cthayer/remote_control/internal/config"
	"github.com/cthayer/remote_control/internal/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
)

func init() {
	// Default configuration settings
	viper.SetDefault("configFile", "")
	viper.SetDefault("port", 4515)
	viper.SetDefault("host", "")
	viper.SetDefault("certDir", "/etc/rc/certs")
	viper.SetDefault("ciphers", "EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4")
	viper.SetDefault("pidFile", "")
	viper.SetDefault("engineOptions.pingTimeout", 5000)
	viper.SetDefault("engineOptions.pingInterval", 1000)
	viper.SetDefault("logLevel", "info")

	// Environment Variables
	viper.SetEnvPrefix("RC")
	_ = viper.BindEnv("configFile")
	_ = viper.BindEnv("certDir")
	_ = viper.BindEnv("logLevel")

	// Flags
	pflag.String("configFile", "", "path to JSON formatted configuration file")
	pflag.String("certDir", "", "path to JSON formatted configuration file")
	pflag.BoolP("version", "v", false, "display version information")
	pflag.BoolP("help", "h", false, "display help information")
	_ = viper.BindPFlags(pflag.CommandLine)

	// Config File
	viper.SetConfigType("json")
}

func initializeConfig() error {
	if configFile := viper.GetString("configFile"); configFile != "" {
		// read config file
		viper.SetConfigFile(configFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		// watch for changes in the config file
		viper.OnConfigChange(func(e fsnotify.Event) {
			_, _ = os.Stdout.WriteString("Config file changed: " + e.Name + "\n")

			reloadConfig()
		})

		viper.WatchConfig()
	}

	// update the config struct
	if err := updateConfig(); err != nil {
		return err
	}

	// initialize the logger
	_, err := logger.InitLogger(config.GetConfig().LogLevel)

	return err
}

func reloadConfig() {
	log := logger.GetLogger()
	defer log.Sync()

	oldLogLevel := config.GetConfig().LogLevel

	if err := updateConfig(); err != nil {
		log.Error("Failed to update config", zap.Error(err))
		return
	}

	if oldLogLevel != config.GetConfig().LogLevel {
		// reconfigure the logger
		log.Info("Changing log level", zap.String("oldLevel", oldLogLevel), zap.String("newLevel", config.GetConfig().LogLevel))

		_, err := logger.InitLogger(config.GetConfig().LogLevel)

		if err != nil {
			log.Error("Failed to change log level", zap.String("oldLevel", oldLogLevel), zap.String("newLevel", config.GetConfig().LogLevel), zap.Error(err))
		} else {
			log.Info("Log level changed\n")
		}
	}

	log.Info("Config updated\n")
}

func updateConfig() error {
	return viper.Unmarshal(config.GetConfig())
}
