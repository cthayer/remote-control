package main

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/cthayer/remote_control/internal/config"
	"github.com/cthayer/remote_control/internal/logger"
)

const (
	DEFAULT_CLI_CONF_PID_FILE    = ""
	DEFAULT_CLI_CONF_CONFIG_FILE = ""
)

var cliRootCmd = cobra.Command{
	Use:     "remote-control",
	Short:   "Runs the remote-control service which allows remote commands to be executed on the system using the rc-protocol",
	Long:    "Runs the remote-control service\n\nThis allows clients to use the rc-protocol to send\n remote commands to be executed on the system",
	Example: "  remote-control -c /path/to/config.json",
	Args:    cobra.ExactArgs(0),
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

type cliConfig struct {
	ConfigFile  string
	Port        int
	CertDir     string
	Ciphers     string
	LogLevel    string
	Host        string
	PidFile     string
	TlsKeyFile  string
	TlsCertFile string
}

var cliConf cliConfig = cliConfig{
	ConfigFile:  DEFAULT_CLI_CONF_CONFIG_FILE,
	Port:        config.DEFAULT_PORT,
	CertDir:     config.DEFAULT_CERT_DIR,
	Ciphers:     config.DEFAULT_CIPHERS,
	LogLevel:    config.DEFAULT_LOG_LEVEL,
	Host:        config.DEFAULT_HOST,
	PidFile:     DEFAULT_CLI_CONF_PID_FILE,
	TlsKeyFile:  config.DEFAULT_TLS_KEY_FILE,
	TlsCertFile: config.DEFAULT_TLS_CERT_FILE,
}

var onConfigUpdateFuncs []func() error = make([]func() error, 0)

func init() {
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.ConfigFile, "config-file", "c", DEFAULT_CLI_CONF_CONFIG_FILE, "path to JSON formatted configuration file")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.Port, "port", "p", config.DEFAULT_PORT, "port to listen on")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.Port, "host", "H", config.DEFAULT_HOST, "the interface to bind to")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.CertDir, "cert-dir", "d", config.DEFAULT_CERT_DIR, "path to the folder that contains authorized client public keys")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.Ciphers, "ciphers", "", config.DEFAULT_CIPHERS, "the list of ciphers to use for TLS encryption")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.LogLevel, "log-level", "", config.DEFAULT_LOG_LEVEL, "the loglevel.  can be one of: error, warn, info, debug")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.PidFile, "pid-file", "", DEFAULT_CLI_CONF_PID_FILE, "the file to write the pid to (used for initv style services")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.TlsKeyFile, "tls-key-file", "", config.DEFAULT_TLS_KEY_FILE, "the path to the private key to use for TLS")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.TlsCertFile, "tls-cert-file", "", config.DEFAULT_TLS_CERT_FILE, "the path to the certificate to use for TLS")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.Host, "host", "H", config.DEFAULT_HOST, "the host address to bind to")

	// Default configuration settings
	viper.SetDefault("configFile", DEFAULT_CLI_CONF_CONFIG_FILE)
	viper.SetDefault("port", config.DEFAULT_PORT)
	viper.SetDefault("host", config.DEFAULT_HOST)
	viper.SetDefault("certDir", config.DEFAULT_CERT_DIR)
	viper.SetDefault("ciphers", config.DEFAULT_CIPHERS)
	viper.SetDefault("pidFile", DEFAULT_CLI_CONF_PID_FILE)
	viper.SetDefault("tlsKeyFile", config.DEFAULT_TLS_KEY_FILE)
	viper.SetDefault("tlsCertFile", config.DEFAULT_TLS_CERT_FILE)
	viper.SetDefault("logLevel", config.DEFAULT_LOG_LEVEL)

	// Environment Variables
	viper.SetEnvPrefix("RC")
	_ = viper.BindEnv("configFile")
	_ = viper.BindEnv("certDir")
	_ = viper.BindEnv("logLevel")
	_ = viper.BindEnv("host")
	_ = viper.BindEnv("port")
	_ = viper.BindEnv("pidFile")
	_ = viper.BindEnv("tlsKey")
	_ = viper.BindEnv("tlsCertFile")
	_ = viper.BindEnv("ciphersFile")

	// Flags
	_ = viper.BindPFlag("configFile", cliRootCmd.PersistentFlags().Lookup("config-file"))
	_ = viper.BindPFlag("certDir", cliRootCmd.PersistentFlags().Lookup("cert-dir"))
	_ = viper.BindPFlag("logLevel", cliRootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("host", cliRootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("port", cliRootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("pidFile", cliRootCmd.PersistentFlags().Lookup("pid-file"))
	_ = viper.BindPFlag("tlsKeyFile", cliRootCmd.PersistentFlags().Lookup("tls-key-file"))
	_ = viper.BindPFlag("tlsCertFile", cliRootCmd.PersistentFlags().Lookup("tls-cert-file"))
	_ = viper.BindPFlag("ciphers", cliRootCmd.PersistentFlags().Lookup("ciphers"))

	// Config File
	viper.SetConfigType("json")
}

func initializeConfig() error {
	// update the config struct
	if err := updateConfig(); err != nil {
		return err
	}

	if cliConf.ConfigFile != "" {
		// read config file
		viper.SetConfigFile(cliConf.ConfigFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		// update the config struct
		if err := updateConfig(); err != nil {
			return err
		}

		// watch for changes in the config file
		viper.OnConfigChange(func(e fsnotify.Event) {
			_, _ = os.Stdout.WriteString("Config file changed: " + e.Name + "\n")

			reloadConfig()
		})

		viper.WatchConfig()
	}

	// initialize the logger
	log, err := logger.InitLogger(config.GetConfig().LogLevel)

	if err == nil && log != nil {
		log.Debug("Configuration File: " + cliConf.ConfigFile)
		log.Debug("", zap.Any("cliConf", cliConf))
	}

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

	// call all registered configuration reload handlers
	for _, fn := range onConfigUpdateFuncs {
		err := fn()

		if err != nil {
			log.Error("Error occurred calling configuration reload handler", zap.Error(err), zap.Any("function", fn))
		}

		log.Debug("Finished calling registered configuration reload function")
	}

	log.Info("Config updated\n")
}

func updateConfig() error {
	return viper.Unmarshal(&cliConf)
}

func RegisterOnConfigReload(fn func() error) {
	onConfigUpdateFuncs = append(onConfigUpdateFuncs, fn)
}
