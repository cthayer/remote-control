package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/cthayer/remote_control/internal/logger"
	config "github.com/cthayer/remote_control/pkg/client_config"
)

const (
	DEFAULT_CLI_CONF_CONFIG_FILE = ""
	DEFAULT_CLI_CONF_BATCH_SIZE  = 5
	DEFAULT_CLI_CONF_DELAY       = 0
	DEFAULT_CLI_CONF_VERBOSE     = 0
	DEFAULT_CLI_CONF_RETRY       = 0
)

var cliRootCmd = cobra.Command{
	Use:     "rc [HOST] COMMAND",
	Short:   "Send a COMMAND to a HOST running the remote-control service",
	Long:    "Send a COMMAND to a HOST running the remote-control service\n\n  HOST        the hostname or ip address of the host to run the command on\n              (omit to read the host(s) from STDIN, 1 host per line)\n\n  COMMAND     the command to run on the host",
	Example: "  rc host1.example.com \"uname -a\" -c config.json\n\n  cat hosts.txt | rc \"uname -a\" -c config.json",
	Args:    cobra.RangeArgs(1, 2),
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(args)
	},
}

type cliConfig struct {
	ConfigFile string `json:"configFile"`
	Port       int    `json:"port"`
	KeyDir     string `json:"keyDir"`
	KeyName    string `json:"keyName"`
	LogLevel   string `json:"logLevel"`
	BatchSize  int    `json:"batchSize"`
	Delay      int    `json:"delay"`
	Verbose    int    `json:"verbose"`
	Retry      int    `json:"retry"`
}

var cliConf cliConfig = cliConfig{
	ConfigFile: DEFAULT_CLI_CONF_CONFIG_FILE,
	Port:       config.DEFAULT_PORT,
	KeyDir:     config.DEFAULT_KEY_DIR,
	KeyName:    config.DEFAULT_KEY_NAME,
	LogLevel:   config.DEFAULT_LOG_LEVEL,
	BatchSize:  DEFAULT_CLI_CONF_BATCH_SIZE,
	Verbose:    DEFAULT_CLI_CONF_VERBOSE,
	Delay:      DEFAULT_CLI_CONF_DELAY,
	Retry:      DEFAULT_CLI_CONF_RETRY,
}

func init() {
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.ConfigFile, "config-file", "c", DEFAULT_CLI_CONF_CONFIG_FILE, "path to JSON formatted configuration file")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.Port, "port", "p", config.DEFAULT_PORT, "port to use when connecting to the server(s)")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.KeyDir, "key-dir", "", config.DEFAULT_KEY_DIR, "path to the folder that contains your private key")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.KeyName, "key-name", "", config.DEFAULT_KEY_NAME, "the filename of your private key without the extension (<name>.key)")
	cliRootCmd.PersistentFlags().StringVarP(&cliConf.LogLevel, "log-level", "", config.DEFAULT_LOG_LEVEL, "the loglevel.  can be one of: error, warn, info, debug")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.BatchSize, "batch-size", "b", DEFAULT_CLI_CONF_BATCH_SIZE, "the max number of hosts to send the command to at once while reading")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.Delay, "delay", "d", DEFAULT_CLI_CONF_DELAY, "the time to wait between batches (in ms)")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.Verbose, "verbose", "", DEFAULT_CLI_CONF_VERBOSE, "set to 1 to display raw response information")
	cliRootCmd.PersistentFlags().IntVarP(&cliConf.Retry, "retry", "r", DEFAULT_CLI_CONF_RETRY, "number of times to retry a failed connection before giving up")

	// Default configuration settings
	viper.SetDefault("configFile", DEFAULT_CLI_CONF_CONFIG_FILE)
	viper.SetDefault("port", config.DEFAULT_PORT)
	viper.SetDefault("keyDir", config.DEFAULT_KEY_DIR)
	viper.SetDefault("keyName", config.DEFAULT_KEY_NAME)
	viper.SetDefault("logLevel", config.DEFAULT_LOG_LEVEL)
	viper.SetDefault("batchSize", DEFAULT_CLI_CONF_BATCH_SIZE)
	viper.SetDefault("delay", DEFAULT_CLI_CONF_DELAY)
	viper.SetDefault("verbose", DEFAULT_CLI_CONF_VERBOSE)
	viper.SetDefault("retry", DEFAULT_CLI_CONF_RETRY)

	// Environment Variables
	viper.SetEnvPrefix("RC")
	_ = viper.BindEnv("configFile")
	_ = viper.BindEnv("port")
	_ = viper.BindEnv("keyDir")
	_ = viper.BindEnv("keyName")
	_ = viper.BindEnv("logLevel")
	_ = viper.BindEnv("batchSize")
	_ = viper.BindEnv("delay")
	_ = viper.BindEnv("verbose")
	_ = viper.BindEnv("retry")

	// Flags
	_ = viper.BindPFlag("configFile", cliRootCmd.PersistentFlags().Lookup("config-file"))
	_ = viper.BindPFlag("port", cliRootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("keyDir", cliRootCmd.PersistentFlags().Lookup("key-dir"))
	_ = viper.BindPFlag("keyName", cliRootCmd.PersistentFlags().Lookup("key-name"))
	_ = viper.BindPFlag("logLevel", cliRootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("batchSize", cliRootCmd.PersistentFlags().Lookup("batch-size"))
	_ = viper.BindPFlag("delay", cliRootCmd.PersistentFlags().Lookup("delay"))
	_ = viper.BindPFlag("verbose", cliRootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("retry", cliRootCmd.PersistentFlags().Lookup("retry"))

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

		// update the config struct again after reading config from file
		if err := updateConfig(); err != nil {
			return err
		}
	}

	// initialize the logger
	log, err := logger.InitLogger(cliConf.LogLevel)

	if err == nil && log != nil {
		log.Debug("Configuration File: " + cliConf.ConfigFile)
		log.Debug("", zap.Any("cliConf", cliConf))
	}

	return err
}

func updateConfig() error {
	return viper.Unmarshal(&cliConf)
}
